package node

import "fmt"

type CandidateNotFound struct {
}

func (e *CandidateNotFound) Error() string {
	return "candidate not found"
}

type BlockNotFound struct {
	resp *GetBlockResponse
}

func NewBlockNotFoundError(resp *GetBlockResponse) *BlockNotFound {
	return &BlockNotFound{resp: resp}
}

func (e *BlockNotFound) Error() string {
	return fmt.Sprintf("%d: %s", e.resp.Error.Code, e.resp.Error.Message)
}
