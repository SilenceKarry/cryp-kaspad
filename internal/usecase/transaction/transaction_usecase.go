package transaction

import (
	"context"
	"cryp-kaspad/configs"
	"cryp-kaspad/internal/domain"
	"cryp-kaspad/internal/domain/entity"
	"cryp-kaspad/internal/domain/repository"
	"cryp-kaspad/internal/domain/usecase"
	"cryp-kaspad/internal/domain/vo"
	crypNotify "cryp-kaspad/internal/libs/cryp-notify"
	kaspa "cryp-kaspad/internal/libs/kaspa"
	"cryp-kaspad/internal/libs/kaspa/core/types"
	"cryp-kaspad/internal/libs/response"
	"cryp-kaspad/internal/utils"
	"errors"
	"fmt"
	"math/big"
	"runtime/debug"
	"sync"
	"time"

	//"github.com/ethereum/go-ethereum/common"
	//"github.com/ethereum/go-ethereum/crypto"
	"github.com/panjf2000/ants"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"

	"github.com/gofrs/uuid"
	"go.uber.org/dig"
)

var logTransferSigHash string
var logERC20TransferSigHash string

func init() {
	// logTransferSigHash = getLogTransferSigHash().Hex()
	// logERC20TransferSigHash = getLogERC20TransferSigHash().Hex()
}

type TransactionUseCaseCond struct {
	dig.In

	BlockHeightRepo repository.BlockHeightRepository
	TransRepo       repository.TransactionRepository
	TokenRepo       repository.TokensRepository
	WithdrawRepo    repository.WithdrawRepository

	AddressUseCase  usecase.AddressUseCase
	TokensUseCase   usecase.TokensUseCase
	WithdrawUseCase usecase.WithdrawUseCase
	ConfigUseCase   usecase.ConfigUseCase

	DB *gorm.DB `name:"dbM"`
}

type avgTransactionFee struct {
	// wei
	gasPrice decimal.Decimal
	gasLimit decimal.Decimal

	// eth
	fee decimal.Decimal
}

type tokenTransactionData struct {
	block     *types.Block
	tokenName string
	txId      string
}

type transactionUseCase struct {
	TransactionUseCaseCond
}

func NewTransactionUseCase(cond TransactionUseCaseCond) usecase.TransactionUseCase {
	uc := &transactionUseCase{
		TransactionUseCaseCond: cond,
	}

	return uc
}
func (uc *transactionUseCase) GetByTxHash(ctx context.Context, req vo.TransGetTxHashReq) (vo.TransGetTxHashResp, response.Status, error) {
	tx, err := uc.TransRepo.GetByTxHash(ctx, req.TxHash)
	if err != nil && err != gorm.ErrRecordNotFound {
		return vo.TransGetTxHashResp{}, response.CodeInternalError, fmt.Errorf("TransRepo.GetByTxHash error: %s", err)
	}

	if err == gorm.ErrRecordNotFound {
		tx, err := uc.WithdrawUseCase.GetByTxHash(ctx, req.TxHash)
		if err != nil && err != gorm.ErrRecordNotFound {
			return vo.TransGetTxHashResp{}, response.CodeInternalError, fmt.Errorf("TransRepo.GetByTxHash error: %s", err)
		}

		if err == gorm.ErrRecordNotFound {
			return vo.TransGetTxHashResp{}, response.CodeTxNotFound, errors.New(response.CodeTxNotFound.Messages)
		}

		return vo.TransGetTxHashResp{
			BlockHeight: 0,
			TxHash:      tx.TxHash,
			CryptoType:  tx.CryptoType,
			ChainType:   tx.ChainType,
			FromAddress: tx.FromAddress,
			ToAddress:   tx.ToAddress,
			Amount:      tx.Amount,
			Fee:         decimal.Zero,
			FeeCrypto:   tx.CryptoType,
			Status:      domain.TxStatusWaitConfirm,
			Memo:        tx.Memo,
		}, response.Status{}, nil
	}

	if tx.CryptoType != req.CryptoType {
		return vo.TransGetTxHashResp{}, response.CodeCryptoNotFound, errors.New(response.CodeCryptoNotFound.Messages)
	}

	result := vo.TransGetTxHashResp{
		BlockHeight: tx.BlockHeight,
		TxHash:      tx.TxHash,
		CryptoType:  tx.CryptoType,
		ChainType:   tx.ChainType,
		FromAddress: tx.FromAddress,
		ToAddress:   tx.ToAddress,
		Amount:      tx.Amount,
		Fee:         tx.Fee,
		FeeCrypto:   tx.FeeCrypto,
		Status:      tx.Status,
		Memo:        tx.Memo,
	}

	return result, response.Status{}, nil
}

func (uc *transactionUseCase) GetBlockHeight(ctx context.Context) (vo.BlockHeightGetResp, response.Status, error) {
	urls := uc.ConfigUseCase.GetNodeUrl(ctx)

	c, err := kaspa.StartDeamon(ctx, urls)
	if err != nil {
		return vo.BlockHeightGetResp{}, response.CodeInternalError, fmt.Errorf("StartDeamon error: %s", err)
	}

	getBlockCountResponse, err := c.RpcClient.GetVirtualSelectedParentBlueScore()
	if err != nil {
		return vo.BlockHeightGetResp{}, response.CodeInternalError, fmt.Errorf("Error Retriving BlockCount: %s", err)
	}

	println("getBlockCountResponse:", getBlockCountResponse.BlueScore)
	latestHeight := int64(getBlockCountResponse.BlueScore)

	dbBlockHeight, err := uc.BlockHeightRepo.Get(ctx)
	if err != nil && err != gorm.ErrRecordNotFound {
		return vo.BlockHeightGetResp{}, response.CodeInternalError, fmt.Errorf("BlockHeightRepo.Get error: %s", err)
	}

	return vo.BlockHeightGetResp{
		DBBlockHeight:   dbBlockHeight.BlockHeight,
		NodeBlockHeight: latestHeight,
		Diff:            latestHeight - dbBlockHeight.BlockHeight,
	}, response.Status{}, nil
}

func (uc *transactionUseCase) ListenBlock() {
	uc.runListenBlock_New()
}

func (uc *transactionUseCase) runListenBlock() {
	uid, err := uuid.NewV4()
	if err != nil {
		log.WithFields(log.Fields{
			"err": fmt.Errorf("uuid.NewV4 error: %s", err),
		}).Error("runListenBlock")

		return
	}

	ctx := context.WithValue(context.Background(), utils.LogUUID, uid.String())
	ctx, cancelCtx := context.WithTimeout(ctx, 3*utils.Time30S)
	defer cancelCtx()

	urls := uc.ConfigUseCase.GetNodeUrl(ctx)

	client, err := kaspa.NewClient(ctx, urls)
	if err != nil {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"err":         fmt.Errorf("kaspa.NewClient error: %s", err),
		}).Error("runListenBlock")
		return
	}

	blockHeight, latestHeight, isHaveNewBlock, err := uc.checkNewBlock(ctx, client)
	if err != nil {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"err":         err,
		}).Error("checkNewBlock")

		return
	}

	log.WithFields(log.Fields{
		utils.LogUUID:    ctx.Value(utils.LogUUID),
		"blockHeight":    blockHeight,
		"latestHeight":   latestHeight,
		"isHaveNewBlock": isHaveNewBlock,
	}).Info("current blockHeight")

	// 注意: 處理遞歸終止條件
	if !isHaveNewBlock {
		return
	}

	blockHeightChan := uc.getBlockHeight(ctx, client, blockHeight, latestHeight)

	blockData := uc.getBlock(blockHeightChan)

	transData := uc.makeTransaction(blockData)

	uc.syncBlockChan(transData)
}

