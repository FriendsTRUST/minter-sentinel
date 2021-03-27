package minter

type MissedBlocksResponse struct {
	MissedBlocks      *string `json:"missed_blocks"`
	MissedBlocksCount *int    `json:"missed_blocks_count,string"`

	Error *Error `json:"error"`
}

type GetAddressResponse struct {
	TransactionCount uint64 `json:"transaction_count,string"`

	Error *Error `json:"error"`
}

type Error struct {
	Code    int                     `json:"code,string"`
	Message string                  `json:"message"`
	Data    *map[string]interface{} `json:"data"`
}
