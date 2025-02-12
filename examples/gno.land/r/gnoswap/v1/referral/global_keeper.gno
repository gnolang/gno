package referral

import "std"

// gReferralKeeper is the global instance of the referral keeper
var gReferralKeeper ReferralKeeper

func init() {
	gReferralKeeper = NewKeeper()
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
		panic(err)
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
