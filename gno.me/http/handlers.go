package http

import (
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"strings"

	gno "github.com/gnolang/gno/gno.me/gno"
)

type app struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type appCall struct {
	Name string `json:"name"`
	Func string `json:"func"`
	Args string `json:"args"`
}

func installApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	var gnoApp app
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := dec.Decode(&gnoApp); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusBadRequest)
		return
	}

	addPkg := gno.NewMsgAddPackage(gnoApp.Name, gnoApp.Code)
	if err := vm.AddPackage(req.Context(), addPkg); err != nil {
		fmt.Println(err, "error adding package")
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	resp.WriteHeader(gohttp.StatusCreated)
}

func callApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	var gnoAppCall appCall
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := dec.Decode(&gnoAppCall); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusBadRequest)
		return
	}

	var args []string
	if len(gnoAppCall.Args) > 0 {
		args = strings.Split(gnoAppCall.Args, ",")
	}

	msgCall := gno.NewMsgCall(gnoAppCall.Name, gnoAppCall.Func, args)
	res, err := vm.Call(req.Context(), msgCall)
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	resp.WriteHeader(gohttp.StatusOK)
	resp.Write([]byte(res))
}

func run(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	var gnoApp app
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := dec.Decode(&gnoApp); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusBadRequest)
		return
	}

	msgRun := gno.MsgRun{Package: gno.NewMemPkg(gnoApp.Name, gnoApp.Code)}
	res, err := vm.Run(req.Context(), msgRun)
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	resp.WriteHeader(gohttp.StatusOK)
	resp.Write([]byte(res))
}

func renderApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	path := strings.Trim(req.URL.Path, "/")
	if path == "favicon.ico" {
		return
	}

	msgCall := gno.NewMsgCall(path, "Render", []string{""})
	res, err := vm.Call(req.Context(), msgCall)
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	// Strip out gno type information and characters that are not nice when
	// rendered in a browser.
	res = res[2 : strings.LastIndex(res, "string")-2]
	res = strings.ReplaceAll(res, "\\n", "")
	res = strings.ReplaceAll(res, "\\t", "")
	res = strings.ReplaceAll(res, "\\\"", "\"")

	resp.Write([]byte(res))
}

func enableCors(w *gohttp.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
}