func (uc *transactionUseCase) checkNewBlock(ctx context.Context, client *kaspa.Eos) (int64, int64, bool, error) {
	latestHeight, err := client.GetBlockNumberLatest(ctx)
	if err != nil {
		return 0, 0, false, fmt.Errorf("eosevm.GetBlockNumberLatest error: %s", err)
	}

	dbBlockHeight, err := uc.BlockHeightRepo.Get(ctx)
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, 0, false, fmt.Errorf("BlockHeightRepo.Get error: %s", err)
	}

	if err == gorm.ErrRecordNotFound {
		dbBlockHeight = entity.BlockHeight{
			BlockHeight: latestHeight,
		}

		_, err := uc.BlockHeightRepo.Create(ctx, dbBlockHeight)
		if err != nil {
			return 0, 0, false, fmt.Errorf("BlockHeightRepo.Create error: %s", err)
		}
	}

	blockHeight := dbBlockHeight.BlockHeight
	isHaveNewBlock := false

	// 注意: 用 <= 判斷原因是，更新邏輯是 blockHeight + 1
	if blockHeight <= latestHeight {
		isHaveNewBlock = true
	}

	return blockHeight, latestHeight, isHaveNewBlock, nil
}

type blockHeightData struct {
	ctx    context.Context
	client *kaspa.Eos

	blockHeight int64
}

func (uc *transactionUseCase) getBlockHeight(ctx context.Context, client *kaspa.Eos, blockHeight, latestHeight int64) <-chan blockHeightData {
	blockHeightChan := make(chan blockHeightData, 1)

	go func() {
		defer func() {
			log.WithFields(log.Fields{}).Debugln("close blockHeightChan")

			close(blockHeightChan)
		}()

		log.WithFields(log.Fields{
			utils.LogUUID:  ctx.Value(utils.LogUUID),
			"blockHeight":  blockHeight,
			"latestHeight": latestHeight,
		}).Debugln("getBlockHeight")

		for i := blockHeight; i <= latestHeight; i++ {
			b := blockHeightData{
				ctx:         ctx,
				client:      client,
				blockHeight: i,
			}
			blockHeightChan <- b
		}
	}()

	return blockHeightChan
}

type blockData struct {
	ctx    context.Context
	client *kaspa.Eos

	blockIsFail bool
	blockHeight int64
	block       *types.Block
}

func (uc *transactionUseCase) getBlock(blockHeightChan <-chan blockHeightData) <-chan blockData {
	blockChan := make(chan blockData, 1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.WithFields(log.Fields{
					"err":   err,
					"Stack": string(debug.Stack()),
				}).Error("recover getBlock")
				return
			}
		}()

		defer func() {
			log.WithFields(log.Fields{}).Debugln("close getBlock")

			close(blockChan)

			// 確保資料都被取出，避免 goroutine 卡住
			for _ = range blockHeightChan {
			}
		}()

		for v := range blockHeightChan {
			blockIsFail := false

			block, err := v.client.GetBlockByNumberCustom(v.ctx, uint32(v.blockHeight))
			if err != nil {
				log.WithFields(log.Fields{
					utils.LogUUID: v.ctx.Value(utils.LogUUID),
					"err":         err,
					"blockHeight": v.blockHeight,
				}).Error("eosevm.GetBlockByNumber")
				return
			}

			b := blockData{
				ctx:    v.ctx,
				client: v.client,

				blockIsFail: blockIsFail,
				blockHeight: v.blockHeight,
				block:       block,
			}
			blockChan <- b
		}
	}()

	return blockChan
}

type transData struct {
	ctx context.Context

	blockHeight int64
	trans       []entity.Transaction
}

func (uc *transactionUseCase) makeTransaction(blockData <-chan blockData) <-chan transData {
	transChan := make(chan transData, 1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.WithFields(log.Fields{
					"err":   err,
					"Stack": string(debug.Stack()),
				}).Error("recover makeTransaction")
				return
			}
		}()

		defer func() {
			log.WithFields(log.Fields{}).Debugln("close makeTransaction")

			close(transChan)

			// 確保資料都被取出，避免 goroutine 卡住
			for _ = range blockData {
			}
		}()

		for v := range blockData {
			if v.blockIsFail {
				t := transData{
					ctx:         v.ctx,
					blockHeight: v.blockHeight,
					trans:       nil,
				}
				transChan <- t

				continue
			}

			func() {
				errChan := make(chan error, 20)
				collectionTrans := make(chan entity.Transaction, 10)

				defer func() {
					log.WithFields(log.Fields{}).Debugln("close makeTransaction by func()")

					// 確保資料都被取出，避免 goroutine 卡住
					for _ = range errChan {
					}

					// 確保資料都被取出，避免 goroutine 卡住
					for _ = range collectionTrans {
					}
				}()

				closeNotify := make(chan struct{}, 1)

				go func() {
					i := 0
					for {
						<-closeNotify
						i++

						// i == 2 代表 makeTransactionByETH & makeTransactionByTokens，已執行結束
						// 多寫 goroutine 已執行完畢，可進行 close 操作
						if i == 1 {
							close(collectionTrans)
							close(errChan)
							close(closeNotify)
							return
						}
					}
				}()

				err := uc.makeTransactionByTokens(v.ctx, v.client, closeNotify, errChan, collectionTrans, v.block)
				if err != nil {
					go func() {
						closeNotify <- struct{}{}
					}()

					log.WithFields(log.Fields{
						utils.LogUUID: v.ctx.Value(utils.LogUUID),
						"err":         err,
						"blockHeight": v.blockHeight,
					}).Error("makeTransactionByTokens")

					return
				}

				result := make([]entity.Transaction, 0, 20)
				for v := range collectionTrans {
					result = append(result, v)
				}

				if len(errChan) > 0 {
					for err := range errChan {
						log.WithFields(log.Fields{
							utils.LogUUID: v.ctx.Value(utils.LogUUID),
							"err":         err,
							"blockHeight": v.blockHeight,
						}).Error("makeTransaction")
					}

					return
				}

				t := transData{
					ctx:         v.ctx,
					blockHeight: v.blockHeight,
					trans:       result,
				}
				transChan <- t
			}()
		}
	}()

	return transChan
}

type transETH struct {
	errChan chan<- error

	ctx             context.Context
	collectionTrans chan<- entity.Transaction

	block     *types.Block
	tx        *types.Transaction
	networkID *big.Int
}

