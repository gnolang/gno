package multisig

import (
	"chain"
	"math"
	"strconv"
	"strings"
)

// Safe is one M-of-N treasury. The hosting realm keeps the pointer
// private: the exported methods mutate, so handing the pointer to
// another realm would hand over the treasury.
type Safe struct {
	id        int64
	name      string
	creator   address
	hostPkg   string
	hostAddr  address
	subpath   string
	addr      address
	signers   []address
	threshold int
	delay     int64
	paused    bool
	nonce     int64
	props     []*proposal
	pending   []int64
	denoms    []string
	tokens    []string
}

// New validates a safe but does not bind its sub-realm address.
// signersCSV is a comma-separated list of bech32 addresses. MsgCall
// cannot pass a slice, so the factory accepts the same encoding.
func New(hostPkg string, hostAddr, creator address, name, signersCSV string, threshold int, delay int64) (*Safe, error) {
	if !strings.HasPrefix(hostPkg, "gno.land/r/") || strings.Contains(hostPkg, "#") {
		return nil, ErrHost
	}
	if !hostAddr.IsValid() || hostAddr != chain.PackageAddress(hostPkg) {
		return nil, ErrHost
	}
	if !validText(name, 1, MaxNameLen) {
		return nil, ErrName
	}
	if delay < 0 || delay > MaxDelay {
		return nil, ErrDelay
	}
	signers, err := parseSigners(signersCSV)
	if err != nil {
		return nil, err
	}
	if threshold < 1 || threshold > len(signers) {
		return nil, ErrThreshold
	}
	for _, s := range signers {
		if s == hostAddr {
			return nil, ErrSelf
		}
	}
	cp := make([]address, len(signers))
	copy(cp, signers)
	return &Safe{
		name:      name,
		creator:   creator,
		hostPkg:   hostPkg,
		hostAddr:  hostAddr,
		signers:   cp,
		threshold: threshold,
		delay:     delay,
	}, nil
}

// Bind assigns id and the sub-realm address s/<id>. It panics the
// transaction only through the returned error; it does not call Sub.
func (s *Safe) Bind(id int64) (address, error) {
	if s == nil {
		return "", ErrUnbound
	}
	if s.subpath != "" {
		return "", ErrBound
	}
	if id <= 0 {
		return "", ErrHost
	}
	sub := "s/" + strconv.FormatInt(id, 10)
	addr := chain.DerivePkgSubAddr(s.hostPkg, sub)
	for _, signer := range s.signers {
		if signer == addr {
			return "", ErrSelf
		}
	}
	s.id = id
	s.subpath = sub
	s.addr = addr
	chain.Emit(
		"SafeCreated",
		"safe", strconv.FormatInt(id, 10),
		"name", s.name,
		"creator", s.creator.String(),
		"threshold", strconv.Itoa(s.threshold),
		"signers", strconv.Itoa(len(s.signers)),
		"delay", strconv.FormatInt(s.delay, 10),
		"address", addr.String(),
	)
	return addr, nil
}

// ID is the factory id, or 0 before Bind.
func (s *Safe) ID() int64 { return s.id }

// Name is the display name.
func (s *Safe) Name() string { return s.name }

// Creator is the address that created the safe. It is not an authority.
func (s *Safe) Creator() address { return s.creator }

// HostPkg is the factory realm path.
func (s *Safe) HostPkg() string { return s.hostPkg }

// Subpath is the cur.Sub argument, empty before Bind.
func (s *Safe) Subpath() string { return s.subpath }

// Address is the treasury account. Coins and GRC20 sent here are the
// safe's funds.
func (s *Safe) Address() address { return s.addr }

// Threshold is the number of current signers required.
func (s *Safe) Threshold() int { return s.threshold }

// Delay is the timelock applied to proposals created from now on.
func (s *Safe) Delay() int64 { return s.delay }

// Paused reports whether spend proposals are blocked.
func (s *Safe) Paused() bool { return s.paused }

// Nonce counts successful executions. Proposal ids are never reused.
func (s *Safe) Nonce() int64 { return s.nonce }

