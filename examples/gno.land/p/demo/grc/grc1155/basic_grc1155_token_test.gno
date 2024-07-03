package grc1155

import (
	"std"
	"testing"
)

const dummyURI = "ipfs://xyz"

func TestNewBasicGRC1155Token(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		return t.Errorf("should not be nil")
	}
}

func TestUri(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	uri := dummy.Uri()
	if uri != dummyURI {
		t.Errorf("expected: (%s), got: (%s)", dummyURI, uri)
	}
}

func TestBalanceOf(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	balanceZeroAddressOfToken1, err := dummy.BalanceOf(zeroAddress, tid1)
	if err == nil {
		t.Errorf("should result in error")
	}
	balanceAddr1OfToken1, err := dummy.BalanceOf(addr1, tid1)
	if err != nil {
		t.Errorf("should not result in error")
	}
	if balanceAddr1OfToken1 != 0 {
		t.Errorf("expected: (%d), got: (%d)", 0, balanceAddr1OfToken1)
	}

	dummy.mintBatch(addr1, []TokenID{tid1, tid2}, []uint64{10, 100})
	dummy.mintBatch(addr2, []TokenID{tid1}, []uint64{20})

	balanceAddr1OfToken1, err = dummy.BalanceOf(addr1, tid1)
	if err != nil {
		t.Errorf("should not result in error")
	}
	balanceAddr1OfToken2, err := dummy.BalanceOf(addr1, tid2)
	if err != nil {
		t.Errorf("should not result in error")
	}
	balanceAddr2OfToken1, err := dummy.BalanceOf(addr2, tid1)
	if err != nil {
		t.Errorf("should not result in error")
	}

	if balanceAddr1OfToken1 != 10 {
		t.Errorf("expected: (%d), got: (%d)", 10, balanceAddr1OfToken1)
	}
	if balanceAddr1OfToken2 != 100 {
		t.Errorf("expected: (%d), got: (%d)", 100, balanceAddr1OfToken2)
	}
	if balanceAddr2OfToken1 != 20 {
		t.Errorf("expected: (%d), got: (%d)", 20, balanceAddr2OfToken1)
	}
}

func TestBalanceOfBatch(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	balanceBatch, err := dummy.BalanceOfBatch([]std.Address{addr1, addr2}, []TokenID{tid1, tid2})
	if err != nil {
		t.Errorf("should not result in error")
	}

	if balanceBatch[0] != 0 {
		t.Errorf("expected: (%d), got: (%d)", 0, balanceBatch[0])
	}
	if balanceBatch[1] != 0 {
		t.Errorf("expected: (%d), got: (%d)", 0, balanceBatch[1])
	}

	dummy.mintBatch(addr1, []TokenID{tid1}, []uint64{10})
	dummy.mintBatch(addr2, []TokenID{tid2}, []uint64{20})

	balanceBatch, err = dummy.BalanceOfBatch([]std.Address{addr1, addr2}, []TokenID{tid1, tid2})
	if err != nil {
		t.Errorf("should not result in error")
	}

	if balanceBatch[0] != 10 {
		t.Errorf("expected: (%d), got: (%d)", 10, balanceBatch[0])
	}
	if balanceBatch[1] != 20 {
		t.Errorf("expected: (%d), got: (%d)", 20, balanceBatch[1])
	}
}

func TestIsApprovedForAll(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	isApprovedForAll := dummy.IsApprovedForAll(addr1, addr2)
	if isApprovedForAll != false {
		t.Errorf("expected: (%v), got: (%v)", false, isApprovedForAll)
	}
}

func TestSetApprovalForAll(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	caller := std.GetOrigCaller()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	isApprovedForAll := dummy.IsApprovedForAll(caller, addr)
	if isApprovedForAll != false {
		t.Errorf("expected: (%v), got: (%v)", false, isApprovedForAll)
	}

	err := dummy.SetApprovalForAll(addr, true)
	if err != nil {
		t.Errorf("should not result in error")
	}

	isApprovedForAll = dummy.IsApprovedForAll(caller, addr)
	if isApprovedForAll != true {
		t.Errorf("expected: (%v), got: (%v)", true, isApprovedForAll)
	}

	err = dummy.SetApprovalForAll(addr, false)
	if err != nil {
		t.Errorf("should not result in error")
	}

	isApprovedForAll = dummy.IsApprovedForAll(caller, addr)
	if isApprovedForAll != false {
		t.Errorf("expected: (%v), got: (%v)", false, isApprovedForAll)
	}
}

