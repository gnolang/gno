package referral

import "std"

// gReferralKeeper is the global instance of the referral keeper
var gReferralKeeper ReferralKeeper

const (
	EventReferralInvalid = "ReferralInvalid"
	EventRegisterFailed  = "ReferralRegistrationFailed"
	EventRegisterSuccess = "ReferralRegistrationSuccess"
)

func init() {
	if gReferralKeeper == nil {
		gReferralKeeper = NewKeeper()
	}
}

// GetKeeper returns the global instance of the referral keeper
//
// Example:
//
//	// In other packages:
//	keeper := referral.GetKeeper()
//	keeper.register(addr, refAddr)
func GetKeeper() ReferralKeeper {
	return gReferralKeeper
}

func GetReferral(addr string) string {
	referral, err := gReferralKeeper.get(std.Address(addr))
	if err != nil {
		return ""
	}
	return referral.String()
}

func HasReferral(addr string) bool {
	referral, err := gReferralKeeper.get(std.Address(addr))
	if err != nil {
		return false
	}
	return referral != zeroAddress
}

func IsEmpty() bool {
	return gReferralKeeper.isEmpty()
}

// TryRegister attempts to register a new referral for an address.
func TryRegister(cur realm, addr std.Address, referral string) bool {
	refAddr := std.Address(referral)
	r := NewReferral()
	err := r.Register(addr, refAddr)
	if err != nil {
		std.Emit(
			EventRegisterFailed,
			"address", addr.String(),
			"error", err.Error(),
		)
		return false
	}

	// update global keeper
	gReferralKeeper = r.keeper

	std.Emit(
		EventRegisterSuccess,
		"address", addr.String(),
		"referral", referral,
	)
	return true
}
