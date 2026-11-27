package modelledger

type Ledger struct{ bal map[address]int64 }

func New() *Ledger { return &Ledger{bal: map[address]int64{}} }

func (l *Ledger) Move(from, to address, n int64) { l.bal[from] -= n; l.bal[to] += n }
func (l *Ledger) Balance(a address) int64        { return l.bal[a] }
