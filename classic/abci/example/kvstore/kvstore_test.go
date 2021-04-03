package kvstore

import (
	"fmt"
	"io/ioutil"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/classic/libs/common"
	"github.com/tendermint/classic/libs/log"

	abcicli "github.com/tendermint/classic/abci/client"
	abciserver "github.com/tendermint/classic/abci/server"
	abci "github.com/tendermint/classic/abci/types"
)

const (
	testKey   = "abc"
	testValue = "def"
)

func testKVStore(t *testing.T, app abci.Application, tx []byte, key, value string) {
	req := abci.RequestDeliverTx{Tx: tx}
	ar := app.DeliverTx(req)
	require.False(t, ar.IsErr(), ar)
	// repeating tx doesn't raise error
	ar = app.DeliverTx(req)
	require.False(t, ar.IsErr(), ar)

	// make sure query is fine
	resQuery := app.Query(abci.RequestQuery{
		Path: "/store",
		Data: []byte(key),
	})
	require.Equal(t, nil, resQuery.Error)
	require.Equal(t, value, string(resQuery.Value))

	// make sure proof is fine
	resQuery = app.Query(abci.RequestQuery{
		Path:  "/store",
		Data:  []byte(key),
		Prove: true,
	})
	require.EqualValues(t, nil, resQuery.Error)
	require.Equal(t, value, string(resQuery.Value))
}

func TestKVStoreKV(t *testing.T) {
	kvstore := NewKVStoreApplication()
	key := testKey
	value := key
	tx := []byte(key)
	testKVStore(t, kvstore, tx, key, value)

	value = testValue
	tx = []byte(key + "=" + value)
	testKVStore(t, kvstore, tx, key, value)
}

func TestPersistentKVStoreKV(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "abci-kvstore-test") // TODO
	if err != nil {
		t.Fatal(err)
	}
	kvstore := NewPersistentKVStoreApplication(dir)
	key := testKey
	value := key
	tx := []byte(key)
	testKVStore(t, kvstore, tx, key, value)

	value = testValue
	tx = []byte(key + "=" + value)
	testKVStore(t, kvstore, tx, key, value)
}

func TestPersistentKVStoreInfo(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "abci-kvstore-test") // TODO
	if err != nil {
		t.Fatal(err)
	}
	kvstore := NewPersistentKVStoreApplication(dir)
	InitKVStore(kvstore)
	height := int64(0)

	resInfo := kvstore.Info(abci.RequestInfo{})
	if resInfo.LastBlockHeight != height {
		t.Fatalf("expected height of %d, got %d", height, resInfo.LastBlockHeight)
	}

	// make and apply block
	height = int64(1)
	hash := []byte("foo")
	header := abci.MockHeader{
		Height: height,
	}
	kvstore.BeginBlock(abci.RequestBeginBlock{Hash: hash, Header: header})
	kvstore.EndBlock(abci.RequestEndBlock{Height: header.Height})
	kvstore.Commit()

	resInfo = kvstore.Info(abci.RequestInfo{})
	if resInfo.LastBlockHeight != height {
		t.Fatalf("expected height of %d, got %d", height, resInfo.LastBlockHeight)
	}

}

// add a validator, remove a validator, update a validator
func TestValUpdates(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "abci-kvstore-test") // TODO
	if err != nil {
		t.Fatal(err)
	}
	kvstore := NewPersistentKVStoreApplication(dir)

	// init with some validators
	total := 10
	nInit := 5
	vals := RandVals(total)
	// iniitalize with the first nInit
	kvstore.InitChain(abci.RequestInitChain{
		Validators: vals[:nInit],
	})

	vals1, vals2 := vals[:nInit], kvstore.Validators()
	valsEqual(t, vals1, vals2)

	var v1, v2, v3 abci.ValidatorUpdate

	// add some validators
	v1, v2 = vals[nInit], vals[nInit+1]
	diff := []abci.ValidatorUpdate{v1, v2}
	tx1 := MakeValSetChangeTx(v1.PubKey, v1.Power)
	tx2 := MakeValSetChangeTx(v2.PubKey, v2.Power)

	makeApplyBlock(t, kvstore, 1, diff, tx1, tx2)

	vals1, vals2 = vals[:nInit+2], kvstore.Validators()
	valsEqual(t, vals1, vals2)

	// remove some validators
	v1, v2, v3 = vals[nInit-2], vals[nInit-1], vals[nInit]
	v1.Power = 0
	v2.Power = 0
	v3.Power = 0
	diff = []abci.ValidatorUpdate{v1, v2, v3}
	tx1 = MakeValSetChangeTx(v1.PubKey, v1.Power)
	tx2 = MakeValSetChangeTx(v2.PubKey, v2.Power)
	tx3 := MakeValSetChangeTx(v3.PubKey, v3.Power)

	makeApplyBlock(t, kvstore, 2, diff, tx1, tx2, tx3)

	vals1 = append(vals[:nInit-2], vals[nInit+1]) // nolint: gocritic
	vals2 = kvstore.Validators()
	valsEqual(t, vals1, vals2)

	// update some validators
	v1 = vals[0]
	if v1.Power == 5 {
		v1.Power = 6
	} else {
		v1.Power = 5
	}
	diff = []abci.ValidatorUpdate{v1}
	tx1 = MakeValSetChangeTx(v1.PubKey, v1.Power)

	makeApplyBlock(t, kvstore, 3, diff, tx1)

	vals1 = append([]abci.ValidatorUpdate{v1}, vals1[1:]...)
	vals2 = kvstore.Validators()
	valsEqual(t, vals1, vals2)

}

