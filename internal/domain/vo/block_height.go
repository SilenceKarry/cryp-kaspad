package vo

type BlockHeightGetResp struct {
	DBBlockHeight   int64 `json:"db_block_height"`
	NodeBlockHeight int64 `json:"node_block_height"`
	Diff            int64 `json:"diff"`
}
