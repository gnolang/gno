package state

type Event struct {
	Sequence string   `json:"sequence"`
	AppName  string   `json:"app_name"`
	Func     string   `json:"func"`
	Args     []string `json:"args"`
}
