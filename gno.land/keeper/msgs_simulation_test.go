package vmk

import (
	_ "embed"
	"sync"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jaekwon/testify/assert"
)

var wg sync.WaitGroup

func setupSimulator(name string) *Simulator {
	var err error
	simulator, err := NewSimulator(name, true, "../../gnovm/stdlibs")
	if err != nil {
		panic(err)
	}
	simulator.simuAddPkg()
	return simulator
}

func (s *Simulator) simuAddPkg() {
	s.addPkgFromPath("../../examples/gno.land/r/demo/x/calls/await/hello_ibc/", "gno.land/r/demo/x/calls/await/hello_ibc")
	s.addPkgFromPath("../../examples/gno.land/r/demo/x/calls/await/hello_vm/", "gno.land/r/demo/x/calls/await/hello_vm")
	s.addPkgFromPath("../../examples/gno.land/r/demo/x/calls/await/greet/", "gno.land/r/demo/x/calls/await/greet")
	s.addPkgFromPath("../../examples/gno.land/r/demo/x/calls/await/hola/", "gno.land/r/demo/x/calls/await/hola")
}

//go:embed simulation_data/msg_call_ibc.json
var msgCallIBCBz []byte

//go:embed simulation_data/msg_call_vm.json
var msgCallVMBz []byte

func TestInternalCallSuccess(t *testing.T) {
	simulator := setupSimulator("first")

	res, _ := simulator.simuCall([][]*std.MemFile{}, msgCallVMBz)
	wg.Wait()
	t.Log("res is: ", string(res.Data))
	assert.NoError(t, res.Error)
	assert.Equal(t, string(res.Data), `("hello(\"greet(\\\"hola\\\" string)\" string)" string)`)
}

func TestIBCCallSuccess(t *testing.T) {
	simulator := setupSimulator("second")

	go simulator.ibcChannelKeeper.OnRecvPacket()
	go simulator.ibcChannelKeeper.OnAcknowledgementPacket()
	res, _ := simulator.simuCall([][]*std.MemFile{}, msgCallIBCBz)

	wg.Wait()
	assert.NoError(t, res.Error)
	assert.Equal(t, string(res.Data), `("hello(\"greet(\\\"hola\\\" string)\" string)" string)`)
}