// SignerCount is the current signer set size.
func (s *Safe) SignerCount() int { return len(s.signers) }

// SignerAt returns the signer at index, or an invalid address.
func (s *Safe) SignerAt(i int) address {
	if i < 0 || i >= len(s.signers) {
		return ""
	}
	return s.signers[i]
}

// IsSigner reports whether addr is a current signer.
func (s *Safe) IsSigner(addr address) bool { return s.isSigner(addr) }

// Signers returns a copy of the signer set.
func (s *Safe) Signers() []address {
	out := make([]address, len(s.signers))
	copy(out, s.signers)
	return out
}

// ProposalCount includes executed and expired proposals.
func (s *Safe) ProposalCount() int { return len(s.props) }

// PendingCount is the number of stored-pending proposals, including
// ones that are past their deadline until the next write sweeps them.
func (s *Safe) PendingCount() int { return len(s.pending) }

// View copies proposal id.
func (s *Safe) View(id int64) (ProposalView, error) {
	p, err := s.prop(id)
	if err != nil {
		return ProposalView{}, err
	}
	return p.view(), nil
}

// ApprovalCount counts approvals that still belong to current signers.
func (s *Safe) ApprovalCount(id int64) (int, error) {
	p, err := s.prop(id)
	if err != nil {
		return 0, err
	}
	return s.countApprovals(p), nil
}

// HasApproval reports whether addr approved and is still a signer.
func (s *Safe) HasApproval(id int64, addr address) (bool, error) {
	p, err := s.prop(id)
	if err != nil {
		return false, err
	}
	return s.isSigner(addr) && hasAddr(p.approvals, addr), nil
}

// Executable reports whether Execute would pass the time and vote
// checks. It does not check balances.
func (s *Safe) Executable(id, now int64) (bool, error) {
	p, err := s.prop(id)
	if err != nil {
		return false, err
	}
	if p.status != StatusPending || now >= p.expiresAt {
		return false, nil
	}
	if s.paused && Spend(p.kind) {
		return false, nil
	}
	if s.countApprovals(p) < s.threshold || p.armedAt == 0 {
		return false, nil
	}
	ready, ok := add64(p.armedAt, p.delay)
	if !ok || now < ready {
		return false, nil
	}
	return true, nil
}

// WatchedDenoms returns a copy of the display denom list.
func (s *Safe) WatchedDenoms() []string {
	out := make([]string, len(s.denoms))
	copy(out, s.denoms)
	return out
}

// WatchedTokens returns a copy of the display token-key list.
func (s *Safe) WatchedTokens() []string {
	out := make([]string, len(s.tokens))
	copy(out, s.tokens)
	return out
}

// WatchDenom adds denom to the render list. Any current signer may
// call it. It does not move funds.
func (s *Safe) WatchDenom(_ int, rlm realm, denom string) error {
	if _, err := s.auth(0, rlm); err != nil {
		return err
	}
	if !validSymbol(denom, MaxDenomLen) {
		return ErrDenom
	}
	if hasString(s.denoms, denom) {
		return nil
	}
	if len(s.denoms) >= MaxWatched {
		return ErrWatch
	}
	s.denoms = append(s.denoms, denom)
	chain.Emit("DenomWatched", "safe", s.idStr(), "denom", denom)
	return nil
}

// WatchToken adds a grc20 registry key to the render list.
func (s *Safe) WatchToken(_ int, rlm realm, tokenKey string) error {
	if _, err := s.auth(0, rlm); err != nil {
		return err
	}
	if !validSymbol(tokenKey, MaxTokenLen) {
		return ErrToken
	}
	if hasString(s.tokens, tokenKey) {
		return nil
	}
	if len(s.tokens) >= MaxWatched {
		return ErrWatch
	}
	s.tokens = append(s.tokens, tokenKey)
	chain.Emit("TokenWatched", "safe", s.idStr(), "token", tokenKey)
	return nil
}

// UnwatchDenom removes denom from the render list.
func (s *Safe) UnwatchDenom(_ int, rlm realm, denom string) error {
	if _, err := s.auth(0, rlm); err != nil {
		return err
	}
	s.denoms = dropString(s.denoms, denom)
	return nil
}

