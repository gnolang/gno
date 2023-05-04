package vmk

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"

	"os"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	vmi "github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	store "github.com/gnolang/gno/tm2/pkg/store/types"
)

var logger log.Logger

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/vm")
}

type simuOptions struct {
	Verbose    bool          `flag:"verbose" help:"verbose"`
	RootDir    string        `flag:"root-dir" help:"clone location of github.com/gnolang/gno (gnodev tries to guess it)"`
	Run        string        `flag:"run" help:"test name filtering pattern"`
	Timeout    time.Duration `flag:"timeout" help:"max execution time (in ns)"`               // FIXME: support ParseDuration: "1s"
	Precompile bool          `flag:"precompile" help:"precompiling gno to go before testing"` // TODO: precompile should be the default, but it needs to automatically precompile dependencies in memory.
	// VM Options
	// A flag about if we should download the production realms
	// UseNativeLibs bool // experimental, but could be useful for advanced developer needs
}

var defaultSimuOptions = simuOptions{
	Verbose:    false,
	Run:        "",
	RootDir:    "",
	Timeout:    0,
	Precompile: false,
}

var baseKey = store.NewStoreKey("base") // in all test apps
var ivalKey = store.NewStoreKey("ival") // in all test apps
var mainKey = store.NewStoreKey("main") // in all test apps

func testCtx(ms store.MultiStore) sdk.Context {
	header := &bft.Header{ChainID: "dev", Height: 1}
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, header, nil)
	return ctx
}

var flags struct {
	skipFailingGenesisTxs bool
	skipStart             bool
	genesisBalancesFile   string
	genesisTxsFile        string
	chainID               string
	genesisRemote         string
}

type Simulator struct {
	mockApp          *sdk.MockApp
	baseApp          *sdk.BaseApp
	VMKpr            vmi.VMKeeperI
	AccK             auth.AccountKeeper
	BanK             bank.BankKeeper
	ibcChannelKeeper *IBCChannelKeeper
	Ctx              sdk.Context
}

type GnoAccount struct {
	std.BaseAccount
}

func ProtoGnoAccount() std.Account {
	return &GnoAccount{}
}

func (s *Simulator) InjectMsgs(msgs []sdk.Msg, mode sdk.RunTxMode) (result sdk.Result) {
	msgLogs := make([]string, 0, len(msgs))

	data := make([]byte, 0, len(msgs))
	err := error(nil)
	events := []sdk.Event{}

	// NOTE: GasWanted is determined by ante handler and GasUsed by the GasMeter.
	for i, msg := range msgs {
		// match message route
		msgRoute := msg.Route()

		handler := s.mockApp.Router().Route(msgRoute)
		if handler == nil {
			result.Error = sdk.ABCIError(std.ErrUnknownRequest("unrecognized message type: " + msgRoute))
			return
		}

		var msgResult sdk.Result

		// run the message!
		// skip actual execution for CheckTx mode
		if mode != sdk.RunTxModeCheck {
			msgResult = handler.Process(s.Ctx, msg)
		}

		// Each message result's Data must be length prefixed in order to separate
		// each result.
		data = append(data, msgResult.Data...)
		events = append(events, msgResult.Events...)
		// TODO append msgevent from ctx. XXX XXX

		// stop execution and return on first failed message
		if !msgResult.IsOK() {
			msgLogs = append(msgLogs,
				fmt.Sprintf("msg:%d,success:%v,log:%s,events:%v",
					i, false, msgResult.Log, events))
			err = msgResult.Error
			break
		}

		msgLogs = append(msgLogs,
			fmt.Sprintf("msg:%d,success:%v,log:%s,events:%v",
				i, true, msgResult.Log, events))
	}

	result.Error = sdk.ABCIError(err)
	result.Data = data
	result.Log = strings.Join(msgLogs, "\n")
	result.GasUsed = s.Ctx.GasMeter().GasConsumed()
	result.Events = events
	return result
}

