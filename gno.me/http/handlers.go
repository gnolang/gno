package http

import (
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"strings"

	gno "github.com/gnolang/gno/gno.me/gno"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type app struct {
	Code      string `json:"code"`
	IsPackage bool   `json:"is_package"`
}

type remoteApp struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type getAppName struct {
	Name string `json:"name"`
}

type appCall struct {
	Name      string `json:"name"`
	IsPackage bool   `json:"is_package"`
	Func      string `json:"func"`
	Args      string `json:"args"`
}

func createApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	var gnoApp app
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := dec.Decode(&gnoApp); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusBadRequest)
		return
	}

	if err := vm.Create(req.Context(), gnoApp.Code, gnoApp.IsPackage); err != nil {
		fmt.Println(err, "error adding package")
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	resp.WriteHeader(gohttp.StatusCreated)
}

func installRemoteApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	var installRemote remoteApp
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := dec.Decode(&installRemote); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusBadRequest)
		return
	}

	remoteBody, err := json.Marshal(getAppName{Name: installRemote.Name})
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	getAppReq, err := gohttp.NewRequest("POST", installRemote.Address+"/system/get-app", strings.NewReader(string(remoteBody)))
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	getAppReq.Header.Set("Content-Type", "application/json")
	getAppResp, err := gohttp.DefaultClient.Do(getAppReq)
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	if getAppResp.StatusCode != gohttp.StatusOK {
		gohttp.Error(resp, "could not get app", getAppResp.StatusCode)
		return
	}

	var memPackage std.MemPackage
	dec = json.NewDecoder(getAppResp.Body)
	defer getAppResp.Body.Close()

	if err := dec.Decode(&memPackage); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	memPackage.Address = installRemote.Address
	if err := vm.CreateMemPackage(req.Context(), &memPackage); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	eventListenerManager.ListenOnPackage(installRemote.Address, memPackage.Path)

	resp.WriteHeader(gohttp.StatusCreated)
}

func getApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	var gnoAppName getAppName
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := dec.Decode(&gnoAppName); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusBadRequest)
		return
	}

	memPackage := vm.QueryMemPackage(req.Context(), "installer")
	if memPackage == nil {
		resp.WriteHeader(gohttp.StatusNotFound)
		return
	}

	resp.WriteHeader(gohttp.StatusOK)
	result, err := json.Marshal(memPackage)
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	resp.Write(result)
}

func callApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	var call appCall
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	if err := dec.Decode(&call); err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusBadRequest)
		return
	}

	var args []string
	if len(call.Args) > 0 {
		args = strings.Split(call.Args, ",")
	}

	res, _, err := vm.Call(req.Context(), call.Name, call.IsPackage, call.Func, args...)
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

	res, err := vm.Run(req.Context(), gnoApp.Code)
	if err != nil {
		gohttp.Error(resp, err.Error(), gohttp.StatusInternalServerError)
		return
	}

	resp.WriteHeader(gohttp.StatusOK)
	resp.Write([]byte(res))
}

// TODO: this should call qrender so it doesn't change state accidentally,
// especially when interacting with remote apps.
func renderApp(resp gohttp.ResponseWriter, req *gohttp.Request) {
	enableCors(&resp)

	path := strings.Trim(req.URL.Path, "/")
	if path == "favicon.ico" {
		return
	}

	isPackage := strings.HasPrefix(path, gno.PkgPrefix)
	renderPathIdx := strings.LastIndex(path, ":")
	var renderPath string
	if renderPathIdx != -1 {
		if renderPathIdx != len(path)-1 {
			renderPath = path[renderPathIdx+1:]
		}
		path = path[:renderPathIdx]
	}

	var appName string
	parts := strings.Split(path, "/")
	appName = parts[len(parts)-1]
	res, _, err := vm.Call(req.Context(), appName, isPackage, "Render", renderPath)
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
