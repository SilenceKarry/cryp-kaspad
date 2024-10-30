package vo

import (
	"github.com/shopspring/decimal"
)

type ContractCreateRequest struct {
	ContractAddr string          `json:"contractAddr" binding:"required"`
	Decimals     int             `json:"decimals" binding:"required"`
	GasLimit     int64           `json:"gasLimit" binding:"required"`
	GasPrice     decimal.Decimal `json:"gasPrice" binding:"required"`
	ContractAbi  any             `json:"contractAbi"`
	CryptoType   string          `json:"cryptoType" binding:"required"`
}

type ContractUpdateRequest struct {
	Decimals     *int             `json:"decimals"`
	GasLimit     *int64           `json:"gasLimit"`
	GasPrice     *decimal.Decimal `json:"gasPrice"`
	ContractAbi  any              `json:"contractAbi"`
	CryptoType   string           `json:"cryptoType"`
	ContractAddr string           `json:"contractAddr"`
}

type GetContractResponse struct {
	ContractAddr string          `json:"contract_addr"`
	Decimals     int             `json:"decimals"`
	GasLimit     int64           `json:"gas_limit"`
	GasPrice     decimal.Decimal `json:"gas_price" `
	ContractAbi  any             `json:"contract_abi"`
	CryptoType   string          `json:"crypto_type"`
	ChainType    string          `json:"chain_type"`
}