// UnwatchToken removes tokenKey from the render list.
func (s *Safe) UnwatchToken(_ int, rlm realm, tokenKey string) error {
	if _, err := s.auth(0, rlm); err != nil {
		return err
	}
	s.tokens = dropString(s.tokens, tokenKey)
	return nil
}

// ProposeCoin queues a native-coin transfer.
func (s *Safe) ProposeCoin(_ int, rlm realm, to address, denom string, amount, expiresAt int64, memo string, now int64) (int64, error) {
	return s.propose(0, rlm, draft{
		kind: KindCoin, memo: memo, expiresAt: expiresAt,
		to: to, denom: denom, amount: amount,
	}, now)
}

// ProposeGRC20 queues a GRC20 transfer of a registry key.
func (s *Safe) ProposeGRC20(_ int, rlm realm, to address, tokenKey string, amount, expiresAt int64, memo string, now int64) (int64, error) {
	return s.propose(0, rlm, draft{
		kind: KindGRC20, memo: memo, expiresAt: expiresAt,
		to: to, tokenKey: tokenKey, amount: amount,
	}, now)
}

// ProposeCall queues a one-shot authorization for targetPkg.
// Execution does not call targetPkg. The target pulls the payload
// with Consume.
func (s *Safe) ProposeCall(_ int, rlm realm, targetPkg, payload string, expiresAt int64, memo string, now int64) (int64, error) {
	return s.propose(0, rlm, draft{
		kind: KindCall, memo: memo, expiresAt: expiresAt,
		targetPkg: targetPkg, payload: payload,
	}, now)
}

// ProposeAddSigner queues a signer addition.
func (s *Safe) ProposeAddSigner(_ int, rlm realm, signer address, expiresAt int64, memo string, now int64) (int64, error) {
	return s.propose(0, rlm, draft{
		kind: KindAddSigner, memo: memo, expiresAt: expiresAt, signer: signer,
	}, now)
}

// ProposeRemoveSigner queues a signer removal.
func (s *Safe) ProposeRemoveSigner(_ int, rlm realm, signer address, expiresAt int64, memo string, now int64) (int64, error) {
	return s.propose(0, rlm, draft{
		kind: KindRemoveSigner, memo: memo, expiresAt: expiresAt, signer: signer,
	}, now)
}

// ProposeThreshold queues a threshold change.
func (s *Safe) ProposeThreshold(_ int, rlm realm, threshold int, expiresAt int64, memo string, now int64) (int64, error) {
	return s.propose(0, rlm, draft{
		kind: KindThreshold, memo: memo, expiresAt: expiresAt, newThreshold: threshold,
	}, now)
}

// ProposeDelay queues a timelock change for future proposals.
func (s *Safe) ProposeDelay(_ int, rlm realm, delay, expiresAt int64, memo string, now int64) (int64, error) {
	return s.propose(0, rlm, draft{
		kind: KindDelay, memo: memo, expiresAt: expiresAt, newDelay: delay,
	}, now)
}

// ProposePause queues a pause of spend proposals.
func (s *Safe) ProposePause(_ int, rlm realm, expiresAt int64, memo string, now int64) (int64, error) {
	return s.propose(0, rlm, draft{kind: KindPause, memo: memo, expiresAt: expiresAt}, now)
}

// ProposeUnpause queues an unpause.
func (s *Safe) ProposeUnpause(_ int, rlm realm, expiresAt int64, memo string, now int64) (int64, error) {
	return s.propose(0, rlm, draft{kind: KindUnpause, memo: memo, expiresAt: expiresAt}, now)
}

// Approve records the caller's vote. The first time the current
// signers reach the threshold, the timelock starts.
func (s *Safe) Approve(_ int, rlm realm, id, now int64) error {
	caller, err := s.auth(0, rlm)
	if err != nil {
		return err
	}
	s.sweep(now)
	p, err := s.prop(id)
	if err != nil {
		return err
	}
	if err := s.expireIfDue(p, now); err != nil {
		return err
	}
	if hasAddr(p.approvals, caller) {
		return ErrDuplicateApproval
	}
	p.approvals = append(p.approvals, caller)
	s.refreshArm(p, now)
	chain.Emit(
		"ProposalApproved",
		"safe", s.idStr(),
		"proposal", strconv.FormatInt(id, 10),
		"signer", caller.String(),
		"approvals", strconv.Itoa(s.countApprovals(p)),
		"armed_at", strconv.FormatInt(p.armedAt, 10),
	)
	return nil
}

