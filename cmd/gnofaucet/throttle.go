package main

import (
	"fmt"
	"net"
	"time"

	"github.com/gnolang/gno/pkgs/service"
)

type SubnetThrottler struct {
	service.BaseService
	ticker   *time.Ticker
	subnets3 [2 << (8 * 3)]uint8
	// subnets2 [2 << (8 * 2)]uint8
	// subnets1 [2 << (8 * 1)]uint8
}

func NewSubnetThrottler() *SubnetThrottler {
	st := &SubnetThrottler{}
	// st.ticker = time.NewTicker(time.Second)
	st.ticker = time.NewTicker(time.Minute)
	st.BaseService = *service.NewBaseService(nil, "SubnetThrottler", st)
	return st
}

func (st *SubnetThrottler) OnStart() error {
	st.BaseService.OnStart()
	go st.routineTimer()
	return nil
}

func (st *SubnetThrottler) routineTimer() {
	for {
		select {
		case <-st.Quit():
			return
		case <-st.ticker.C:
			// run something every time interval.
			for i := range st.subnets3 {
				st.subnets3[i] /= 2
			}
		}
	}
}

func (st *SubnetThrottler) Request(ip net.IP) bool {
	ip = ip.To4()
	if len(ip) != 4 {
		fmt.Println("not 4", len(ip), ip)
		return false
	}
	bucket3 := int(ip[0])*256*256 +
		int(ip[1])*256 +
		int(ip[2])
	v := st.subnets3[bucket3]
	if v > 5 {
		fmt.Println("> 5")
		return false
	} else {
		st.subnets3[bucket3] += 1
	}
	return true
}
