package governance

import (
	"errors"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gnoswap/v1/common"

	en "gno.land/r/gnoswap/v1/emission"
)

// Governance can execute multiple messages in a single proposal
// each message is a string with the following format:
// <pkgPath>*EXE*<function>*EXE*<params>
// To execute a message, we need to parse the message and call the corresponding function
// with the given parameters
const (
	EXECUTE_SEPARATOR = "*EXE*"
)

// Function signature for different parameter handlers
type ParameterHandler func([]string) error

// Registry for parameter handlers
type ParameterRegistry struct {
	handlers *avl.Tree
}

func NewParameterRegistry() *ParameterRegistry {
	return &ParameterRegistry{
		handlers: avl.NewTree(),
	}
}

func (r *ParameterRegistry) Register(
	pkgPath, function string,
	hdlr ParameterHandler,
) {
	key := makeHandlerKey(pkgPath, function)
	r.handlers.Set(key, hdlr)
}

func (r *ParameterRegistry) Handler(pkgPath, function string) (ParameterHandler, error) {
	key := makeHandlerKey(pkgPath, function)
	hdlr, exists := r.handlers.Get(key)
	if !exists {
		return nil, ufmt.Errorf("handler not found for %s", key)
	}
	return hdlr.(ParameterHandler), nil
}

func (r *ParameterRegistry) Get(pkgPath, function string) (ParameterHandler, error) {
	key := makeHandlerKey(pkgPath, function)
	hdlr, exists := r.handlers.Get(key)
	if !exists {
		return nil, ufmt.Errorf("handler not found for %s", key)
	}
	return hdlr.(ParameterHandler), nil
}

func makeHandlerKey(pkgPath, function string) string {
	return ufmt.Sprintf("%s:%s", pkgPath, function)
}

///////////////////// EXECUTION /////////////////////
// region: Execute

type ExecutionContext struct {
	ProposalId  uint64
	Now         uint64
	Config      *Config
	Proposal    *ProposalInfo
	WindowStart uint64
	WindowEnd   uint64
}

func (e *ExecutionContext) String() string {
	return ufmt.Sprintf(
		"ProposalId: %d, Now: %d, Config: %v, Proposal: %v, WindowStart: %d, WindowEnd: %d",
		e.ProposalId, e.Now, e.Config, e.Proposal, e.WindowStart, e.WindowEnd,
	)
}

func Execute(proposalId uint64) error {
	ctx, err := prepareExecution(proposalId)
	if err != nil {
		panic(err)
	}

	if err := validateVotes(ctx.Proposal); err != nil {
		panic(err)
	}

	if err := validateCommunityPoolToken(ctx.Proposal); err != nil {
		panic(err)
	}

	registry := createParameterHandlers()
	if err := executeProposal(ctx, registry); err != nil {
		return err
	}

	updateProposalState(ctx)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"Execute",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"proposalId", formatUint(proposalId),
	)

	return nil
}

func executeProposal(ctx *ExecutionContext, registry *ParameterRegistry) error {
	switch ctx.Proposal.ProposalType {
	case ParameterChange, CommunityPoolSpend:
		return executeParameterChange(ctx.Proposal.Execution.Msgs, registry)
	default:
		return errUnsupportedProposalType
	}
}

func executeParameterChange(msgs []string, registry *ParameterRegistry) error {
	for _, msg := range msgs {
		pkgPath, function, params, err := parseMessage(msg)
		if err != nil {
			return err
		}

		handler, err := registry.Handler(pkgPath, function)
		if err != nil {
			return err
		}

		if err := handler(params); err != nil {
			return err
		}
	}

	return nil
}

func parseMessage(msg string) (pkgPath string, function string, params []string, err error) {
	parts := strings.Split(msg, EXECUTE_SEPARATOR)
	if len(parts) != 3 {
		return "", "", nil, errInvalidMessageFormat
	}

	return parts[0], parts[1], strings.Split(parts[2], ","), nil
}

///////////////////// VALIDATION /////////////////////

type ExecutionValidator struct {
	isTextProposal    bool
	isAlreadyExecuted bool
	isAlreadyCanceled bool
	isAlreadyRejected bool
	hasPassed         bool
}

