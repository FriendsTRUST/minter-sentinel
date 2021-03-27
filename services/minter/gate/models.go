package gate

type SendTransactionRequest struct {
	Tx string `json:"tx"`
}

type SendTransactionResponse struct {
	Hash string `json:"string"`

	Error *Error `json:"error"`
}

type Error struct {
	Code  uint64 `json:"code"`
	Log   string `json:"log"`
	Value int    `json:"value"`
	Coin  string `json:"coin"`
}
