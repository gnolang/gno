package grc1155

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

const dummyURI = "ipfs://xyz"

func TestNewBasicGRC1155Token(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")
}

func TestUri(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")
	uassert.Equal(t, dummyURI, dummy.Uri())
}

func TestBalanceOf(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	_, err := dummy.BalanceOf(zeroAddress, tid1)
	uassert.Error(t, err, "should result in error")

	balanceAddr1OfToken1, err := dummy.BalanceOf(addr1, tid1)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, int64(0), balanceAddr1OfToken1)

	dummy.mintBatch(addr1, []TokenID{tid1, tid2}, []int64{10, 100})
	dummy.mintBatch(addr2, []TokenID{tid1}, []int64{20})

	balanceAddr1OfToken1, err = dummy.BalanceOf(addr1, tid1)
	uassert.NoError(t, err, "should not result in error")
	balanceAddr1OfToken2, err := dummy.BalanceOf(addr1, tid2)
	uassert.NoError(t, err, "should not result in error")
	balanceAddr2OfToken1, err := dummy.BalanceOf(addr2, tid1)
	uassert.NoError(t, err, "should not result in error")

	uassert.Equal(t, int64(10), balanceAddr1OfToken1)
	uassert.Equal(t, int64(100), balanceAddr1OfToken2)
	uassert.Equal(t, int64(20), balanceAddr2OfToken1)
}

func TestBalanceOfBatch(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	balanceBatch, err := dummy.BalanceOfBatch([]address{addr1, addr2}, []TokenID{tid1, tid2})
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, int64(0), balanceBatch[0])
	uassert.Equal(t, int64(0), balanceBatch[1])

	dummy.mintBatch(addr1, []TokenID{tid1}, []int64{10})
	dummy.mintBatch(addr2, []TokenID{tid2}, []int64{20})

	balanceBatch, err = dummy.BalanceOfBatch([]address{addr1, addr2}, []TokenID{tid1, tid2})
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, int64(10), balanceBatch[0])
	uassert.Equal(t, int64(20), balanceBatch[1])
}

func TestIsApprovedForAll(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	isApprovedForAll := dummy.IsApprovedForAll(addr1, addr2)
	uassert.False(t, isApprovedForAll)
}

func TestSetApprovalForAll(t *testing.T) {
	//alice := testutils.TestAddress("alice")
	//testing.SetOriginCaller(alice)

	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := runtime.OriginCaller()
	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	isApprovedForAll := dummy.IsApprovedForAll(caller, addr)
	uassert.False(t, isApprovedForAll)

	err := dummy.SetApprovalForAll(addr, true)
	uassert.NoError(t, err, "should not result in error")

	isApprovedForAll = dummy.IsApprovedForAll(caller, addr)
	uassert.True(t, isApprovedForAll)

	err = dummy.SetApprovalForAll(addr, false)
	uassert.NoError(t, err, "should not result in error")

	isApprovedForAll = dummy.IsApprovedForAll(caller, addr)
	uassert.False(t, isApprovedForAll)
}

func TestSafeTransferFrom(t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)

	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := runtime.OriginCaller()
	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	tid := TokenID("1")

	dummy.mintBatch(caller, []TokenID{tid}, []int64{100})

	err := dummy.SafeTransferFrom(caller, zeroAddress, tid, 10)
	uassert.Error(t, err, "should result in error")

	err = dummy.SafeTransferFrom(caller, addr, tid, 160)
	uassert.Error(t, err, "should result in error")

	err = dummy.SafeTransferFrom(caller, addr, tid, 60)
	uassert.NoError(t, err, "should not result in error")

	// Check balance of caller after transfer
	balanceOfCaller, err := dummy.BalanceOf(caller, tid)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, int64(40), balanceOfCaller)

	// Check balance of addr after transfer
	balanceOfAddr, err := dummy.BalanceOf(addr, tid)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, int64(60), balanceOfAddr)
}

func TestSafeBatchTransferFrom(t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)

	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := runtime.OriginCaller()
	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	dummy.mintBatch(caller, []TokenID{tid1, tid2}, []int64{10, 100})

	err := dummy.SafeBatchTransferFrom(caller, zeroAddress, []TokenID{tid1, tid2}, []int64{4, 60})
	uassert.Error(t, err, "should result in error")
	err = dummy.SafeBatchTransferFrom(caller, addr, []TokenID{tid1, tid2}, []int64{40, 60})
	uassert.Error(t, err, "should result in error")
	err = dummy.SafeBatchTransferFrom(caller, addr, []TokenID{tid1}, []int64{40, 60})
	uassert.Error(t, err, "should result in error")
	err = dummy.SafeBatchTransferFrom(caller, addr, []TokenID{tid1, tid2}, []int64{4, 60})
	uassert.NoError(t, err, "should not result in error")

	balanceBatch, err := dummy.BalanceOfBatch([]address{caller, addr, caller, addr}, []TokenID{tid1, tid1, tid2, tid2})
	uassert.NoError(t, err, "should not result in error")

	// Check token1's balance of caller after batch transfer
	uassert.Equal(t, int64(6), balanceBatch[0])

	// Check token1's balance of addr after batch transfer
	uassert.Equal(t, int64(4), balanceBatch[1])

	// Check token2's balance of caller after batch transfer
	uassert.Equal(t, int64(40), balanceBatch[2])

	// Check token2's balance of addr after batch transfer
	uassert.Equal(t, int64(60), balanceBatch[3])
}