type transTokens struct {
	errChan chan<- error

	ctx             context.Context
	client          *kaspa.Eos
	collectionTrans chan<- entity.Transaction

	block *types.Block
	tx    types.Transaction

	contractAddr2Tokens map[string]entity.Tokens
}

func (uc *transactionUseCase) makeTransactionByTokens(ctx context.Context, client *kaspa.Eos, closeNotify chan<- struct{},
	errChan chan<- error, collectionTrans chan<- entity.Transaction, block *types.Block) error {
	// contractAddress, err := uc.TokensUseCase.GetContractAddress(ctx)
	// if err != nil {
	// 	return fmt.Errorf("TokensUseCase.GetContractAddress error: %s", err)
	// }
	contractAddress := "temp"
	if len(contractAddress) == 0 {
		go func() {
			closeNotify <- struct{}{}
		}()

		return nil
	}

	contractAddr2Tokens, err := uc.TokensUseCase.GetContractAddr2Tokens(ctx)
	if err != nil {
		return fmt.Errorf("TokensUseCase.GetContractAddr2Token error: %s", err)
	}

	transTokensChan := make(chan transTokens, 20)
	closeChan := make(chan struct{}, 1)
	size := 5

	go func() {
		i := 0

		for {
			<-closeChan
			i++

			// 多寫 goroutine 已執行完畢，可通知更上層
			if i == size {
				closeNotify <- struct{}{}
				close(closeChan)
				return
			}
		}
	}()

	for i := 0; i < size; i++ {
		go func() {
			uc.getTransactionByTokens(closeChan, transTokensChan)
		}()
	}

	go func() {
		defer func() {
			log.WithFields(log.Fields{}).Debugln("close makeTransactionByTokens")

			close(transTokensChan)
		}()
		fmt.Printf("tx count: %v\n", len(block.Transactions))
		for _, tx := range block.Transactions {
			t := transTokens{
				errChan:             errChan,
				ctx:                 ctx,
				client:              client,
				collectionTrans:     collectionTrans,
				block:               block,
				tx:                  tx,
				contractAddr2Tokens: contractAddr2Tokens,
			}
			transTokensChan <- t

			// switch vLog.Topics[0].Hex() {
			// case logTransferSigHash, logERC20TransferSigHash:
			// 	t := transTokens{
			// 		errChan:             errChan,
			// 		ctx:                 ctx,
			// 		client:              client,
			// 		collectionTrans:     collectionTrans,
			// 		block:               block,
			// 	}
			// 	transTokensChan <- t
			// }
		}
	}()

	return nil
}

func (uc *transactionUseCase) getTransactionByTokens(closeChan chan<- struct{}, transTokensChan <-chan transTokens) {
	defer func() {
		if err := recover(); err != nil {
			log.WithFields(log.Fields{
				"err":   err,
				"Stack": string(debug.Stack()),
			}).Error("recover getTransactionByTokens")
			return
		}
	}()

	defer func() {
		log.WithFields(log.Fields{}).Debugln("close getTransactionByTokens")

		closeChan <- struct{}{}

		// 確保資料都被取出，避免 goroutine 卡住
		for _ = range transTokensChan {
		}
	}()

	for v := range transTokensChan {
		actions := gjson.Get(v.tx.Trx.Raw, "transaction.actions")
		if !actions.Exists() {
			continue
		}
		txHash := gjson.Get(v.tx.Trx.Raw, "id").String()
		isNewTransaction, err := uc.isNewTransaction(v.ctx, txHash)
		if err != nil {
			// 避免卡住
			select {
			case v.errChan <- fmt.Errorf("isNewTransaction error: %s", err):
			default:
			}

			return
		}

		if !isNewTransaction {
			continue
		}

		for _, action := range actions.Array() {
			account := action.Get("account").String()
			tokens, ok := v.contractAddr2Tokens[account]
			if !ok {
				continue
			}

			name := action.Get("name").String()
			if name != "transfer" {
				continue
			}

			fmt.Printf(action.String(), "\n")
			fmt.Printf("action: %v\n", action.Get("account").String())
			fmt.Printf("	action: %v\n", action.Get("name").String())
			fmt.Printf("	action data: %v\n", action.Get("data").String())

			fromAddr := action.Get("data.from").String()
			toAddr := action.Get("data.to").String()
			quantity := action.Get("data.quantity").String()
			memo := action.Get("data.memo").String()

			addressMap, err := uc.getAddressMap(v.ctx, fromAddr, toAddr)
			if err != nil {
				// 注意: 代表此地址，不是服務產生過的，略過不處理
				continue
			}

			log.WithFields(log.Fields{
				"txHash": txHash,
				"action": action.String(),
			}).Info("action info")

			amount, err := kaspa.AssetConvertToDecimal(quantity)
			if err != nil {
				select {
				case v.errChan <- fmt.Errorf("kaspa.AssetConvertToDecimal error: %s", err):
				default:
				}
			}

			txType, err := uc.getTxType(addressMap, fromAddr, toAddr)
			if err != nil {
				select {
				case v.errChan <- fmt.Errorf("geTxType error: %s", err):
				default:
				}

				return
			}
			//2024-08-07T03:21:01.000
			timestamp, err := time.Parse("2006-01-02T15:04:05.000", v.block.Timestamp)
			if err != nil {
				select {
				case v.errChan <- fmt.Errorf("time.Parse error: %s", err):
				default:
				}

				return
			}

			if err != nil {
				select {
				case v.errChan <- fmt.Errorf("time.Parse error: %s", err):
				default:
				}

				return
			}

			trans := entity.Transaction{
				TxType:           txType,
				TransactionTime:  timestamp.Unix(),
				BlockHeight:      v.block.BlockNum,
				TransactionIndex: 0,
				TxHash:           txHash,
				CryptoType:       tokens.CryptoType,
				ChainType:        tokens.ChainType,
				ContractAddr:     tokens.ContractAddr,
				FromAddress:      fromAddr,
				ToAddress:        toAddr,
				Amount:           amount,
				Gas:              0,
				GasUsed:          0,
				GasPrice:         decimal.Zero,
				CpuUsage:         v.tx.CPUUsageUs,
				NetUsageWords:    v.tx.NetUsageWords,
				Fee:              decimal.Zero,
				FeeCrypto:        domain.CryptoType,
				Confirm:          0,
				Status:           domain.TxStatusWaitConfirm,
				Memo:             memo,
				NotifyStatus:     domain.TxNotifyStatusNotYetProcessed,
			}

			v.collectionTrans <- trans
		}
	}
}

func (uc *transactionUseCase) isNewTransaction(ctx context.Context, txHash string) (bool, error) {
	_, err := uc.TransRepo.GetByTxHash(ctx, txHash)
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("TransRepo.GetByTxHash, txHash: %s, error: %s", txHash, err)
	}

	if err == gorm.ErrRecordNotFound {
		return true, nil
	}

	return false, nil
}

var errTxChainIDNotEqualNetworkID = errors.New("error tx chainID not equal networkID")