// Revoke removes the caller's vote. Dropping under the threshold
// clears the timelock so it starts again if the proposal is re-armed.
func (s *Safe) Revoke(_ int, rlm realm, id, now int64) error {
	caller, err := s.auth(0, rlm)
	if err != nil {
		return err
	}
	s.sweep(now)
	p, err := s.prop(id)
	if err != nil {
		return err
	}
	if err := s.expireIfDue(p, now); err != nil {
		return err
	}
	if !hasAddr(p.approvals, caller) {
		return ErrNotApproved
	}
	p.approvals = dropAddr(p.approvals, caller)
	s.refreshArm(p, now)
	chain.Emit(
		"ProposalApprovalRevoked",
		"safe", s.idStr(),
		"proposal", strconv.FormatInt(id, 10),
		"signer", caller.String(),
		"approvals", strconv.Itoa(s.countApprovals(p)),
	)
	return nil
}

// Cancel drops an unarmed proposal. Any current signer may cancel
// one that has not reached the threshold. An armed proposal stays
// until signers revoke below the threshold, or it expires.
func (s *Safe) Cancel(_ int, rlm realm, id, now int64) error {
	if _, err := s.auth(0, rlm); err != nil {
		return err
	}
	s.sweep(now)
	p, err := s.prop(id)
	if err != nil {
		return err
	}
	if err := s.expireIfDue(p, now); err != nil {
		return err
	}
	if s.countApprovals(p) >= s.threshold && p.armedAt != 0 {
		return ErrArmed
	}
	p.status = StatusCancelled
	s.dropPending(id)
	chain.Emit(
		"ProposalCancelled",
		"safe", s.idStr(),
		"proposal", strconv.FormatInt(id, 10),
	)
	return nil
}

// Execute marks a ready proposal executed and returns the effect the
// realm must apply. afford, when set, runs after the proposal is known
// to be ready and before it is marked, so a rejected spend leaves the
// proposal pending.
func (s *Safe) Execute(_ int, rlm realm, id, now int64) (Effect, error) {
	return s.execute(0, rlm, id, now, nil)
}

// ExecuteChecked is Execute with an afford check. The realm uses it to
// reject an underfunded transfer without consuming the proposal.
func (s *Safe) ExecuteChecked(_ int, rlm realm, id, now int64, afford func(Effect) error) (Effect, error) {
	return s.execute(0, rlm, id, now, afford)
}

func (s *Safe) execute(_ int, rlm realm, id, now int64, afford func(Effect) error) (Effect, error) {
	caller, err := s.auth(0, rlm)
	if err != nil {
		return Effect{}, err
	}
	if s.subpath == "" {
		return Effect{}, ErrUnbound
	}
	s.sweep(now)
	p, err := s.prop(id)
	if err != nil {
		return Effect{}, err
	}
	if err := s.expireIfDue(p, now); err != nil {
		return Effect{}, err
	}
	if s.paused && Spend(p.kind) {
		return Effect{}, ErrPaused
	}
	if s.countApprovals(p) < s.threshold {
		return Effect{}, ErrNotReady
	}
	if p.armedAt == 0 {
		return Effect{}, ErrTimelock
	}
	ready, ok := add64(p.armedAt, p.delay)
	if !ok || now < ready {
		return Effect{}, ErrTimelock
	}
	if err := s.applyGovernance(p); err != nil {
		return Effect{}, err
	}
	eff := Effect{
		Kind: p.kind, SafeID: s.id, PropID: p.id,
		To: p.to, Denom: p.denom, Amount: p.amount,
		TokenKey: p.tokenKey, TargetPkg: p.targetPkg, Payload: p.payload,
	}
	if Spend(p.kind) && afford != nil {
		if err := afford(eff); err != nil {
			return Effect{}, err
		}
	}
	p.status = StatusExecuted
	s.dropPending(id)
	s.nonce++
	s.refreshAll(now)
	chain.Emit(
		"ProposalExecuted",
		"safe", s.idStr(),
		"proposal", strconv.FormatInt(id, 10),
		"kind", p.kind.String(),
		"executor", caller.String(),
		"nonce", strconv.FormatInt(s.nonce, 10),
	)
	return eff, nil
}

