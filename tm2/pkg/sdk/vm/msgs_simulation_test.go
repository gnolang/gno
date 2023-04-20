package vm

import (
	_ "embed"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jaekwon/testify/assert"
	"testing"
)

var simulator *Simulator

func init() {
	var err error
	simulator, err = NewSimulator(true, "../../../../gnovm/stdlibs")
	if err != nil {
		panic(err)
	}
	simuAddPkg()
	println("pkg added")
	go simulator.startServer()
	go simulator.ibc.OnRecvPacket()
}

func simuAddPkg() {
	simulator.addPkgFromPath("../../../../examples/gno.land/r/demo/hello/", "gno.land/r/demo/hello")
	simulator.addPkgFromPath("../../../../examples/gno.land/r/demo/greet/", "gno.land/r/demo/greet")
}

//go:embed simulation_data/msg_call_success.json
var msgCallSuccessBz []byte

// No innerMsgs nested
func TestMsgSuccess(t *testing.T) {
	res, _ := simulator.simuCall([][]*std.MemFile{}, msgCallSuccessBz)
	assert.NoError(t, res.Error)

	// l := int(res.Data[0])
	// v := string(res.Data[1 : l+1])
	// assert.Equal(t, v, `hello from contract master, no innerMsgs replied.`)
	select {}
}