func (uc *transactionUseCase) getAddressMap(ctx context.Context, fromAddr, toAddr string) (map[string]struct{}, error) {
	isExistFromAddress, isExistToAddress := false, false
	address2Struct := make(map[string]struct{})

	fromAddress, err := uc.AddressUseCase.GetByAddress(ctx, fromAddr)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("AddressUseCase.GetByAddress by fromAddress error: %s", err)
	}

	if fromAddress.ID > 0 {
		isExistFromAddress = true
		address2Struct[fromAddr] = struct{}{}
	}

	toAddress, err := uc.AddressUseCase.GetByAddress(ctx, toAddr)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("AddressUseCase.GetByAddress by toAddress error: %s", err)
	}

	if toAddress.ID > 0 {
		isExistToAddress = true
		address2Struct[toAddr] = struct{}{}
	}

	if !isExistFromAddress && !isExistToAddress {
		return nil, errors.New("address not found")
	}

	return address2Struct, nil
}

func (uc *transactionUseCase) getTxType(address2Struct map[string]struct{}, fromAddr, toAddr string) (int, error) {
	_, ok := address2Struct[fromAddr]
	if ok {
		return domain.TxTypeWithdraw, nil
	}

	_, ok = address2Struct[toAddr]
	if ok {
		return domain.TxTypeDeposit, nil
	}

	return 0, errors.New("txType not found")
}

// func getLogTransferSigHash() common.Hash {
// 	logTransferSig := []byte("Transfer(address,address,uint256)")
// 	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)

// 	return logTransferSigHash
// }

// func getLogERC20TransferSigHash() common.Hash {
// 	logTransferSig := []byte("ERC20Transfer(address,address,uint256)")
// 	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
// 	return logTransferSigHash
// }

func (uc *transactionUseCase) syncBlockChan(transData <-chan transData) []entity.Transaction {
	defer func() {
		log.WithFields(log.Fields{}).Debugln("close syncBlockChan")

		// 確保資料都被取出，避免 goroutine 卡住
		for _ = range transData {
		}
	}()

	var list []entity.Transaction

	for v := range transData {
		if err := uc.syncBlock(v.ctx, v.blockHeight, v.trans); err != nil {
			log.WithFields(log.Fields{
				utils.LogUUID: v.ctx.Value(utils.LogUUID),
				"err":         err,
				"blockHeight": v.blockHeight,
			}).Error("syncBlock")

			return []entity.Transaction{}
		}

		list = append(list, v.trans...)
	}

	return list
}

func (uc *transactionUseCase) syncTransactionChan(transData <-chan transData) []entity.Transaction {
	defer func() {
		log.WithFields(log.Fields{}).Debugln("close syncTransactionChan")

		// 確保資料都被取出，避免 goroutine 卡住
		for _ = range transData {
		}
	}()

	var list []entity.Transaction

	for v := range transData {
		if err := uc.syncTransaction(v.ctx, v.blockHeight, v.trans); err != nil {
			log.WithFields(log.Fields{
				utils.LogUUID: v.ctx.Value(utils.LogUUID),
				"err":         err,
				"blockHeight": v.blockHeight,
			}).Error("syncTransaction")

			return []entity.Transaction{}
		}

		list = append(list, v.trans...)
	}

	return list
}

func (uc *transactionUseCase) syncBlock(ctx context.Context, blockHeight int64, trans []entity.Transaction) error {
	tx := uc.DB.Begin()
	defer tx.Rollback()

	if err := uc.createTransaction(tx, ctx, blockHeight, trans); err != nil {
		return err
	}

	if err := uc.updateBlockHeight(tx, ctx, blockHeight, true); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (uc *transactionUseCase) syncTransaction(ctx context.Context, blockHeight int64, trans []entity.Transaction) error {
	tx := uc.DB.Begin()
	defer tx.Rollback()

	if err := uc.createTransaction(tx, ctx, blockHeight, trans); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (uc *transactionUseCase) createTransaction(tx *gorm.DB, ctx context.Context, blockHeight int64, trans []entity.Transaction) error {
	if len(trans) == 0 {
		return nil
	}

	transRepo := uc.TransRepo.New(tx)

	for i, t := range trans {

		_, err := transRepo.Create(ctx, t)
		if err != nil {
			log.WithFields(log.Fields{
				utils.LogUUID:     ctx.Value(utils.LogUUID),
				"err":             err,
				"blockHeight":     blockHeight,
				"for range index": i,
				"trans":           fmt.Sprintf("%+v", trans[i]),
			}).Error("TransRepo.Create")

			return fmt.Errorf("TransRepo.Create error: %s", err)
		}
	}

	return nil
}

func (uc *transactionUseCase) updateBlockHeight(tx *gorm.DB, ctx context.Context, blockHeight int64, mode bool) error {
	blockHeightRepo := uc.BlockHeightRepo.New(tx)
	bh, err := blockHeightRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("BlockHeightRepo.Get error: %s", err)
	}

	if bh.BlockHeight >= blockHeight || !mode {
		return nil
	}

	bh.BlockHeight = blockHeight + 1
	err = blockHeightRepo.Update(ctx, bh)
	if err != nil {
		return fmt.Errorf("BlockHeightRepo.Update error: %s", err)
	}

	return nil
}

func (uc *transactionUseCase) TransactionConfirm() {
	uid, err := uuid.NewV4()
	if err != nil {
		log.WithFields(log.Fields{
			"err": fmt.Errorf("uuid.NewV4 error: %s", err),
		}).Error("RunTransactionConfirm")

		return
	}

	ctx := context.WithValue(context.Background(), utils.LogUUID, uid.String())
	ctx, cancelCtx := context.WithTimeout(ctx, 3*utils.Time30S)
	defer cancelCtx()

	urls := uc.ConfigUseCase.GetNodeUrl(ctx)

	client, err := kaspa.NewClient(ctx, urls)
	if err != nil {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"err":         fmt.Errorf("kaspa.NewClient error: %s", err),
		}).Error("runListenBlock")
		return
	}

	trans, err := uc.getTransactionConfirm(ctx, client)
	if err != nil {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"err":         err,
		}).Error("getTransactionConfirm")

		return
	}

	err = uc.runTransactionConfirm(ctx, client, trans)
	if err != nil {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"err":         err,
		}).Error("runTransactionConfirm")

		return
	}
}

func (uc *transactionUseCase) getTransactionConfirm(ctx context.Context, client *kaspa.Eos) ([]entity.Transaction, error) {
	latestHeight, err := client.GetBlockNumberLatest(ctx)
	if err != nil {
		return nil, fmt.Errorf("eosevm.GetBlockNumberLatest error: %s", err)
	}

	trans, err := uc.TransRepo.GetListByStatusAndBlockHeight(ctx, domain.TxStatusWaitConfirm, 20, latestHeight)
	if err != nil {
		return nil, fmt.Errorf("TransRepo.GetListByStatusAndBlockHeight error: %s", err)
	}

	return trans, nil
}

