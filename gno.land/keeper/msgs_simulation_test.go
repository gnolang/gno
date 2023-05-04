package vmk

import (
	_ "embed"
	"sync"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jaekwon/testify/assert"
)

var wg sync.WaitGroup

func setupSimulator(name string, dir string) *Simulator {
	var err error
	simulator, err := NewSimulator(name, dir, true, "../../gnovm/stdlibs")
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
	simulator := setupSimulator("first", "d1")
	// bootstrap handleMsg routine
	wg := &sync.WaitGroup{}
	go simulator.VMKpr.HandleMsg(wg)

	res, _ := simulator.simuCall([][]*std.MemFile{}, msgCallVMBz)
	wg.Wait()
	assert.NoError(t, res.Error)
}

func TestIBCCallSuccess(t *testing.T) {
	simulator := setupSimulator("second", "d2")
	// bootstrap handleMsg routine
	wg := &sync.WaitGroup{}
	go simulator.VMKpr.HandleMsg(wg)

	go simulator.ibcChannelKeeper.OnRecvPacket()
	go simulator.ibcChannelKeeper.OnAcknowledgementPacket()
	res, _ := simulator.simuCall([][]*std.MemFile{}, msgCallIBCBz)

	wg.Wait()
	assert.NoError(t, res.Error)
}
