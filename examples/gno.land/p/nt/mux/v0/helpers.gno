package mux

func defaultNotFoundHandler(res *ResponseWriter, req *Request) {
	res.Write("404")
}