func TestSafeMint(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	err := dummy.SafeMint(zeroAddress, tid1, 100)
	uassert.Error(t, err, "should result in error")
	err = dummy.SafeMint(addr1, tid1, 100)
	uassert.NoError(t, err, "should not result in error")
	err = dummy.SafeMint(addr1, tid2, 200)
	uassert.NoError(t, err, "should not result in error")
	err = dummy.SafeMint(addr2, tid1, 50)
	uassert.NoError(t, err, "should not result in error")

	balanceBatch, err := dummy.BalanceOfBatch([]address{addr1, addr2, addr1}, []TokenID{tid1, tid1, tid2})
	uassert.NoError(t, err, "should not result in error")
	// Check token1's balance of addr1 after mint
	uassert.Equal(t, int64(100), balanceBatch[0])
	// Check token1's balance of addr2 after mint
	uassert.Equal(t, int64(50), balanceBatch[1])
	// Check token2's balance of addr1 after mint
	uassert.Equal(t, int64(200), balanceBatch[2])
}

func TestSafeBatchMint(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	err := dummy.SafeBatchMint(zeroAddress, []TokenID{tid1, tid2}, []int64{100, 200})
	uassert.Error(t, err, "should result in error")
	err = dummy.SafeBatchMint(addr1, []TokenID{tid1, tid2}, []int64{100, 200})
	uassert.NoError(t, err, "should not result in error")
	err = dummy.SafeBatchMint(addr2, []TokenID{tid1, tid2}, []int64{300, 400})
	uassert.NoError(t, err, "should not result in error")

	balanceBatch, err := dummy.BalanceOfBatch([]address{addr1, addr2, addr1, addr2}, []TokenID{tid1, tid1, tid2, tid2})
	uassert.NoError(t, err, "should not result in error")
	// Check token1's balance of addr1 after batch mint
	uassert.Equal(t, int64(100), balanceBatch[0])
	// Check token1's balance of addr2 after batch mint
	uassert.Equal(t, int64(300), balanceBatch[1])
	// Check token2's balance of addr1 after batch mint
	uassert.Equal(t, int64(200), balanceBatch[2])
	// Check token2's balance of addr2 after batch mint
	uassert.Equal(t, int64(400), balanceBatch[3])
}

func TestBurn(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")

	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	dummy.mintBatch(addr, []TokenID{tid1, tid2}, []int64{100, 200})
	err := dummy.Burn(zeroAddress, tid1, int64(60))
	uassert.Error(t, err, "should result in error")
	err = dummy.Burn(addr, tid1, int64(160))
	uassert.Error(t, err, "should result in error")
	err = dummy.Burn(addr, tid1, int64(60))
	uassert.NoError(t, err, "should not result in error")
	err = dummy.Burn(addr, tid2, int64(60))
	uassert.NoError(t, err, "should not result in error")
	balanceBatch, err := dummy.BalanceOfBatch([]address{addr, addr}, []TokenID{tid1, tid2})
	uassert.NoError(t, err, "should not result in error")

	// Check token1's balance of addr after burn
	uassert.Equal(t, int64(40), balanceBatch[0])
	// Check token2's balance of addr after burn
	uassert.Equal(t, int64(140), balanceBatch[1])
}

func TestBatchBurn(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	uassert.True(t, dummy != nil, "should not be nil")

	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	dummy.mintBatch(addr, []TokenID{tid1, tid2}, []int64{100, 200})
	err := dummy.BatchBurn(zeroAddress, []TokenID{tid1, tid2}, []int64{60, 60})
	uassert.Error(t, err, "should result in error")
	err = dummy.BatchBurn(addr, []TokenID{tid1, tid2}, []int64{160, 60})
	uassert.Error(t, err, "should result in error")
	err = dummy.BatchBurn(addr, []TokenID{tid1, tid2}, []int64{60, 60})
	uassert.NoError(t, err, "should not result in error")
	balanceBatch, err := dummy.BalanceOfBatch([]address{addr, addr}, []TokenID{tid1, tid2})
	uassert.NoError(t, err, "should not result in error")

	// Check token1's balance of addr after batch burn
	uassert.Equal(t, int64(40), balanceBatch[0])
	// Check token2's balance of addr after batch burn
	uassert.Equal(t, int64(140), balanceBatch[1])
}
