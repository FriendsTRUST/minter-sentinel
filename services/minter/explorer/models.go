package explorer

import "time"

type ListBlocksResponse struct {
	Data []struct {
		Height           int       `json:"height"`
		Size             int       `json:"size"`
		TransactionCount int       `json:"transaction_count"`
		BlockTime        float64   `json:"block_time"`
		Timestamp        time.Time `json:"timestamp"`
		Reward           string    `json:"reward"`
		Hash             string    `json:"hash"`
		ValidatorsCount  int       `json:"validators_count"`
	} `json:"data"`
	Links struct {
		First string `json:"first"`
		Last  string `json:"last"`
		Prev  string `json:"prev"`
		Next  string `json:"next"`
	} `json:"links"`
	Meta struct {
		CurrentPage int    `json:"current_page"`
		LastPage    int    `json:"last_page"`
		Path        string `json:"path"`
		PerPage     int    `json:"per_page"`
		Total       int    `json:"total"`
	} `json:"meta"`

	Error
}

type GetBlockResponse struct {
	Data struct {
		Height     int       `json:"height"`
		BlockTime  float64   `json:"block_time"`
		Timestamp  time.Time `json:"timestamp"`
		Validators []struct {
			Validator struct {
				PublicKey string `json:"public_key"`
			} `json:"validator"`
			Signed bool `json:"signed"`
		} `json:"validators"`
	} `json:"data"`

	Error *Error `json:"error"`
}

type Error struct {
	Code    *int    `json:"code"`
	Message *string `json:"message"`
}