// Consume lets targetPkg pull a call payload once. rlm.Previous()
// must be that package. The safe does not call the target.
func (s *Safe) Consume(_ int, rlm realm, id int64) (string, error) {
	if !rlm.IsCurrent() {
		return "", ErrStaleRealm
	}
	if s.paused {
		return "", ErrPaused
	}
	p, err := s.prop(id)
	if err != nil {
		return "", err
	}
	if p.kind != KindCall || p.status != StatusExecuted {
		return "", ErrNotConsumable
	}
	if p.consumed {
		return "", ErrConsumed
	}
	if rlm.Previous().PkgPath() != p.targetPkg {
		return "", ErrUnauthorized
	}
	p.consumed = true
	chain.Emit(
		"AuthorizationConsumed",
		"safe", s.idStr(),
		"proposal", strconv.FormatInt(id, 10),
		"target", p.targetPkg,
	)
	return p.payload, nil
}

func (s *Safe) propose(_ int, rlm realm, d draft, now int64) (int64, error) {
	caller, err := s.auth(0, rlm)
	if err != nil {
		return 0, err
	}
	if s.subpath == "" {
		return 0, ErrUnbound
	}
	s.sweep(now)
	if len(s.pending) >= MaxPending {
		return 0, ErrTooManyPending
	}
	if err := s.validateDraft(d, now); err != nil {
		return 0, err
	}
	id := int64(len(s.props)) + 1
	p := &proposal{
		id: id, kind: d.kind, proposer: caller, createdAt: now,
		delay: s.delay, expiresAt: d.expiresAt, memo: d.memo,
		to: d.to, denom: d.denom, amount: d.amount, tokenKey: d.tokenKey,
		targetPkg: d.targetPkg, payload: d.payload, signer: d.signer,
		newThreshold: d.newThreshold, newDelay: d.newDelay,
		status: StatusPending,
	}
	s.props = append(s.props, p)
	s.pending = append(s.pending, id)
	chain.Emit(
		"ProposalCreated",
		"safe", s.idStr(),
		"proposal", strconv.FormatInt(id, 10),
		"kind", d.kind.String(),
		"proposer", caller.String(),
		"expires_at", strconv.FormatInt(d.expiresAt, 10),
		"delay", strconv.FormatInt(s.delay, 10),
	)
	return id, nil
}

func (s *Safe) validateDraft(d draft, now int64) error {
	if !validText(d.memo, 0, MaxMemoLen) {
		return ErrMemo
	}
	ready, ok := add64(now, s.delay)
	if !ok || d.expiresAt <= ready {
		return ErrExpiry
	}
	horizon, ok := add64(now, MaxLifetime)
	if !ok || d.expiresAt > horizon {
		return ErrExpiry
	}
	switch d.kind {
	case KindCoin:
		if err := s.validPay(d.to, d.amount); err != nil {
			return err
		}
		if !validSymbol(d.denom, MaxDenomLen) {
			return ErrDenom
		}
	case KindGRC20:
		if err := s.validPay(d.to, d.amount); err != nil {
			return err
		}
		if !validSymbol(d.tokenKey, MaxTokenLen) {
			return ErrToken
		}
	case KindCall:
		if !validTarget(d.targetPkg) || d.targetPkg == s.hostPkg {
			return ErrTarget
		}
		if !validText(d.payload, 0, MaxPayloadLen) {
			return ErrPayload
		}
	case KindAddSigner:
		if err := s.validNewSigner(d.signer); err != nil {
			return err
		}
	case KindRemoveSigner:
		if !s.isSigner(d.signer) {
			return ErrUnknownSigner
		}
		if len(s.signers) == 1 || s.threshold > len(s.signers)-1 {
			return ErrLastSigner
		}
	case KindThreshold:
		if d.newThreshold < 1 || d.newThreshold > len(s.signers) {
			return ErrThreshold
		}
	case KindDelay:
		if d.newDelay < 0 || d.newDelay > MaxDelay {
			return ErrDelay
		}
	case KindPause:
		if s.paused {
			return ErrAlreadyPaused
		}
	case KindUnpause:
		if !s.paused {
			return ErrNotPaused
		}
	default:
		return ErrKind
	}
	return nil
}