func (uc *transactionUseCase) runTransactionConfirm(ctx context.Context, client *kaspa.Eos, trans []entity.Transaction) error {
	for i := range trans {
		v := trans[i]

		block, err := client.GetBlockByNumberCustom(ctx, uint32(v.BlockHeight))
		if err != nil {
			return fmt.Errorf("kaspa.GetBlockByNumberCustom error: %s", err)
		}

		var t *types.Transaction
		for _, tx := range block.Transactions {
			txHash, err := tx.Trx.ID()
			if err != nil {
				log.WithFields(log.Fields{
					utils.LogUUID: ctx.Value(utils.LogUUID),
					"err":         err,
					"tx":          fmt.Sprintf("%+v", tx),
				}).Error("runTransactionConfirm, tx.Trx get ID error")
				continue
			}

			if txHash != v.TxHash {
				continue
			}
			t = &tx
			break
		}

		if t == nil {
			log.WithFields(log.Fields{
				utils.LogUUID: ctx.Value(utils.LogUUID),
				"txHash":      v.TxHash,
				"blockHeight": v.BlockHeight,
			}).Error("runTransactionConfirm, txHash not found in block")
			continue
		}

		if t.Status == "executed" {
			v.Status = domain.TxStatusSuccess
		} else {
			//只看過executed，若出現其他狀態，視為失敗並記錄
			log.WithFields(log.Fields{
				utils.LogUUID: ctx.Value(utils.LogUUID),
				"txHash":      v.TxHash,
				"blockHeight": v.BlockHeight,
				"status":      t.Status,
			}).Warn("runTransactionConfirm, tx.Status not executed")
			v.Status = domain.TxStatusFail
		}

		// v.Status = domain.TxStatusSuccess
		// if receipt.Status != eosevm.TransactionStatusSuccess {
		// 	v.Status = domain.TxStatusFail
		// }

		v.NotifyStatus = domain.TxNotifyStatusWaitNotify

		// gasUsed := int64(receipt.GasUsed)
		// txFree := eosevm.GetTransactionFree(v.GasPrice, gasUsed)

		v.Confirm = configs.App.GetNodeConfirm()
		// v.TransactionIndex = int(receipt.TransactionIndex)
		// v.GasUsed = gasUsed
		// v.Fee = txFree
		//v.FeeCrypto = domain.CryptoType

		err = uc.TransRepo.Update(ctx, v)
		if err != nil {
			return fmt.Errorf("TransRepo.Update, txHash %s, error: %s", v.TxHash, err)
		}
	}

	return nil
}

func (uc *transactionUseCase) TransactionNotify() {
	uid, err := uuid.NewV4()
	if err != nil {
		log.WithFields(log.Fields{
			"err": fmt.Errorf("uuid.NewV4 error: %s", err),
		}).Error("RunTransactionNotify")

		return
	}

	ctx := context.WithValue(context.Background(), utils.LogUUID, uid.String())
	ctx, cancelCtx := context.WithTimeout(ctx, 3*utils.Time30S)
	defer cancelCtx()

	err = uc.runTransactionNotify(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"err":         err,
		}).Error("runTransactionNotify")

		return
	}
}

func (uc *transactionUseCase) runTransactionNotify(ctx context.Context) error {
	trans, err := uc.TransRepo.GetListByNotifyStatus(ctx, domain.TxNotifyStatusWaitNotify)
	if err != nil {
		return fmt.Errorf("TransRepo.GetListByNotifyStatus error: %s", err)
	}

	for i := range trans {
		v := trans[i]

		host, err := uc.getTransactionNotifyHost(ctx, v)
		if err != nil {
			log.WithFields(log.Fields{
				"txHash":      v.TxHash,
				"fromAddress": v.FromAddress,
				"toAddress":   v.ToAddress,
				"err":         err,
			}).Error("getTransactionNotifyHost")

			continue
		}

		req := crypNotify.CreateTransactionNotifyReq{
			TxType:          v.TxType,
			TransactionTime: v.TransactionTime,
			BlockHeight:     v.BlockHeight,
			TxHash:          v.TxHash,
			CryptoType:      v.CryptoType,
			ChainType:       v.ChainType,
			FromAddress:     v.FromAddress,
			ToAddress:       v.ToAddress,
			Amount:          v.Amount,
			Fee:             v.Fee,
			FeeCrypto:       v.FeeCrypto,
			Status:          v.Status,
			Memo:            v.Memo,
		}
		curl, notifyStatus, err := crypNotify.Transaction.CreateTransactionNotify(ctx, host, req)
		if err != nil {
			log.WithFields(log.Fields{
				utils.LogUUID: ctx.Value(utils.LogUUID),
				"err":         err,
				"curl":        curl,
				"txHash":      v.TxHash,
			}).Error("crypNotify.Transaction.CreateTransactionNotify")

			continue
		}

		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"curl":        curl,
		}).Info("crypNotify.Transaction.CreateTransactionNotify curl")

		v.NotifyStatus = notifyStatus

		err = uc.TransRepo.Update(ctx, v)
		if err != nil {
			return fmt.Errorf("TransRepo.Update, txHash %s, error: %s", v.TxHash, err)
		}
	}

	return nil
}

func (uc *transactionUseCase) RiskControlNotify() {
	uid, err := uuid.NewV4()
	if err != nil {
		log.WithFields(log.Fields{
			"err": fmt.Errorf("uuid.NewV4 error: %s", err),
		}).Error("runRiskControlNotify")

		return
	}

	ctx := context.WithValue(context.Background(), utils.LogUUID, uid.String())
	ctx, cancelCtx := context.WithTimeout(ctx, utils.Time30S*3)
	defer cancelCtx()

	err = uc.runRiskControlNotify(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"err":         err,
		}).Error("runRiskControlNotify")

		return
	}
}

func (uc *transactionUseCase) runRiskControlNotify(ctx context.Context) error {
	trans, err := uc.TransRepo.GetListByRiskControlStatus(ctx, domain.TxStatusWaitConfirm)
	if err != nil {
		return fmt.Errorf("TransRepo.GetListByRiskControlStatus error: %s", err)
	}

	for i := range trans {
		v := trans[i]

		if v.Status == domain.TxStatusSuccess || v.Status == domain.TxStatusFail {
			v.RiskControlStatus = domain.TxNotifyStatusSuccess
		} else {
			host, err := uc.getTransactionNotifyHost(ctx, v)
			if err != nil {
				log.WithFields(log.Fields{
					"txHash":      v.TxHash,
					"fromAddress": v.FromAddress,
					"toAddress":   v.ToAddress,
					"err":         err,
				}).Error("getTransactionNotifyHost")

				continue
			}

			req := crypNotify.CreateTransactionNotifyReq{
				TxType:          v.TxType,
				TransactionTime: v.TransactionTime,
				BlockHeight:     v.BlockHeight,
				TxHash:          v.TxHash,
				CryptoType:      v.CryptoType,
				ChainType:       v.ChainType,
				FromAddress:     v.FromAddress,
				ToAddress:       v.ToAddress,
				Amount:          v.Amount,
				Fee:             v.Fee,
				FeeCrypto:       v.FeeCrypto,
				Status:          v.Status,
				Memo:            v.Memo,
			}
			curl, notifyStatus, err := crypNotify.Transaction.CreateTransactionNotify(ctx, host, req)
			if err != nil {
				log.WithFields(log.Fields{
					utils.LogUUID: ctx.Value(utils.LogUUID),
					"err":         err,
					"curl":        curl,
					"txHash":      v.TxHash,
				}).Error("crypNotify.Transaction.CreateTransactionNotify")

				continue
			}

			log.WithFields(log.Fields{
				utils.LogUUID: ctx.Value(utils.LogUUID),
				"curl":        curl,
			}).Info("cryptoNotify.Transaction.CreateTransactionNotify curl")

			v.RiskControlStatus = notifyStatus
		}

		err = uc.TransRepo.Update(ctx, v)
		if err != nil {
			return fmt.Errorf("TransRepo.Update, txHash %s, error: %s", v.TxHash, err)
		}
	}

	return nil
}

