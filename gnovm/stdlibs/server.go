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
var ResQueue chan string

func init() {
	MsgQueue = make(chan *Request)
	ResQueue = make(chan string)
}

// this is called by stdlib function
func SendData(call, callback *CallMsg) {
	req := &Request{
		Call:     call,
		Callback: callback,
	}
	MsgQueue <- req
}

func SendCall(call *CallMsg) {
	req := &Request{
		Call: call,
	}
	MsgQueue <- req
}