func (s *Safe) applyGovernance(p *proposal) error {
	switch p.kind {
	case KindCoin, KindGRC20, KindCall:
		return nil
	case KindAddSigner:
		if err := s.validNewSigner(p.signer); err != nil {
			return err
		}
		s.signers = append(s.signers, p.signer)
		chain.Emit("SignerAdded", "safe", s.idStr(), "signer", p.signer.String())
	case KindRemoveSigner:
		if !s.isSigner(p.signer) {
			return ErrUnknownSigner
		}
		if len(s.signers) == 1 || s.threshold > len(s.signers)-1 {
			return ErrLastSigner
		}
		s.signers = dropAddr(s.signers, p.signer)
		s.stripApproval(p.signer)
		chain.Emit("SignerRemoved", "safe", s.idStr(), "signer", p.signer.String())
	case KindThreshold:
		if p.newThreshold < 1 || p.newThreshold > len(s.signers) {
			return ErrThreshold
		}
		s.threshold = p.newThreshold
		chain.Emit("ThresholdChanged", "safe", s.idStr(), "threshold", strconv.Itoa(s.threshold))
	case KindDelay:
		if p.newDelay < 0 || p.newDelay > MaxDelay {
			return ErrDelay
		}
		s.delay = p.newDelay
		chain.Emit("DelayChanged", "safe", s.idStr(), "delay", strconv.FormatInt(s.delay, 10))
	case KindPause:
		if s.paused {
			return ErrAlreadyPaused
		}
		s.paused = true
		chain.Emit("SafePaused", "safe", s.idStr())
	case KindUnpause:
		if !s.paused {
			return ErrNotPaused
		}
		s.paused = false
		chain.Emit("SafeUnpaused", "safe", s.idStr())
	default:
		return ErrKind
	}
	return nil
}

func (s *Safe) validPay(to address, amount int64) error {
	if !to.IsValid() {
		return ErrInvalidAddress
	}
	if to == s.addr || to == s.hostAddr {
		return ErrSelf
	}
	if amount <= 0 {
		return ErrAmount
	}
	return nil
}

func (s *Safe) validNewSigner(signer address) error {
	if !signer.IsValid() {
		return ErrInvalidAddress
	}
	if signer == s.hostAddr || signer == s.addr {
		return ErrSelf
	}
	if s.isSigner(signer) {
		return ErrDuplicateSigner
	}
	if len(s.signers) >= MaxSigners {
		return ErrSignerCount
	}
	return nil
}

func (s *Safe) auth(_ int, rlm realm) (address, error) {
	if s == nil {
		return "", ErrUnbound
	}
	if !rlm.IsCurrent() {
		return "", ErrStaleRealm
	}
	addr := rlm.Previous().Address()
	if !s.isSigner(addr) {
		return "", ErrNotSigner
	}
	return addr, nil
}

func (s *Safe) prop(id int64) (*proposal, error) {
	if id <= 0 || int(id) > len(s.props) {
		return nil, ErrUnknownProposal
	}
	p := s.props[id-1]
	if p == nil || p.id != id {
		return nil, ErrUnknownProposal
	}
	return p, nil
}

func (s *Safe) expireIfDue(p *proposal, now int64) error {
	if p.status == StatusExpired || (p.status == StatusPending && now >= p.expiresAt) {
		if p.status == StatusPending {
			p.status = StatusExpired
			s.dropPending(p.id)
			chain.Emit(
				"ProposalExpired",
				"safe", s.idStr(),
				"proposal", strconv.FormatInt(p.id, 10),
			)
		}
		return ErrExpired
	}
	if p.status != StatusPending {
		return ErrNotPending
	}
	return nil
}