func TestSafeTransferFrom(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	caller := std.GetOrigCaller()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	tid := TokenID("1")

	dummy.mintBatch(caller, []TokenID{tid}, []uint64{100})

	err := dummy.SafeTransferFrom(caller, zeroAddress, tid, 10)
	if err == nil {
		t.Errorf("should result in error")
	}

	err = dummy.SafeTransferFrom(caller, addr, tid, 160)
	if err == nil {
		t.Errorf("should result in error")
	}

	err = dummy.SafeTransferFrom(caller, addr, tid, 60)
	if err != nil {
		t.Errorf("should not result in error")
	}

	// Check balance of caller after transfer
	balanceOfCaller, err := dummy.BalanceOf(caller, tid)
	if err != nil {
		t.Errorf("should not result in error")
	}
	if balanceOfCaller != 40 {
		t.Errorf("expected: (%d), got: (%d)", 40, balanceOfCaller)
	}

	// Check balance of addr after transfer
	balanceOfAddr, err := dummy.BalanceOf(addr, tid)
	if err != nil {
		t.Errorf("should not result in error")
	}
	if balanceOfAddr != 60 {
		t.Errorf("expected: (%d), got: (%d)", 60, balanceOfAddr)
	}
}

func TestSafeBatchTransferFrom(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	caller := std.GetOrigCaller()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	dummy.mintBatch(caller, []TokenID{tid1, tid2}, []uint64{10, 100})

	err := dummy.SafeBatchTransferFrom(caller, zeroAddress, []TokenID{tid1, tid2}, []uint64{4, 60})
	if err == nil {
		t.Errorf("should result in error")
	}
	err = dummy.SafeBatchTransferFrom(caller, addr, []TokenID{tid1, tid2}, []uint64{40, 60})
	if err == nil {
		t.Errorf("should result in error")
	}
	err = dummy.SafeBatchTransferFrom(caller, addr, []TokenID{tid1}, []uint64{40, 60})
	if err == nil {
		t.Errorf("should result in error")
	}
	err = dummy.SafeBatchTransferFrom(caller, addr, []TokenID{tid1, tid2}, []uint64{4, 60})
	if err != nil {
		t.Errorf("should not result in error")
	}

	balanceBatch, err := dummy.BalanceOfBatch([]std.Address{caller, addr, caller, addr}, []TokenID{tid1, tid1, tid2, tid2})
	if err != nil {
		t.Errorf("should not result in error")
	}

	// Check token1's balance of caller after batch transfer
	if balanceBatch[0] != 6 {
		t.Errorf("expected: (%d), got: (%d)", 6, balanceBatch[0])
	}

	// Check token1's balance of addr after batch transfer
	if balanceBatch[1] != 4 {
		t.Errorf("expected: (%d), got: (%d)", 4, balanceBatch[1])
	}

	// Check token2's balance of caller after batch transfer
	if balanceBatch[2] != 40 {
		t.Errorf("expected: (%d), got: (%d)", 40, balanceBatch[2])
	}

	// Check token2's balance of addr after batch transfer
	if balanceBatch[3] != 60 {
		t.Errorf("expected: (%d), got: (%d)", 60, balanceBatch[3])
	}
}

func TestSafeMint(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	err := dummy.SafeMint(zeroAddress, tid1, 100)
	if err == nil {
		t.Errorf("should result in error")
	}
	err = dummy.SafeMint(addr1, tid1, 100)
	if err != nil {
		t.Errorf("should not result in error")
	}
	err = dummy.SafeMint(addr1, tid2, 200)
	if err != nil {
		t.Errorf("should not result in error")
	}
	err = dummy.SafeMint(addr2, tid1, 50)
	if err != nil {
		t.Errorf("should not result in error")
	}

	balanceBatch, err := dummy.BalanceOfBatch([]std.Address{addr1, addr2, addr1}, []TokenID{tid1, tid1, tid2})
	if err != nil {
		t.Errorf("should not result in error")
	}
	// Check token1's balance of addr1 after mint
	if balanceBatch[0] != 100 {
		t.Errorf("expected: (%d), got: (%d)", 100, balanceBatch[0])
	}
	// Check token1's balance of addr2 after mint
	if balanceBatch[1] != 50 {
		t.Errorf("expected: (%d), got: (%d)", 50, balanceBatch[1])
	}
	// Check token2's balance of addr1 after mint
	if balanceBatch[2] != 200 {
		t.Errorf("expected: (%d), got: (%d)", 200, balanceBatch[2])
	}
}

