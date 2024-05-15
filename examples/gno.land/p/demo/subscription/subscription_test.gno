package subscription

import (
    "std"
    "testing"
    "time"

	"gno.land/p/demo/avl"
)

// Test the initialization of a recurring subscription.
func TestNewRecurringSubscription(t *testing.T) {
    duration := time.Hour * 24 * 30 // 30 days
    amount := int64(100)
    rs := NewRecurringSubscription(duration, amount)
    if rs.duration != duration {
        t.Errorf("Expected duration %v, got %v", duration, rs.duration)
    }
    if rs.amount != amount {
        t.Errorf("Expected amount %d, got %d", amount, rs.amount)
    }
}

// Test access check for an active subscription with better error handling.
func TestRecurringSubscriptionCheckAccessActive(t *testing.T) {
    rs := NewRecurringSubscription(time.Hour*24*30, 100)
    caller := std.Address("test-address")
    // Assuming the subscription expects a previous payment, simulate it:
    rs.subs.Set(caller.String(), time.Now().Add(time.Hour*24*31)) // Set expiration in the future

    defer func() {
        if r := recover(); r != nil {
            t.Errorf("Access check failed unexpectedly: %v", r)
        }
    }()
    // Simulate sending the correct amount to access the feature
	std.TestSetOrigSend(std.Coins{{"ugnot", 100}}, nil)
    rs.CheckAccess(caller)
}


// Test access check for an expired subscription.
func TestRecurringSubscriptionCheckAccessExpired(t *testing.T) {
    rs := NewRecurringSubscription(time.Hour*24*30, 100)
    caller := std.Address("test-address")
    rs.subs.Set(caller.String(), time.Now().Add(-time.Hour)) // Set expiration in the past
    defer func() {
        if r := recover(); r == nil {
            t.Errorf("Expected access check to fail for expired subscription")
        }
    }()
    rs.CheckAccess(caller)
}

// Test processing payments correctly extends the subscription.
func TestRecurringSubscriptionProcessPayment(t *testing.T) {
    rs := NewRecurringSubscription(time.Hour*24*30, 100)
    caller := std.Address("test-address")
    // Initial payment processing
    rs.ProcessPayment(caller, 100)
    expiration := rs.GetExpiration(caller)
    if time.Now().After(expiration) {
        t.Errorf("Payment did not extend subscription as expected")
    }
}


// Test the initialization of a lifetime subscription.
func TestNewLifetimeSubscription(t *testing.T) {
    amount := int64(500)
    ls := NewLifetimeSubscription(amount)
    if ls.amount != amount {
        t.Errorf("Expected amount %d, got %d", amount, ls.amount)
    }
}

// Test processing payments for lifetime subscription.
func TestLifetimeSubscriptionProcessPayment(t *testing.T) {
    ls := NewLifetimeSubscription(500)
    caller := std.Address("test-address")

    // Test insufficient payment.
    defer func() {
        if r := recover(); r == nil {
            t.Errorf("Expected payment error for insufficient amount")
        }
    }()
    ls.ProcessPayment(caller, 400)

    // Test exact payment.
    defer func() {
        if r := recover(); r != nil {
            t.Errorf("Payment failed unexpectedly: %v", r)
        }
    }()
    ls.ProcessPayment(caller, 500)

    _, exists := ls.subs.Get(caller.String())
    if !exists {
        t.Errorf("Lifetime subscription was not recorded")
    }
}

// Test access check for lifetime subscription.
func TestLifetimeSubscriptionCheckAccess(t *testing.T) {
    ls := NewLifetimeSubscription(500)
    caller := std.Address("test-address")

    // Simulate payment and check access.
    ls.subs.Set(caller.String(), true)
    defer func() {
        if r := recover(); r != nil {
            t.Errorf("Access denied unexpectedly")
        }
    }()
    ls.CheckAccess(caller)

    // Check access without payment.
    ls.subs = avl.NewTree() // Reset the tree to simulate no payment
    defer func() {
        if r := recover(); r == nil {
            t.Errorf("Expected access to be denied for unpaid subscription")
        }
    }()
    ls.CheckAccess(caller)
}