func (s *Safe) sweep(now int64) {
	ids := make([]int64, len(s.pending))
	copy(ids, s.pending)
	s.pending = nil
	for _, id := range ids {
		p := s.props[id-1]
		if p.status == StatusPending && now >= p.expiresAt {
			p.status = StatusExpired
			chain.Emit(
				"ProposalExpired",
				"safe", s.idStr(),
				"proposal", strconv.FormatInt(id, 10),
			)
			continue
		}
		if p.status == StatusPending {
			s.pending = append(s.pending, id)
		}
	}
}

func (s *Safe) dropPending(id int64) {
	next := make([]int64, 0, len(s.pending))
	for _, x := range s.pending {
		if x != id {
			next = append(next, x)
		}
	}
	s.pending = next
}

func (s *Safe) refreshArm(p *proposal, now int64) {
	if p.status != StatusPending {
		return
	}
	if s.countApprovals(p) >= s.threshold {
		if p.armedAt == 0 {
			p.armedAt = now
		}
		return
	}
	p.armedAt = 0
}

func (s *Safe) refreshAll(now int64) {
	for _, id := range s.pending {
		s.refreshArm(s.props[id-1], now)
	}
}

func (s *Safe) countApprovals(p *proposal) int {
	n := 0
	for _, a := range p.approvals {
		if s.isSigner(a) {
			n++
		}
	}
	return n
}

func (s *Safe) stripApproval(signer address) {
	for _, p := range s.props {
		if p.status != StatusPending {
			continue
		}
		p.approvals = dropAddr(p.approvals, signer)
	}
}

func (s *Safe) isSigner(addr address) bool { return hasAddr(s.signers, addr) }

func (s *Safe) idStr() string { return strconv.FormatInt(s.id, 10) }

func parseSigners(csv string) ([]address, error) {
	parts := strings.Split(csv, ",")
	if len(parts) == 0 {
		return nil, ErrSignerCount
	}
	out := make([]address, 0, len(parts))
	for _, part := range parts {
		text := strings.TrimSpace(part)
		if text == "" {
			return nil, ErrInvalidAddress
		}
		addr := address(text)
		if !addr.IsValid() {
			return nil, ErrInvalidAddress
		}
		if hasAddr(out, addr) {
			return nil, ErrDuplicateSigner
		}
		out = append(out, addr)
	}
	if len(out) < 1 || len(out) > MaxSigners {
		return nil, ErrSignerCount
	}
	return out, nil
}

func validText(s string, min, max int) bool {
	if len(s) < min || len(s) > max {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c == 0x7f {
			return false
		}
	}
	return true
}

func validSymbol(s string, max int) bool {
	if s == "" || len(s) > max {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '/' || c == ':' || c == '.' || c == '_' || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

func validTarget(s string) bool {
	if len(s) < len("gno.land/r/a") || len(s) > MaxTargetLen {
		return false
	}
	if !strings.HasPrefix(s, "gno.land/r/") {
		return false
	}
	hashes := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '#' {
			hashes++
			if hashes > 1 {
				return false
			}
			continue
		}
		ok := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '/' || c == '_' || c == '.' || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

func hasAddr(list []address, addr address) bool {
	for _, a := range list {
		if a == addr {
			return true
		}
	}
	return false
}

func dropAddr(list []address, addr address) []address {
	next := make([]address, 0, len(list))
	for _, a := range list {
		if a != addr {
			next = append(next, a)
		}
	}
	return next
}

func hasString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func dropString(list []string, s string) []string {
	next := make([]string, 0, len(list))
	for _, item := range list {
		if item != s {
			next = append(next, item)
		}
	}
	return next
}

func add64(a, b int64) (int64, bool) {
	if a < 0 || b < 0 {
		return 0, false
	}
	if a > math.MaxInt64-b {
		return 0, false
	}
	return a + b, true
}
