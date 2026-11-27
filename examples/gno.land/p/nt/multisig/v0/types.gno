package multisig

// Limits keep a single safe from becoming an unbounded storage or
// render walk. Pending is the cap on proposals that are still live;
// executed history is not capped and is paged by the realm.
const (
	MaxSigners    = 32
	MaxPending    = 64
	MaxWatched    = 8
	MaxNameLen    = 64
	MaxMemoLen    = 280
	MaxPayloadLen = 512
	MaxDenomLen   = 128
	MaxTokenLen   = 256
	MaxTargetLen  = 200

	// MaxDelay is 30 days. A longer delay belongs in a new version.
	MaxDelay int64 = 30 * 24 * 60 * 60
	// MaxLifetime is how far ahead expiresAt may sit.
	MaxLifetime int64 = 366 * 24 * 60 * 60
)

// Kind is the action a proposal performs when it executes.
type Kind int

const (
	KindCoin Kind = iota + 1
	KindGRC20
	KindCall
	KindAddSigner
	KindRemoveSigner
	KindThreshold
	KindDelay
	KindPause
	KindUnpause
)

// String returns a stable kind name for events and pages.
func (k Kind) String() string {
	switch k {
	case KindCoin:
		return "coin"
	case KindGRC20:
		return "grc20"
	case KindCall:
		return "call"
	case KindAddSigner:
		return "add_signer"
	case KindRemoveSigner:
		return "remove_signer"
	case KindThreshold:
		return "threshold"
	case KindDelay:
		return "delay"
	case KindPause:
		return "pause"
	case KindUnpause:
		return "unpause"
	default:
		return "unknown"
	}
}

// Status is the stored lifecycle of a proposal. Expiry is also
// reported by DisplayStatus before the next write marks it.
type Status int

const (
	StatusPending Status = iota
	StatusExecuted
	StatusCancelled
	StatusExpired
)

// String returns a stable status name.
func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusExecuted:
		return "executed"
	case StatusCancelled:
		return "cancelled"
	case StatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// DisplayStatus reports expired for a still-stored pending proposal
// whose deadline has passed. It does not write.
func DisplayStatus(stored Status, expiresAt, now int64) Status {
	if stored == StatusPending && now >= expiresAt {
		return StatusExpired
	}
	return stored
}

// Spend reports whether kind moves value or authorizes an outside realm.
func Spend(k Kind) bool {
	return k == KindCoin || k == KindGRC20 || k == KindCall
}

// Effect is what the hosting realm must perform after a successful
// Execute. Governance kinds are already applied to the Safe.
type Effect struct {
	Kind      Kind
	SafeID    int64
	PropID    int64
	To        address
	Denom     string
	Amount    int64
	TokenKey  string
	TargetPkg string
	Payload   string
}

// ProposalView is a copy of a proposal. Approvals is a fresh slice.
type ProposalView struct {
	ID        int64
	Kind      Kind
	Proposer  address
	CreatedAt int64
	Delay     int64
	ExpiresAt int64
	ArmedAt   int64
	Memo      string
	To        address
	Denom     string
	Amount    int64
	TokenKey  string
	TargetPkg string
	Payload   string
	Signer    address
	Threshold int
	NewDelay  int64
	Approvals []address
	Status    Status
	Consumed  bool
}

type proposal struct {
	id           int64
	kind         Kind
	proposer     address
	createdAt    int64
	delay        int64
	expiresAt    int64
	armedAt      int64
	memo         string
	to           address
	denom        string
	amount       int64
	tokenKey     string
	targetPkg    string
	payload      string
	signer       address
	newThreshold int
	newDelay     int64
	approvals    []address
	status       Status
	consumed     bool
}

func (p *proposal) view() ProposalView {
	ap := make([]address, len(p.approvals))
	copy(ap, p.approvals)
	return ProposalView{
		ID:        p.id,
		Kind:      p.kind,
		Proposer:  p.proposer,
		CreatedAt: p.createdAt,
		Delay:     p.delay,
		ExpiresAt: p.expiresAt,
		ArmedAt:   p.armedAt,
		Memo:      p.memo,
		To:        p.to,
		Denom:     p.denom,
		Amount:    p.amount,
		TokenKey:  p.tokenKey,
		TargetPkg: p.targetPkg,
		Payload:   p.payload,
		Signer:    p.signer,
		Threshold: p.newThreshold,
		NewDelay:  p.newDelay,
		Approvals: ap,
		Status:    p.status,
		Consumed:  p.consumed,
	}
}

type draft struct {
	kind         Kind
	memo         string
	expiresAt    int64
	to           address
	denom        string
	amount       int64
	tokenKey     string
	targetPkg    string
	payload      string
	signer       address
	newThreshold int
	newDelay     int64
}