func prepareExecution(proposalId uint64) (*ExecutionContext, error) {
	validateInitialState()

	proposal, err := getProposal(proposalId)
	if err != nil {
		return nil, err
	}

	validator := validateProposalState(proposal)
	if err := checkProposalValidation(validator); err != nil {
		return nil, err
	}

	ctx, err := createExecutionContext(proposalId, proposal)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func validateInitialState() {
	common.IsHalted()

	en.MintAndDistributeGns()
	updateProposalsState()
}

func getProposal(proposalId uint64) (*ProposalInfo, error) {
	result, exists := proposals.Get(strconv.Itoa(int(proposalId)))
	if !exists {
		return nil, ufmt.Errorf("proposal %d not found", proposalId)
	}
	proposal, exists := result.(ProposalInfo)
	if !exists {
		return nil, ufmt.Errorf("proposal %d not found", proposalId)
	}
	return &proposal, nil
}

func validateProposalState(proposal *ProposalInfo) ExecutionValidator {
	return ExecutionValidator{
		isTextProposal:    proposal.ProposalType == Text,
		isAlreadyExecuted: proposal.State.Executed,
		isAlreadyCanceled: proposal.State.Canceled,
		isAlreadyRejected: proposal.State.Rejected,
		hasPassed:         proposal.State.Passed,
	}
}

func checkProposalValidation(v ExecutionValidator) error {
	if v.isTextProposal {
		return errTextProposalNotExecutable
	}

	if v.isAlreadyExecuted || v.isAlreadyCanceled || v.isAlreadyRejected {
		return errors.New("proposal already executed, canceled, or rejected")
	}

	if !v.hasPassed {
		return errProposalNotPassed
	}

	return nil
}

func createExecutionContext(proposalId uint64, proposal *ProposalInfo) (*ExecutionContext, error) {
	now := uint64(time.Now().Unix())
	config := GetConfigVersion(proposal.ConfigVersion)

	votingEnd := calculateVotingEnd(proposal, &config)
	windowStart := calculateWindowStart(votingEnd, &config)
	windowEnd := calculateWindowEnd(windowStart, &config)

	if err := validateExecutionWindow(now, windowStart, windowEnd); err != nil {
		return nil, err
	}

	return &ExecutionContext{
		ProposalId:  proposalId,
		Now:         now,
		Config:      &config,
		Proposal:    proposal,
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
	}, nil
}

func calculateVotingEnd(proposal *ProposalInfo, config *Config) uint64 {
	return proposal.State.CreatedAt +
		config.VotingStartDelay +
		config.VotingPeriod
}

func calculateWindowStart(votingEnd uint64, config *Config) uint64 {
	return votingEnd + config.ExecutionDelay
}

func calculateWindowEnd(windowStart uint64, config *Config) uint64 {
	return windowStart + config.ExecutionWindow
}

func validateExecutionWindow(now, windowStart, windowEnd uint64) error {
	if now < windowStart {
		return ufmt.Errorf("execution window not started (now(%d) < windowStart(%d))", now, windowStart)
	}

	if now >= windowEnd {
		return ufmt.Errorf("execution window over (now(%d) >= windowEnd(%d))", now, windowEnd)
	}

	return nil
}

func validateVotes(pp *ProposalInfo) error {
	yea := pp.Yea.Uint64()
	nea := pp.Nay.Uint64()
	quorum := pp.QuorumAmount

	if yea < quorum {
		return ufmt.Errorf("quorum not met (yes(%d) < quorum(%d))", yea, quorum)
	}

	if yea < nea {
		return ufmt.Errorf("no majority (yes(%d) < no(%d))", yea, nea)
	}

	return nil
}

func updateProposalState(ctx *ExecutionContext) {
	ctx.Proposal.State.Executed = true
	ctx.Proposal.State.ExecutedAt = ctx.Now
	ctx.Proposal.State.Upcoming = false
	ctx.Proposal.State.Active = false
	proposals.Set(strconv.Itoa(int(ctx.ProposalId)), *ctx.Proposal)
}

func validateCommunityPoolToken(pp *ProposalInfo) error {
	if pp.ProposalType != CommunityPoolSpend {
		return nil
	}

	common.MustRegistered(pp.CommunityPoolSpend.TokenPath)

	return nil
}

func hasDesiredParams(params []string, expected int) error {
	if len(params) != expected {
		return ufmt.Errorf("invalid parameters for %s. expected %d but got %d", params, expected, len(params))
	}
	return nil
}
