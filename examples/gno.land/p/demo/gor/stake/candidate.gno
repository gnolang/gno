package stake

import (
	"errors"
	"gno.land/p/demo/avl"
	std "std"
)

// assume an mechanism that paticipants in network could apply for a delegate
func (s *Stake) apply(addr std.Address) {
	// TODO: a slice of applications, to be aprroved
}

func (s *Stake) approve(addr std.Address) {
	d := &Delegate{
		addr: addr,
	}
	s.delegates.Set(addr.String(), d)
}

func (s *Stake) GetDelegateByAddr(addr string) (*Delegate, error) {
	d, found := s.delegates.Get(addr)
	if !found {
		return &Delegate{}, errors.New("delegate not exist")
	}
	return d.(*Delegate), nil
}

func (s *Stake) IterateDelegates(cb func(n *avl.Node) bool) {
	s.delegates.Iterate("", "", cb)
}

func (s *Stake) RemoveDelegate(delegateAddr string) bool {
	_, ok := s.delegates.Remove(delegateAddr)
	return ok
}

// TODO : mock a reward to delegate, that will chagne tokens/shares factor
