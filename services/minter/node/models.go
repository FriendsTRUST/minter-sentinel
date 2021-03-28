package node

import "time"

type StatusResponse struct {
	LatestBlockHeight int  `json:"latest_block_height,string"`
	CatchingUp        bool `json:"catching_up"`
}

type GetBlockResponse struct {
	Hash             string        `json:"hash"`
	Height           string        `json:"height"`
	Time             time.Time     `json:"time"`
	TransactionCount string        `json:"transaction_count"`
	Transactions     []interface{} `json:"transactions"`
	BlockReward      string        `json:"block_reward"`
	Size             string        `json:"size"`
	Proposer         string        `json:"proposer"`
	Validators       []struct {
		PublicKey string `json:"public_key"`
		Signed    bool   `json:"signed"`
	} `json:"validators"`
	Evidence struct {
		Evidence []interface{} `json:"evidence"`
	} `json:"evidence"`
	Missed []interface{} `json:"missed"`

	Error *Error `json:"error"`
}

type MissedBlocksResponse struct {
	MissedBlocks      *string `json:"missed_blocks"`
	MissedBlocksCount *int    `json:"missed_blocks_count,string"`

	Error *Error `json:"error"`
}

type GetAddressResponse struct {
	TransactionCount uint64 `json:"transaction_count,string"`

	Error *Error `json:"error"`
}

type SendTransactionRequest struct {
	Tx string `json:"tx"`
}

type SendTransactionResponse struct {
	Hash string `json:"string"`

	Error *Error `json:"error"`
}

type Error struct {
	Code    int                     `json:"code,string"`
	Message string                  `json:"message"`
	Data    *map[string]interface{} `json:"data"`
}