func makeApplyBlock(t *testing.T, kvstore abci.Application, heightInt int, diff []abci.ValidatorUpdate, txs ...[]byte) {
	// make and apply block
	height := int64(heightInt)
	hash := []byte("foo")
	header := abci.MockHeader{
		Height: height,
	}
	kvstore.BeginBlock(abci.RequestBeginBlock{Hash: hash, Header: header})
	for _, tx := range txs {
		if r := kvstore.DeliverTx(abci.RequestDeliverTx{Tx: tx}); r.IsErr() {
			t.Fatal(r)
		}
	}
	resEndBlock := kvstore.EndBlock(abci.RequestEndBlock{Height: header.Height})
	kvstore.Commit()

	valsEqual(t, diff, resEndBlock.ValidatorUpdates)

}

// order doesn't matter
func valsEqual(t *testing.T, vals1, vals2 []abci.ValidatorUpdate) {
	if len(vals1) != len(vals2) {
		t.Fatalf("vals dont match in len. got %d, expected %d", len(vals2), len(vals1))
	}
	sort.Sort(abci.ValidatorUpdates(vals1))
	sort.Sort(abci.ValidatorUpdates(vals2))
	for i, v1 := range vals1 {
		v2 := vals2[i]
		if !v1.PubKey.Equals(v2.PubKey) ||
			v1.Power != v2.Power {
			t.Fatalf("vals dont match at index %d. got %X/%d , expected %X/%d", i, v2.PubKey, v2.Power, v1.PubKey, v1.Power)
		}
	}
}

func makeSocketClientServer(app abci.Application, name string) (abcicli.Client, cmn.Service, error) {
	// Start the listener
	socket := fmt.Sprintf("unix://%s.sock", name)
	logger := log.TestingLogger()

	server := abciserver.NewSocketServer(socket, app)
	server.SetLogger(logger.With("module", "abci-server"))
	if err := server.Start(); err != nil {
		return nil, nil, err
	}

	// Connect to the socket
	client := abcicli.NewSocketClient(socket, false)
	client.SetLogger(logger.With("module", "abci-client"))
	if err := client.Start(); err != nil {
		server.Stop()
		return nil, nil, err
	}

	return client, server, nil
}

func TestClientServer(t *testing.T) {
	// set up socket app
	kvstore := NewKVStoreApplication()
	client, server, err := makeSocketClientServer(kvstore, "kvstore-socket")
	require.Nil(t, err)
	defer server.Stop()
	defer client.Stop()

	runClientTests(t, client)
}

func runClientTests(t *testing.T, client abcicli.Client) {
	// run some tests....
	key := testKey
	value := key
	tx := []byte(key)
	testClient(t, client, tx, key, value)

	value = testValue
	tx = []byte(key + "=" + value)
	testClient(t, client, tx, key, value)
}

func testClient(t *testing.T, app abcicli.Client, tx []byte, key, value string) {
	ar, err := app.DeliverTxSync(abci.RequestDeliverTx{Tx: tx})
	require.NoError(t, err)
	require.False(t, ar.IsErr(), ar)
	// repeating tx doesn't raise error
	ar, err = app.DeliverTxSync(abci.RequestDeliverTx{Tx: tx})
	require.NoError(t, err)
	require.False(t, ar.IsErr(), ar)

	// make sure query is fine
	resQuery, err := app.QuerySync(abci.RequestQuery{
		Path: "/store",
		Data: []byte(key),
	})
	require.Nil(t, err)
	require.Equal(t, nil, resQuery.Error)
	require.Equal(t, value, string(resQuery.Value))

	// make sure proof is fine
	resQuery, err = app.QuerySync(abci.RequestQuery{
		Path:  "/store",
		Data:  []byte(key),
		Prove: true,
	})
	require.Nil(t, err)
	require.Equal(t, nil, resQuery.Error)
	require.Equal(t, value, string(resQuery.Value))
}
