package portalloop

type TendermintStatus struct {
	Result Result `json:"result"`
}

type Result struct {
	SyncInfo SyncInfo `json:"sync_info"`
}

type SyncInfo struct {
	LatestBlockHeight string `json:"latest_block_height"`
}