func TestSafeBatchMint(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	err := dummy.SafeBatchMint(zeroAddress, []TokenID{tid1, tid2}, []uint64{100, 200})
	if err == nil {
		t.Errorf("should result in error")
	}
	err = dummy.SafeBatchMint(addr1, []TokenID{tid1, tid2}, []uint64{100, 200})
	if err != nil {
		t.Errorf("should not result in error")
	}
	err = dummy.SafeBatchMint(addr2, []TokenID{tid1, tid2}, []uint64{300, 400})
	if err != nil {
		t.Errorf("should not result in error")
	}

	balanceBatch, err := dummy.BalanceOfBatch([]std.Address{addr1, addr2, addr1, addr2}, []TokenID{tid1, tid1, tid2, tid2})
	if err != nil {
		t.Errorf("should not result in error")
	}
	// Check token1's balance of addr1 after batch mint
	if balanceBatch[0] != 100 {
		t.Errorf("expected: (%d), got: (%d)", 100, balanceBatch[0])
	}
	// Check token1's balance of addr2 after batch mint
	if balanceBatch[1] != 300 {
		t.Errorf("expected: (%d), got: (%d)", 300, balanceBatch[1])
	}
	// Check token2's balance of addr1 after batch mint
	if balanceBatch[2] != 200 {
		t.Errorf("expected: (%d), got: (%d)", 200, balanceBatch[2])
	}
	// Check token2's balance of addr2 after batch mint
	if balanceBatch[3] != 400 {
		t.Errorf("expected: (%d), got: (%d)", 400, balanceBatch[3])
	}
}

func TestBurn(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	dummy.mintBatch(addr, []TokenID{tid1, tid2}, []uint64{100, 200})
	err := dummy.Burn(zeroAddress, tid1, 60)
	if err == nil {
		t.Errorf("should result in error")
	}
	err = dummy.Burn(addr, tid1, 160)
	if err == nil {
		t.Errorf("should result in error")
	}
	err = dummy.Burn(addr, tid1, 60)
	if err != nil {
		t.Errorf("should not result in error")
	}
	err = dummy.Burn(addr, tid2, 60)
	if err != nil {
		t.Errorf("should not result in error")
	}
	balanceBatch, err := dummy.BalanceOfBatch([]std.Address{addr, addr}, []TokenID{tid1, tid2})
	if err != nil {
		t.Errorf("should not result in error")
	}

	// Check token1's balance of addr after burn
	if balanceBatch[0] != 40 {
		t.Errorf("expected: (%d), got: (%d)", 40, balanceBatch[0])
	}
	// Check token2's balance of addr after burn
	if balanceBatch[1] != 140 {
		t.Errorf("expected: (%d), got: (%d)", 140, balanceBatch[1])
	}
}

func TestBatchBurn(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	tid1 := TokenID("1")
	tid2 := TokenID("2")

	dummy.mintBatch(addr, []TokenID{tid1, tid2}, []uint64{100, 200})
	err := dummy.BatchBurn(zeroAddress, []TokenID{tid1, tid2}, []uint64{60, 60})
	if err == nil {
		t.Errorf("should result in error")
	}
	err = dummy.BatchBurn(addr, []TokenID{tid1, tid2}, []uint64{160, 60})
	if err == nil {
		t.Errorf("should result in error")
	}
	err = dummy.BatchBurn(addr, []TokenID{tid1, tid2}, []uint64{60, 60})
	if err != nil {
		t.Errorf("should not result in error")
	}
	balanceBatch, err := dummy.BalanceOfBatch([]std.Address{addr, addr}, []TokenID{tid1, tid2})
	if err != nil {
		t.Errorf("should not result in error")
	}

	// Check token1's balance of addr after batch burn
	if balanceBatch[0] != 40 {
		t.Errorf("expected: (%d), got: (%d)", 40, balanceBatch[0])
	}
	// Check token2's balance of addr after batch burn
	if balanceBatch[1] != 140 {
		t.Errorf("expected: (%d), got: (%d)", 140, balanceBatch[1])
	}
}
