package grc1155

import (
	"chain/runtime"
	"math/overflow"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

type basicGRC1155Token struct {
	uri               string
	balances          avl.Tree // "TokenId:Address" -> int64
	operatorApprovals avl.Tree // "OwnerAddress:OperatorAddress" -> bool
}

var _ IGRC1155 = (*basicGRC1155Token)(nil)

// Returns new basic GRC1155 token
func NewBasicGRC1155Token(uri string) *basicGRC1155Token {
	return &basicGRC1155Token{
		uri:               uri,
		balances:          avl.Tree{},
		operatorApprovals: avl.Tree{},
	}
}

func (s *basicGRC1155Token) Uri() string { return s.uri }

// BalanceOf returns the input address's balance of the token type requested
func (s *basicGRC1155Token) BalanceOf(addr address, tid TokenID) (int64, error) {
	if !isValidAddress(addr) {
		return 0, ErrInvalidAddress
	}

	key := string(tid) + ":" + addr.String()
	balance, found := s.balances.Get(key)
	if !found {
		return 0, nil
	}

	return balance.(int64), nil
}

// BalanceOfBatch returns the balance of multiple account/token pairs
func (s *basicGRC1155Token) BalanceOfBatch(owners []address, batch []TokenID) ([]int64, error) {
	if len(owners) != len(batch) {
		return nil, ErrMismatchLength
	}

	balanceOfBatch := make([]int64, len(owners))

	for i := 0; i < len(owners); i++ {
		balanceOfBatch[i], _ = s.BalanceOf(owners[i], batch[i])
	}

	return balanceOfBatch, nil
}

// SetApprovalForAll can approve the operator to operate on all tokens
func (s *basicGRC1155Token) SetApprovalForAll(operator address, approved bool) error {
	if !isValidAddress(operator) {
		return ErrInvalidAddress
	}

	caller := runtime.OriginCaller()
	return s.setApprovalForAll(caller, operator, approved)
}

// IsApprovedForAll returns true if operator is the owner or is approved for all by the owner.
// Otherwise, returns false
func (s *basicGRC1155Token) IsApprovedForAll(owner, operator address) bool {
	if operator == owner {
		return true
	}
	key := owner.String() + ":" + operator.String()
	approved, found := s.operatorApprovals.Get(key)
	if !found {
		return false
	}

	ab, ok := approved.(bool)
	return ok && ab
}

// Safely transfers `tokenId` token from `from` to `to`, checking that
// contract recipients are aware of the GRC1155 protocol to prevent
// tokens from being forever locked.
func (s *basicGRC1155Token) SafeTransferFrom(from, to address, tid TokenID, amount int64) error {
	caller := runtime.OriginCaller()
	if !s.IsApprovedForAll(caller, from) {
		return ErrCallerIsNotOwnerOrApproved
	}

	err := s.safeBatchTransferFrom(from, to, []TokenID{tid}, []int64{amount})
	if err != nil {
		return err
	}

	if !s.doSafeTransferAcceptanceCheck(caller, from, to, tid, amount) {
		return ErrTransferToRejectedOrNonGRC1155Receiver
	}

	emit(&TransferSingleEvent{caller, from, to, tid, amount})

	return nil
}

// Safely transfers a `batch` of tokens from `from` to `to`, checking that
// contract recipients are aware of the GRC1155 protocol to prevent
// tokens from being forever locked.
func (s *basicGRC1155Token) SafeBatchTransferFrom(from, to address, batch []TokenID, amounts []int64) error {
	caller := runtime.OriginCaller()
	if !s.IsApprovedForAll(caller, from) {
		return ErrCallerIsNotOwnerOrApproved
	}

	err := s.safeBatchTransferFrom(from, to, batch, amounts)
	if err != nil {
		return err
	}

	if !s.doSafeBatchTransferAcceptanceCheck(caller, from, to, batch, amounts) {
		return ErrTransferToRejectedOrNonGRC1155Receiver
	}

	emit(&TransferBatchEvent{caller, from, to, batch, amounts})

	return nil
}

// Creates `amount` tokens of token type `id`, and assigns them to `to`. Also checks that
// contract recipients are using GRC1155 protocol.
func (s *basicGRC1155Token) SafeMint(to address, tid TokenID, amount int64) error {
	caller := runtime.OriginCaller()

	err := s.mintBatch(to, []TokenID{tid}, []int64{amount})
	if err != nil {
		return err
	}

	if !s.doSafeTransferAcceptanceCheck(caller, zeroAddress, to, tid, amount) {
		return ErrTransferToRejectedOrNonGRC1155Receiver
	}

	emit(&TransferSingleEvent{caller, zeroAddress, to, tid, amount})

	return nil
}

// Batch version of `SafeMint()`. Also checks that
// contract recipients are using GRC1155 protocol.
func (s *basicGRC1155Token) SafeBatchMint(to address, batch []TokenID, amounts []int64) error {
	caller := runtime.OriginCaller()

	err := s.mintBatch(to, batch, amounts)
	if err != nil {
		return err
	}

	if !s.doSafeBatchTransferAcceptanceCheck(caller, zeroAddress, to, batch, amounts) {
		return ErrTransferToRejectedOrNonGRC1155Receiver
	}

	emit(&TransferBatchEvent{caller, zeroAddress, to, batch, amounts})

	return nil
}

// Destroys `amount` tokens of token type `id` from `from`.
func (s *basicGRC1155Token) Burn(from address, tid TokenID, amount int64) error {
	caller := runtime.OriginCaller()

	err := s.burnBatch(from, []TokenID{tid}, []int64{amount})
	if err != nil {
		return err
	}

	emit(&TransferSingleEvent{caller, from, zeroAddress, tid, amount})

	return nil
}

// Batch version of `Burn()`
func (s *basicGRC1155Token) BatchBurn(from address, batch []TokenID, amounts []int64) error {
	caller := runtime.OriginCaller()

	err := s.burnBatch(from, batch, amounts)
	if err != nil {
		return err
	}

	emit(&TransferBatchEvent{caller, from, zeroAddress, batch, amounts})

	return nil
}

/* Helper methods */

// Helper for SetApprovalForAll(): approve `operator` to operate on all of `owner` tokens
func (s *basicGRC1155Token) setApprovalForAll(owner, operator address, approved bool) error {
	if owner == operator {
		return nil
	}

	key := owner.String() + ":" + operator.String()
	if approved {
		s.operatorApprovals.Set(key, approved)
	} else {
		s.operatorApprovals.Remove(key)
	}

	emit(&ApprovalForAllEvent{owner, operator, approved})

	return nil
}

// Helper for SafeTransferFrom() and SafeBatchTransferFrom()
func (s *basicGRC1155Token) safeBatchTransferFrom(from, to address, batch []TokenID, amounts []int64) error {
	if len(batch) != len(amounts) {
		return ErrMismatchLength
	}
	if !isValidAddress(from) || !isValidAddress(to) {
		return ErrInvalidAddress
	}
	if from == to {
		return ErrCannotTransferToSelf
	}
	for _, amount := range amounts {
		if amount < 0 {
			return ErrInvalidAmount
		}
	}

	caller := runtime.OriginCaller()
	s.beforeTokenTransfer(caller, from, to, batch, amounts)

	for i := 0; i < len(batch); i++ {
		tid := batch[i]
		amount := amounts[i]
		fromBalance, err := s.BalanceOf(from, tid)
		if err != nil {
			return err
		}
		if fromBalance < amount {
			return ErrInsufficientBalance
		}
		toBalance, err := s.BalanceOf(to, tid)
		if err != nil {
			return err
		}

		fromBalance = overflow.Sub64p(fromBalance, amount)
		toBalance = overflow.Add64p(toBalance, amount)
		fromBalanceKey := string(tid) + ":" + from.String()
		toBalanceKey := string(tid) + ":" + to.String()
		s.balances.Set(fromBalanceKey, fromBalance)
		s.balances.Set(toBalanceKey, toBalance)
	}

	s.afterTokenTransfer(caller, from, to, batch, amounts)

	return nil
}

// Helper for SafeMint() and SafeBatchMint()
func (s *basicGRC1155Token) mintBatch(to address, batch []TokenID, amounts []int64) error {
	if len(batch) != len(amounts) {
		return ErrMismatchLength
	}
	if !isValidAddress(to) {
		return ErrInvalidAddress
	}
	for _, amount := range amounts {
		if amount < 0 {
			return ErrInvalidAmount
		}
	}

	caller := runtime.OriginCaller()
	s.beforeTokenTransfer(caller, zeroAddress, to, batch, amounts)

	for i := 0; i < len(batch); i++ {
		tid := batch[i]
		amount := amounts[i]
		toBalance, err := s.BalanceOf(to, tid)
		if err != nil {
			return err
		}
		toBalance = overflow.Add64p(toBalance, amount)
		toBalanceKey := string(tid) + ":" + to.String()
		s.balances.Set(toBalanceKey, toBalance)
	}

	s.afterTokenTransfer(caller, zeroAddress, to, batch, amounts)

	return nil
}

// Helper for Burn() and BurnBatch()
func (s *basicGRC1155Token) burnBatch(from address, batch []TokenID, amounts []int64) error {
	if len(batch) != len(amounts) {
		return ErrMismatchLength
	}
	if !isValidAddress(from) {
		return ErrInvalidAddress
	}
	for _, amount := range amounts {
		if amount < 0 {
			return ErrInvalidAmount
		}
	}

	caller := runtime.OriginCaller()
	s.beforeTokenTransfer(caller, from, zeroAddress, batch, amounts)

	for i := 0; i < len(batch); i++ {
		tid := batch[i]
		amount := amounts[i]
		fromBalance, err := s.BalanceOf(from, tid)
		if err != nil {
			return err
		}
		if fromBalance < amount {
			return ErrBurnAmountExceedsBalance
		}
		fromBalance = overflow.Sub64p(fromBalance, amount)
		fromBalanceKey := string(tid) + ":" + from.String()
		s.balances.Set(fromBalanceKey, fromBalance)
	}

	s.afterTokenTransfer(caller, from, zeroAddress, batch, amounts)

	return nil
}

func (s *basicGRC1155Token) setUri(newUri string) {
	s.uri = newUri
	emit(&UpdateURIEvent{newUri})
}

func (s *basicGRC1155Token) beforeTokenTransfer(operator, from, to address, batch []TokenID, amounts []int64) {
	// TODO: Implementation
}

func (s *basicGRC1155Token) afterTokenTransfer(operator, from, to address, batch []TokenID, amounts []int64) {
	// TODO: Implementation
}

func (s *basicGRC1155Token) doSafeTransferAcceptanceCheck(operator, from, to address, tid TokenID, amount int64) bool {
	// TODO: Implementation
	return true
}

func (s *basicGRC1155Token) doSafeBatchTransferAcceptanceCheck(operator, from, to address, batch []TokenID, amounts []int64) bool {
	// TODO: Implementation
	return true
}

func (s *basicGRC1155Token) RenderHome() (str string) {
	str += ufmt.Sprintf("# URI:%s\n", s.uri)

	return
}

func (mt *basicGRC1155Token) Getter() MultiTokenGetter {
	return func() IGRC1155 {
		return mt
	}
}
