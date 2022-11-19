package data

type Record struct {
	Total          int              `json:"total"`
	TokenTransfers []TokenTransfers `json:"token_transfers"`
}

type TokenTransfers struct {
	Quant         string `json:"quant"`
	FromAddress   string `json:"from_address"`
	ToAddress     string `json:"to_address"`
	Block         int    `json:"block"`
	TransactionId string `json:"transaction_id"`
	Confirmed     bool   `json:"confirmed"`
	Time          int64  `json:"block_ts"`
	TokenInfo     `json:"tokenInfo"`
}

type TokenInfo struct {
	TokenAbbr string `json:"tokenAbbr"`
}

type Data struct {
	TokenAbbr   string  `json:"tokenAbbr"`
	AmountInUsd float64 `json:"amountInUsd"`
}

type RecordBalance struct {
	Data []Data `json:"data"`
}
