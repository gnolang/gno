package rootmulti

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"

	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func TestVerifyIAVLStoreQueryProof(t *testing.T) {
	t.Parallel()

	// Create main tree for testing.
	db := memdb.NewMemDB()
	opts := types.StoreOptions{
		PruningOptions: types.PruneNothing,
	}
	iStore := iavl.StoreConstructor(db, opts)
	store := iStore.(*iavl.Store)
	store.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := store.Commit()

	// Get Proof
	res := store.Query(abci.RequestQuery{
		Path:  "/key", // required path to get key/value+proof
		Data:  []byte("MYKEY"),
		Prove: true,
	})
	require.NotNil(t, res.Proof)

	// Verify proof.
	prt := DefaultProofRuntime()
	err := prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY", []byte("MYVALUE"))
	require.Nil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY_NOT", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY/MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY", []byte("MYVALUE_NOT"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY", []byte(nil))
	require.NotNil(t, err)
}

func TestVerifyMultiStoreQueryProof(t *testing.T) {
	t.Parallel()

	// Create main tree for testing.
	db := memdb.NewMemDB()
	store := NewMultiStore(db)
	iavlStoreKey := types.NewStoreKey("iavlStoreKey")

	store.MountStoreWithDB(iavlStoreKey, iavl.StoreConstructor, nil)
	store.LoadVersion(0)

	iavlStore := store.GetCommitStore(iavlStoreKey).(*iavl.Store)
	iavlStore.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := store.Commit()

	// Get Proof
	res := store.Query(abci.RequestQuery{
		Path:  "/iavlStoreKey/key", // required path to get key/value+proof
		Data:  []byte("MYKEY"),
		Prove: true,
	})
	require.NotNil(t, res.Proof)

	// Verify proof.
	prt := DefaultProofRuntime()
	err := prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY", []byte("MYVALUE"))
	require.Nil(t, err)

	// Verify proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY", []byte("MYVALUE"))
	require.Nil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY_NOT", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY/MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "iavlStoreKey/MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY", []byte("MYVALUE_NOT"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY", []byte(nil))
	require.NotNil(t, err)
}

func TestVerifyMultiStoreQueryProofAbsence(t *testing.T) {
	t.Parallel()

	// Create main tree for testing.
	db := memdb.NewMemDB()
	store := NewMultiStore(db)
	iavlStoreKey := types.NewStoreKey("iavlStoreKey")

	store.MountStoreWithDB(iavlStoreKey, iavl.StoreConstructor, nil)
	store.LoadVersion(0)

	iavlStore := store.GetCommitStore(iavlStoreKey).(*iavl.Store)
	iavlStore.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := store.Commit() // Commit with empty iavl store.

	// Get Proof
	res := store.Query(abci.RequestQuery{
		Path:  "/iavlStoreKey/key", // required path to get key/value+proof
		Data:  []byte("MYABSENTKEY"),
		Prove: true,
	})
	require.NotNil(t, res.Proof)

	// Verify proof.
	prt := DefaultProofRuntime()
	err := prt.VerifyAbsence(res.Proof, cid.Hash, "/iavlStoreKey/MYABSENTKEY")
	require.Nil(t, err)

	// Verify (bad) proof.
	prt = DefaultProofRuntime()
	err = prt.VerifyAbsence(res.Proof, cid.Hash, "/MYABSENTKEY")
	require.NotNil(t, err)

	// Verify (bad) proof.
	prt = DefaultProofRuntime()
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYABSENTKEY", []byte(""))
	require.NotNil(t, err)
}
