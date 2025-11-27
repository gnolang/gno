package subscriptions

import (
	"chain"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/seqid"
)

type service struct {
	id            seqid.ID
	owner         *ownable.Ownable
	displayName   string
	description   string
	renewalPeriod time.Duration
	price         chain.Coins
	subscribers   *avl.Tree // key: address, value: *subscription
}

func (svc *service) balance() chain.Coins {
	balance := chain.Coins{}
	svc.subscribers.Iterate("", "", func(key string, value interface{}) bool {
		sub := value.(*subscription)
		due := sub.dueAmount(svc.renewalPeriod, svc.price)
		balance = addCoins(balance, due)
		return false
	})
	return balance
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

func mustGetServiceByName(displayName string) *service {
	svc := getServiceByName(displayName)
	if svc == nil {
		panic("service not found")
	}
	return svc
}

func getServiceByName(displayName string) *service {
	value, exists := services.Get(displayName)
	if !exists {
		return nil
	}
	return value.(*service)
}
