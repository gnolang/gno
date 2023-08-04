package subscription

import "testing"

func TestPackage(t *testing.T) {
	// Example: Duration of 30 days and amount of 1000 ugno
	sub := subscription.NewRecurringSubscription(time.Hour*24*30, 1000)
	// simulate send with std.TestSetOrigSend
	// simulate orig caller with std.TestSetOrigCaller
	sub.CheckOrigCaller()
}
