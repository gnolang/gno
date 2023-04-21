package stdlibs

type CallMsg struct {
	Caller  string   `json:"caller,omitempty"`
	PkgPath string   `json:"pkg_path,omitempty"`
	Fn      string   `json:"fn,omitempty"`
	Args    []string `json:"args,omitempty"`
}

type Request struct {
	Call     *CallMsg
	Callback *CallMsg
}

var MsgQueue chan *Request

// deprecated
// var ResQueue chan string // receive response with timeout, used for synchronous call

func init() {
	MsgQueue = make(chan *Request)
	// ResQueue = make(chan string)
}

// this is called by stdlib function
func Send(call, callback *CallMsg, mc chan<- *Request) {
	req := &Request{
		Call:     call,
		Callback: callback,
	}
	mc <- req
}

// synchronous call, use ResQueue to get result
// func SendCall(call *CallMsg) {
// 	req := &Request{
// 		Call: call,
// 	}
// 	MsgQueue <- req
// }
