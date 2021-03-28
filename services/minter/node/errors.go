package node

import "fmt"

type BlockNotFound struct {
	resp *GetBlockResponse
}

func NewBlockNotFoundError(resp *GetBlockResponse) *BlockNotFound {
	return &BlockNotFound{resp: resp}
}

func (e *BlockNotFound) Error() string {
	return fmt.Sprintf("%d: %s", e.resp.Error.Code, e.resp.Error.Message)
}