func (uc *transactionUseCase) getTransactionNotifyHost(ctx context.Context, trans entity.Transaction) (string, error) {
	merchantType, err := uc.getMerchantType(ctx, trans)
	if err != nil {
		return "", fmt.Errorf("getMerchantType error: %s", err)
	}

	notifyURL, ok := crypNotify.MerchantType2URL[merchantType]
	if !ok {
		return "", fmt.Errorf("not found notifyURL merchantType: %d", merchantType)
	}

	return notifyURL, nil
}

func (uc *transactionUseCase) getMerchantType(ctx context.Context, trans entity.Transaction) (int, error) {
	if trans.TxType == domain.TxTypeWithdraw {
		addr, err := uc.AddressUseCase.GetByAddress(ctx, trans.FromAddress)
		if err != nil {
			return 0, fmt.Errorf("AddressUseCase.GetByAddress by fromAddress error: %s", err)
		}

		return addr.MerchantType, nil
	}

	addr, err := uc.AddressUseCase.GetByAddress(ctx, trans.ToAddress)
	if err != nil {
		return 0, fmt.Errorf("AddressUseCase.GetByAddress by toAddress error: %s", err)
	}

	return addr.MerchantType, nil
}

func (uc *transactionUseCase) CreateTransactionByBlockNumber(ctx context.Context, req vo.CreateTransactionByBlockNumberReq) ([]entity.Transaction, response.Status, error) {
	urls := uc.ConfigUseCase.GetNodeUrl(ctx)

	client, err := kaspa.NewClient(ctx, urls)
	if err != nil {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"err":         fmt.Errorf("kaspa.NewClient error: %s", err),
		}).Error("runListenBlock")
		return nil, response.CodeInternalError, err
	}

	blockHeightChan := uc.getBlockHeight(ctx, client, req.BlockNumber, req.BlockNumber)

	blockData := uc.getBlock(blockHeightChan)

	transData := uc.makeTransaction(blockData)
	//fmt.Printf("transData:%+v", &transData)
	list := uc.syncTransactionChan(transData)

	return list, response.Status{}, nil
}

func (t *transactionUseCase) updateToken(ctx context.Context, feeMap map[string]avgTransactionFee) {
	for tokenName, value := range feeMap {
		updateModel := entity.Tokens{
			CryptoType:     tokenName,
			GasLimit:       value.gasLimit.IntPart(),
			GasPrice:       value.gasPrice,
			TransactionFee: value.fee,
			UpdateTime:     time.Now().Unix(),
		}

		if err := t.TokenRepo.UpdateToken(ctx, updateModel); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("updateToken")
		}
	}
}

type blockData_New struct {
	ctx         context.Context
	blockHeight int64
	block       *types.Block
	tokens      map[string]entity.Tokens
}

