package response

type Status struct {
	Code     int    `json:"code"`
	Messages string `json:"messages"`
}

func (s *Status) WithMsg(msg string) Status {
	return Status{
		Code:     s.Code,
		Messages: s.Messages + msg,
	}
}

func new(code int, messages string) Status {
	return Status{
		Code:     code,
		Messages: messages,
	}
}

var (
	codeOk = new(200, "success")

	CodeInternalError              = new(10400, "internal error")
	CodeParamInvalid               = new(10401, "param invalid")
	CodeBalanceInsufficient        = new(10402, "balance insufficient")
	CodeTxNotFound                 = new(10403, "tx not found")
	CodeAddressInvalidLength       = new(10404, "address invalid length")
	CodeCryptoNotFound             = new(10405, "crypto not found")
	CodeTransactionFeeInsufficient = new(10406, "transaction fee insufficient")
	CodeWithdrawFailed             = new(10407, "withdraw failed")
	CodeTransactionOnChain         = new(10408, "transaction on chain")
)
