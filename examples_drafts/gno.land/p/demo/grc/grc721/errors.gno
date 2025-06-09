package grc721

import "errors"

var (
	ErrInvalidTokenId              = errors.New("invalid token id")
	ErrInvalidAddress              = errors.New("invalid address")
	ErrTokenIdNotHasApproved       = errors.New("token id not approved for anyone")
	ErrApprovalToCurrentOwner      = errors.New("approval to current owner")
	ErrCallerIsNotOwner            = errors.New("caller is not token owner")
	ErrCallerNotApprovedForAll     = errors.New("caller is not approved for all")
	ErrCannotTransferToSelf        = errors.New("cannot send transfer to self")
	ErrTransferFromIncorrectOwner  = errors.New("transfer from incorrect owner")
	ErrTransferToNonGRC721Receiver = errors.New("transfer to non GRC721Receiver implementer")
	ErrCallerIsNotOwnerOrApproved  = errors.New("caller is not token owner or approved")
	ErrTokenIdAlreadyExists        = errors.New("token id already exists")

	// ERC721Royalty
	ErrInvalidRoyaltyPercentage     = errors.New("invalid royalty percentage")
	ErrInvalidRoyaltyPaymentAddress = errors.New("invalid royalty paymentAddress")
	ErrCannotCalculateRoyaltyAmount = errors.New("cannot calculate royalty amount")
)