func (uc *transactionUseCase) syncBlockWithTransactionByHeight(ctx context.Context, blockHeight int64, latestHeight int64, updateDBHeight bool) map[int64][]entity.Transaction {
	maxSyncNum := int64(300)
	diff := latestHeight - blockHeight
	if diff > maxSyncNum {
		latestHeight = blockHeight + maxSyncNum
	} else {
		latestHeight = blockHeight + diff
	}
	startTime := time.Now()
	log.WithFields(log.Fields{
		"blockHeight":        blockHeight,
		"latestHeight":       latestHeight,
		"goroutineSizeBlock": configs.App.GetGoroutineSizeBlock(),
		"goroutineSizeTrans": configs.App.GetGoroutineSizeTrans(),
	}).Debug("syncBlockWithTransactionByHeight")

	contractAddr2Tokens, err := uc.TokensUseCase.GetContractAddr2Tokens(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("syncBlockWithTransactionByHeight.TokensUseCase.GetContractAddr2Tokens")
		return nil
	}

	blockChan := make(chan blockData_New, 15)
	transChan := make(chan transData, 1000)
	blockCtx, blockCancel := context.WithCancel(ctx)
	defer blockCancel()

	client, err := kaspa.NewClient(ctx, uc.ConfigUseCase.GetNodeUrl(ctx))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("syncBlockWithTransactionByHeight.kaspa.NewClient")

		return nil
	}

	var blockTransactionwithHeightMap sync.Map
	var wg sync.WaitGroup
	var blockWg sync.WaitGroup
	blockAnts, err := ants.NewPoolWithFunc(configs.App.GetGoroutineSizeBlock(), func(i interface{}) {
		select {
		case <-blockCtx.Done():
			blockWg.Done()
			return
		default:
			v := i.(int64)
			data, err := uc.getBlock_New(blockCtx, v, client)
			if err != nil {
				log.WithFields(log.Fields{
					utils.LogUUID: blockCtx.Value(utils.LogUUID),
					"err":         err,
					"blockHeight": v,
				}).Error("syncBlockWithTransactionByHeight.blockAnts.getBlock")
				blockWg.Done()
				blockCancel()
				return
			}
			data.tokens = contractAddr2Tokens
			log.WithFields(log.Fields{
				utils.LogUUID:       blockCtx.Value(utils.LogUUID),
				"blockHeight":       v,
				"len(data.blockTx)": len(data.block.Transactions),
			}).Debug("syncBlockWithTransactionByHeight.blockAnts.getBlock.Write")

			select {
			case <-blockCtx.Done():
				log.Debug("blockCtx.Done in ants")
				blockWg.Done()
				return
			case blockChan <- data:
				log.Debug("blockChan <- data in ants")
			}
		}
	})

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("blockAnts.ants.NewPoolWithFunc")
		return nil
	}

	transCtx, transCancel := context.WithCancel(ctx)
	defer transCancel()

	defer blockAnts.Release()
	var transWg sync.WaitGroup
	TransAnts, err := ants.NewPoolWithFunc(configs.App.GetGoroutineSizeTrans(), func(i interface{}) {
		select {
		case <-transCtx.Done():
			transWg.Done()
			return
		default:
			bd := i.(blockData_New)
			var enityTransations []entity.Transaction
			if len(bd.block.Transactions) == 0 {
				log.WithFields(log.Fields{"height": bd.blockHeight}).Debug("blocktx empty")
				enityTransations = make([]entity.Transaction, 0)
			} else {
				log.WithFields(log.Fields{"height": bd.blockHeight, "tx len": len(bd.block.Transactions)}).Debug("TransAnts Invoke")
				startTime := time.Now()
				enityTransations, err = uc.buildTransactionsFromChen(bd, client)
				if err != nil {
					log.WithFields(log.Fields{
						"err":       err,
						"blockData": bd,
					}).Error("syncBlockWithTransactionByHeight.TransAnts.buildTransactionsFromChen")
					transWg.Done()
					transCancel()
					blockCancel()
					return
				}
				log.WithFields(log.Fields{
					"height": bd.blockHeight,
					"len":    len(enityTransations),
					"cost":   time.Since(startTime).Seconds(),
				}).Debug("buildTransactionsFromChen done")
			}

			log.WithFields(log.Fields{"height": bd.blockHeight}).Debug("enityTransations builded")
			select {
			case <-transCtx.Done():
				transWg.Done()
				log.Debug("transCtx.Done in ants, blockHeight: ", bd.blockHeight)
				return
			default:
				transChan <- transData{
					ctx:         bd.ctx,
					blockHeight: bd.blockHeight,
					trans:       enityTransations,
				}
			}

			log.Debug("Send transChan done, blockHeight: ", bd.blockHeight)
		}
	})

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("TransAnts.ants.NewPoolWithFunc")
		return nil
	}

	defer TransAnts.Release()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-transChan:
				if !ok {
					log.WithFields(log.Fields{}).Debug("transChan is closed")
					return
				}
				blockTransactionwithHeightMap.Store(data.blockHeight, data.trans)
				log.Debug("blockTransactionwithHeightMap.Store, blockHeight: ", data.blockHeight)
				transWg.Done()
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-blockChan:
				if !ok {
					log.WithFields(log.Fields{}).Debug("blockChan is closed")
					return
				}
				transWg.Add(1)
				data.ctx = transCtx
				if err := TransAnts.Invoke(data); err != nil {
					log.WithFields(log.Fields{
						"height":   data.blockHeight,
						"len(txs)": len(data.block.Transactions),
						"err":      err,
					}).Error("syncBlockWithTransactionByHeight.TransAnts")
				}
				log.Debug("TransAnts.Invoke insert, blockHeight: ", data.blockHeight)
				blockWg.Done()

			}
		}
	}()
	wg.Add(1)
	go func() {
		defer func() {
			log.WithFields(log.Fields{}).Debug("close transChan")
			close(transChan)
			wg.Done()
		}()
		c := make(chan struct{})
		go func() {
			defer close(c)
			defer transCancel()
			<-blockCtx.Done()
			transWg.Wait()
		}()
		select {
		case <-ctx.Done():
			return
		case <-c:
			return
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.WithFields(log.Fields{}).Debug("ticker syncBlockToDB done")
		ticker := time.NewTicker(time.Second)
		top := blockHeight
	finish:
		for {
			select {
			case <-ctx.Done():
				break finish
			case <-blockCtx.Done():
				<-transCtx.Done()

				//前面兩個goroutine都結束後，把全部明細一次同步
				log.Debug("transCtx.Done , start last syncBlockToDB")
				_, ok := blockTransactionwithHeightMap.Load(top)
				if !ok {
					log.Debug("last check load map not found, blockHeight: ", top)
					break finish
				}

				var allLastTrans []entity.Transaction
				for ; top <= latestHeight; top++ {
					value, ok := blockTransactionwithHeightMap.Load(top)
					if !ok {
						log.Debug("last load map not found, blockHeight: ", top)
						top--
						//如果找不到連續的高度，代可能有一個高度失敗了，top-1,並且後面的就直接不要了
						break
					}

					allLastTrans = append(allLastTrans, value.([]entity.Transaction)...)
				}

				err := uc.syncBlockToDB(ctx, top, allLastTrans, updateDBHeight)
				if err != nil {
					log.WithFields(log.Fields{
						utils.LogUUID: ctx.Value(utils.LogUUID),
						"err":         err,
						"blockHeight": top,
						"entity":      allLastTrans,
					}).Error("last syncBlockWithTransactionByHeight.syncBlockToDB")
					break finish
				}

				log.Debug("last syncBlockToDB done, blockHeight: ", top)
				break finish
			case <-ticker.C:
				for top <= latestHeight {
					value, ok := blockTransactionwithHeightMap.Load(top)
					if !ok {
						log.WithFields(log.Fields{"height": top}).Debug("blockTransactionwithHeightMap.Load(top) not found")
						break
					}

					trans := value.([]entity.Transaction)
					err := uc.syncBlockToDB(ctx, top, trans, updateDBHeight)
					if err != nil {
						log.WithFields(log.Fields{
							utils.LogUUID: ctx.Value(utils.LogUUID),
							"err":         err,
							"blockHeight": top,
							"entity":      trans,
						}).Error("syncBlockWithTransactionByHeight.syncBlockToDB")
						blockCancel()
						transCancel()
						break finish
					}

					top++
				}
			}
			if top > latestHeight {
				break finish
			}
		}
	}()
	for i := blockHeight; i <= latestHeight; i++ {
		blockWg.Add(1)
		if err := blockAnts.Invoke(i); err != nil {
			log.WithFields(log.Fields{
				"height": i,
				"err":    err,
			}).Error("syncBlockWithTransactionByHeight.blockAnts")
			blockWg.Done()
		}
		log.Debug("blockAnts.Invoke insert, blockHeight: ", i)
		if i == latestHeight {
			log.Debug("last blockAnts.Invoke insert, blockHeight: ", i)
		}
	}
	wg.Add(1)
	go func() {
		defer func() {
			log.WithFields(log.Fields{}).Debug("close blockChan")
			close(blockChan)
			wg.Done()
		}()
		c := make(chan struct{})
		go func() {
			defer close(c)
			defer blockCancel()
			blockWg.Wait()
		}()
		select {
		case <-ctx.Done():
			return
		case <-c:
			return
		}
	}()
	wg.Wait()
	result := make(map[int64][]entity.Transaction)

	tCtx, tCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer tCancel()
	blockHeightModel, err := uc.BlockHeightRepo.Get(tCtx)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("BlockHeightRepo.Get")
	} else {
		endtime := time.Now()
		log.WithFields(log.Fields{
			"startHeight":  blockHeight,
			"nowHeight":    blockHeightModel.BlockHeight,
			"syncedHeight": blockHeightModel.BlockHeight - blockHeight,
			"cost":         fmt.Sprintf("%v%s", endtime.Sub(startTime).Seconds(), "s"),
		}).Info("Sync block finished")
	}

	blockTransactionwithHeightMap.Range(func(key, value interface{}) bool {
		result[key.(int64)] = value.([]entity.Transaction)
		return true
	})

	return result
}