func NewSimulator(skipFailingGenesisTxs bool, stdLibPath string) (*Simulator, error) {
	rootDir := "testdir"
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("simulator")
	s := &Simulator{}

	// Get main DB.
	db := dbm.NewDB("gnoland", dbm.GoLevelDBBackend, filepath.Join(rootDir, "data"))

	// Capabilities keys.
	mainKey := store.NewStoreKey("main")
	baseKey := store.NewStoreKey("base")

	// Create BaseApp.
	mockApp := sdk.NewMockApp("gnoland", nil, db, baseKey, mainKey)

	// Set mounts for BaseApp's MultiStore.
	mockApp.MountStoreWithDB(mainKey, iavl.StoreConstructor, db)
	mockApp.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)

	// Construct keepers.
	acctKpr := auth.NewAccountKeeper(mainKey, std.ProtoBaseAccount)

	bankKpr := bank.NewBankKeeper(acctKpr)

	// vmKpr := NewVMKeeper(baseKey, mainKey, acctKpr, bankKpr, "./stdlibs")
	vmKpr := NewVMKeeper(baseKey, mainKey, acctKpr, bankKpr, stdLibPath)

	ibcChannelKeeper := NewIBCChannelKeeper(vmKpr)
	vmKpr.IBCChannelKeeper = ibcChannelKeeper

	// dispatcher := NewDispatcher(logger)
	// dispatcher.Router().AddRoute("vm", vmi.NewHandler(vmKpr))
	// dispatcher.icbChan = ibc

	// vmKpr.SetDispatcher(dispatcher)
	// Set a handler Route.
	mockApp.Router().AddRoute("auth", auth.NewHandler(acctKpr))
	mockApp.Router().AddRoute("bank", bank.NewHandler(bankKpr))
	mockApp.Router().AddRoute("vm", vmi.NewHandler(vmKpr))

	// Load latest version.
	if err := mockApp.LoadLatestVersion(); err != nil {
		return nil, err
	}

	// Initialize the VMKeeper.
	println("simulation vmKpr initialize")
	vmKpr.Initialize(mockApp.GetCacheMultiStore())

	s.mockApp = mockApp
	s.VMKpr = vmKpr
	s.AccK = acctKpr
	s.BanK = bankKpr
	s.ibcChannelKeeper = ibcChannelKeeper
	// from test machine
	s.Ctx = testCtx((s.mockApp).GetCacheMultiStore())

	return s, nil
}

func (s *Simulator) addPkgFromMemfile(ctx sdk.Context, pkgPath string, memfiles []*std.MemFile) {
	creator := crypto.AddressFromPreimage([]byte("addr1"))
	acc := s.AccK.NewAccountWithAddress(ctx, creator)
	s.AccK.SetAccount(ctx, acc)
	s.BanK.SetCoins(ctx, creator, std.MustParseCoins("10000000ugnot"))

	msgAdd := vmi.NewMsgAddPackage(creator, pkgPath, memfiles)
	msgs := []sdk.Msg{msgAdd}

	res := s.InjectMsgs(msgs, sdk.RunTxModeDeliver)
	if res.Error != nil {
		logger.Debug("Result: ", res.Error)
	} else {
		logger.Debug("Data: ", string(res.Data))
		logger.Debug("GasWanted: ", res.GasWanted)
		logger.Debug("GasUsed: ", res.GasUsed)
	}
}

func (s *Simulator) addPkgFromPath(dir string, pkgpath string) {
	// add pkg
	memPkg := gno.ReadMemPackage(dir, pkgpath)
	creator := crypto.AddressFromPreimage([]byte("addr1"))
	acc := s.AccK.NewAccountWithAddress(s.Ctx, creator)
	s.AccK.SetAccount(s.Ctx, acc)
	s.BanK.SetCoins(s.Ctx, creator, std.MustParseCoins("10000000ugnot"))

	msgAdd := vmi.MsgAddPackage{
		Creator: creator,
		Package: memPkg,
	}

	msgs := []sdk.Msg{msgAdd}

	res := s.InjectMsgs(msgs, sdk.RunTxModeDeliver)
	if res.Error != nil {
		logger.Debug("Result: ", res.Error)
	} else {
		logger.Debug("Data: ", string(res.Data))
		logger.Debug("GasWanted: ", res.GasWanted)
		logger.Debug("GasUsed: ", res.GasUsed)
	}
}

// func simuApp(cmd *command.Command, args []string, iopts interface{}) error {
func (s *Simulator) simuCall(mfs [][]*std.MemFile, callMsg []byte) (sdk.Result, error) {
	println("simu call")
	// call
	caller := crypto.AddressFromPreimage([]byte("addr2"))
	acc2 := s.AccK.NewAccountWithAddress(s.Ctx, caller)
	s.AccK.SetAccount(s.Ctx, acc2)
	s.BanK.SetCoins(s.Ctx, caller, std.MustParseCoins("20000000ugnot"))

	var msgCalls []vmi.MsgCall
	err := json.Unmarshal(callMsg, &msgCalls)
	if err != nil {
		return sdk.Result{}, err
	}
	var msgs []sdk.Msg
	for _, m := range msgCalls {
		m.Caller = caller
		msgs = append(msgs, m)
	}

	fmt.Printf("%+v \n", msgs)

	res := s.InjectMsgs(msgs, sdk.RunTxModeDeliver)
	fmt.Printf("%+v \n", res)

	if res.Error != nil {
		logger.Debug("simulation, error: ", res.Error.Error())
	} else {
		logger.Debug("simulation, Response is : ", string(res.Data))
		// println(res.GasWanted)
		// println(res.GasUsed)
	}

	return res, nil
}
