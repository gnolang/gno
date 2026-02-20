package subscriptions

import (
	"chain"
	"strings"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/seqid"
	"gno.land/r/samcrew/normalizedcoins"
)

const serviceKeySeparator = ":"

func serviceKey(owner address, displayName string) string {
	return owner.String() + serviceKeySeparator + displayName
}

func parseServiceKey(key string) (owner address, displayName string) {
	idx := strings.Index(key, serviceKeySeparator)
	if idx == -1 {
		return "", ""
	}
	return address(key[:idx]), key[idx+1:]
}

type service struct {
	id            seqid.ID
	owner         *ownable.Ownable
	displayName   string
	description   string
	renewalPeriod time.Duration
	price         chain.Coins
	subscribers   *avl.Tree // key: address, value: *subscription
}

// balance calculates the total amount owed by all subscribers.
// WARNING: This iterates through all subscribers and is gas-expensive.
// If charge is true, it will also charge all subscribers for their due amounts.
func (svc *service) balance(charge bool) chain.Coins {
	total := chain.Coins{}
	svc.subscribers.Iterate("", "", func(key string, value interface{}) bool {
		sub := value.(*subscription)
		due, _ := sub.dueAmount(charge, svc.renewalPeriod, svc.price)
		total = normalizedcoins.AddCoins(total, due)
		return false
	})
	return total
}

func (svc *service) subscriberCount() int {
	count := 0
	svc.subscribers.Iterate("", "", func(key string, value interface{}) bool {
		s := value.(*subscription)
		if s.isActive(svc.renewalPeriod, svc.price) {
			count++
		}
		return false
	})
	return count
}

func mustGetService(key string) *service {
	svc := getService(key)
	if svc == nil {
		panic("service not found")
	}
	return svc
}

func getService(key string) *service {
	value, exists := services.Get(key)
	if !exists {
		return nil
	}
	return value.(*service)
}