func (uc *transactionUseCase) buildTransactionsFromChen(bd blockData_New, client *kaspa.Eos) ([]entity.Transaction, error) {
	type data struct {
		Result []entity.Transaction
		err    error
	}
	result := make([]entity.Transaction, 0)
	var transBuildWg sync.WaitGroup
	var wg sync.WaitGroup

	transChan := make(chan data, configs.App.GetGoroutineSizeTrans())

	p, err := ants.NewPoolWithFunc(configs.App.GetGoroutineSizeTrans(), func(i interface{}) {
		defer transBuildWg.Done()

		select {
		case <-bd.ctx.Done():
			log.Debug("bd.ctx.Done in ants, get tx stop")
			return
		default:
		}

		index := i.(int)
		tx := bd.block.Transactions[index]
		txHash, err := tx.Trx.ID()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("tx.Trx.ID parse error")
			return
		}

		log.WithFields(log.Fields{
			"txHash": txHash,
		}).Debug("buildTransactionsFromChen tx parse")

		isNewTransaction, err := uc.isNewTransaction(bd.ctx, txHash)
		if err != nil {
			log.WithFields(log.Fields{
				"txHash": txHash,
				"err":    err,
			}).Error("isNewTransaction")
			return
		}

		if !isNewTransaction {
			log.WithFields(log.Fields{
				"txHash": txHash,
				"isNew":  isNewTransaction,
			}).Debug("tx is exist")
			return
		}
		var entityTxs []entity.Transaction

		actions := gjson.Get(tx.Trx.Raw, "transaction.actions")
		if !actions.Exists() {
			return
		}

		for _, action := range actions.Array() {
			account := action.Get("account").String()
			tokens, ok := bd.tokens[account]
			if !ok {
				continue
			}

			name := action.Get("name").String()
			if name != "transfer" {
				continue
			}

			fmt.Printf(action.String(), "\n")
			fmt.Printf("action: %v\n", action.Get("account").String())
			fmt.Printf("	action: %v\n", action.Get("name").String())
			fmt.Printf("	action data: %v\n", action.Get("data").String())

			fromAddr := action.Get("data.from").String()
			toAddr := action.Get("data.to").String()
			quantity := action.Get("data.quantity").String()
			memo := action.Get("data.memo").String()

			addressMap, err := uc.getAddressMap(bd.ctx, fromAddr, toAddr)
			if err != nil {
				// 注意: 代表此地址，不是服務產生過的，略過不處理
				continue
			}

			log.WithFields(log.Fields{
				"txHash": txHash,
				"action": action.String(),
			}).Info("action info")

			amount, err := kaspa.AssetConvertToDecimal(quantity)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("buildTransactionsFromChen.AssetConvertToDecimal")
				return
			}

			txType, err := uc.getTxType(addressMap, fromAddr, toAddr)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("buildTransactionsFromChen.getTxType")
				return
			}
			//2024-08-07T03:21:01.000
			timestamp, err := time.Parse("2006-01-02T15:04:05.000", bd.block.Timestamp)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("buildTransactionsFromChen.time.parse")
				return
			}

			trans := entity.Transaction{
				TxType:           txType,
				TransactionTime:  timestamp.Unix(),
				BlockHeight:      bd.blockHeight,
				TransactionIndex: 0,
				TxHash:           txHash,
				CryptoType:       tokens.CryptoType,
				ChainType:        tokens.ChainType,
				ContractAddr:     tokens.ContractAddr,
				FromAddress:      fromAddr,
				ToAddress:        toAddr,
				Amount:           amount,
				Gas:              0,
				GasUsed:          0,
				GasPrice:         decimal.Zero,
				CpuUsage:         tx.CPUUsageUs,
				NetUsageWords:    tx.NetUsageWords,
				Fee:              decimal.Zero,
				FeeCrypto:        domain.CryptoType,
				Confirm:          0,
				Status:           domain.TxStatusWaitConfirm,
				Memo:             memo,
				NotifyStatus:     domain.TxNotifyStatusNotYetProcessed,
			}

			entityTxs = append(entityTxs, trans)

		}

		transBuildWg.Add(1)
		transChan <- data{Result: entityTxs}
	})

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("buildTransactionsFromChen new ants")
		return nil, err
	}

	defer p.Release()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-bd.ctx.Done():
				return
			case t, ok := <-transChan:
				if !ok {
					return
				}
				if t.err != nil {
					err = t.err
				} else {
					result = append(result, t.Result...)
				}
				transBuildWg.Done()
			}
		}
	}()

	for i := range bd.block.Transactions {
		transBuildWg.Add(1)
		if err := p.Invoke(i); err != nil {
			log.WithFields(log.Fields{
				"index": i,
				"err":   err,
			}).Error("buildTransactionsFromChen ants")
			transBuildWg.Done()
			return nil, err
		}
	}

	wg.Add(1)
	go func() {
		defer func() {
			log.WithFields(log.Fields{}).Debugln("build close transChan")
			close(transChan)
			wg.Done()
		}()
		transBuildWg.Wait()
	}()
	wg.Wait()
	return result, err
}

func (uc *transactionUseCase) getBlock_New(ctx context.Context, blockHeight int64, client *kaspa.Eos) (blockData_New, error) {
	block, err := client.GetBlockByNumberCustom(ctx, uint32(blockHeight))
	if err != nil {
		return blockData_New{}, fmt.Errorf("client.GetBlockByNumber error: %s", err)
	}
	return blockData_New{ctx: ctx, blockHeight: blockHeight, block: block}, nil
}

func (uc *transactionUseCase) syncBlockToDB(ctx context.Context, blockHeight int64, trans []entity.Transaction, mode bool) error {
	tx := uc.DB.Begin()
	defer tx.Rollback()

	if len(trans) != 0 {
		if err := uc.createTransaction(tx, ctx, blockHeight, trans); err != nil {
			return err
		}
	}

	if err := uc.updateBlockHeight(tx, ctx, blockHeight, mode); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (uc *transactionUseCase) runListenBlock_New() {
	uid, err := uuid.NewV4()
	if err != nil {
		log.WithFields(log.Fields{
			"err": fmt.Errorf("uuid.NewV4 error: %s", err),
		}).Error("RunListenBlock")

		return
	}

	ctx := context.WithValue(context.Background(), utils.LogUUID, uid.String())
	ctx, cancelCtx := context.WithTimeout(ctx, 6*utils.Time30S)
	defer cancelCtx()

	client, err := kaspa.NewClient(ctx, uc.ConfigUseCase.GetNodeUrl(ctx))
	if err != nil {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"err":         fmt.Errorf("kaspa.NewClient error: %s", err),
		}).Error("runListenBlock")
		return
	}

	blockHeight, latestHeight, isHaveNewBlock, err := uc.checkNewBlock(ctx, client)
	if err != nil {
		log.WithFields(log.Fields{
			utils.LogUUID: ctx.Value(utils.LogUUID),
			"err":         err,
		}).Error("checkNewBlock")

		return
	}

	log.WithFields(log.Fields{
		utils.LogUUID:    ctx.Value(utils.LogUUID),
		"blockHeight":    blockHeight,
		"latestHeight":   latestHeight,
		"isHaveNewBlock": isHaveNewBlock,
	}).Debug("runListenBlock")

	if !isHaveNewBlock {
		return
	}

	uc.syncBlockWithTransactionByHeight(ctx, blockHeight, latestHeight, true)
}
