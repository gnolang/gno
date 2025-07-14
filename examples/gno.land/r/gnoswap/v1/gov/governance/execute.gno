package governance

import (
	"errors"
	"std"

	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gnoswap/v1/common"

	en "gno.land/r/gnoswap/v1/emission"

	"gno.land/r/gnoswap/v1/halt"
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
type ParameterHandler interface {
	// Execute processes the parameters
	Execute(params []string) error
	// GetState returns the current state (optional)
	GetState() interface{}
	// RestoreState restores the given state (optional)
	RestoreState(state interface{}) error
}

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
	paramHdlr, ok := hdlr.(ParameterHandler)
	if !ok {
		return nil, ufmt.Errorf("failed to cast handler %s to ParameterHandler", key)
	}
	return paramHdlr, nil
}

func makeHandlerKey(pkgPath, function string) string {
	return ufmt.Sprintf("%s:%s", pkgPath, function)
}

///////////////////// EXECUTION /////////////////////

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
		`
ProposalId: %d,
Now: %d,
Config: %v,
Proposal: %v,
WindowStart: %d,
WindowEnd: %d
`,
		e.ProposalId, e.Now, e.Config, e.Proposal, e.WindowStart, e.WindowEnd,
	)
}

func Execute(cur realm, proposalId uint64) error {
	ctx, err := prepareExecution(cur, proposalId)
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

	// store the original state
	originalState := *ctx.Proposal

	// try to execute
	if err := executeProposal(cur, ctx, registry); err != nil {
		// rollback the original state
		*ctx.Proposal = originalState
		return err
	}

	// update the proposal state when success
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

func executeProposal(cur realm, ctx *ExecutionContext, registry *ParameterRegistry) error {
	switch ctx.Proposal.ProposalType {
	case ParameterChange, CommunityPoolSpend:
		if err := executeParameterChange(cur, ctx.Proposal.Execution.Msgs, registry); err != nil {
			return err
		}
		return nil
	case RoleChange:
		return dispatchRoleChange(cur, ctx.Proposal)
	default:
		return errUnsupportedProposalType
	}
}

func executeParameterChange(cur realm, msgs []string, registry *ParameterRegistry) error {
	stateManager := newStateManager()

	for _, msg := range msgs {
		pkgPath, function, params, err := parseMessage(msg)
		if err != nil {
			return err
		}

		handler, err := registry.Handler(pkgPath, function)
		if err != nil {
			return err
		}

		// Save state before execution
		if err := stateManager.SaveState(pkgPath, function, handler); err != nil {
			return err
		}

		// Execute the message
		if err := handler.Execute(params); err != nil {
			// restore the original state when failed
			if rollbackErr := stateManager.RestoreStates(); rollbackErr != nil {
				panic(rollbackErr)
			}
			return ufmt.Errorf("failed to execute message: %s, error: %v", msg, err)
		}
	}

	return nil
}

// dispatchRoleChange function calls different functions depending on the type of role change
func dispatchRoleChange(cur realm, proposal *ProposalInfo) error {
	roleInfo := proposal.RoleChange

	switch roleInfo.Action {
	case AddRole:
		return executeAddRole(cur, roleInfo)
	case RemoveRole:
		return executeRemoveRole(cur, roleInfo)
	case AddPermission:
		return executeAddPermission(cur, roleInfo)
	case RemovePermission:
		return executeRemovePermission(cur, roleInfo)
	case UpdatePermission:
		return executeUpdatePermission(cur, roleInfo)
	default:
		return ufmt.Errorf("unsupported role change action: %s", roleInfo.Action.String())
	}
}

// makeExecuteMessage creates a message to execute a function
// message format: <pkgPath>*EXE*<function>*EXE*<params>
func makeExecuteMessage(pkgPath, function string, params []string) string {
	messageParams := make([]string, 0)

	messageParams = append(messageParams, pkgPath)
	messageParams = append(messageParams, function)
	messageParams = append(messageParams, strings.Join(params, ","))

	return strings.Join(messageParams, EXECUTE_SEPARATOR)
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

func prepareExecution(cur realm, proposalId uint64) (*ExecutionContext, error) {
	validateInitialState(cur)

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

func validateInitialState(cur realm) {
	currentLevel := halt.GetCurrentHaltLevel()
	if currentLevel != halt.LvMainnetSafeMode {
		// Check if withdrawals are specifically halted, not just if the system is halted
		if err := halt.IsHalted(halt.OpTypeWithdraw); err != nil {
			panic(err.Error())
		}

		if halt.IsContractHalted(halt.OpTypeGovernance) {
			panic("gov/governance contract operations are currently disabled")
		}
	}

	en.MintAndDistributeGns(cross)

	// Update proposal states before execution
	transition := UpdateProposalStateTransition()
	if err := applyTransition(transition); err != nil {
		panic(err)
	}
}

func getProposal(proposalId uint64) (*ProposalInfo, error) {
	proposal, exists := governanceState.GetProposal(proposalId)
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
	// Use state transition to mark proposal as executed
	transition := ExecuteProposalTransition(ctx.ProposalId)
	if err := applyTransition(transition); err != nil {
		panic(err)
	}
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

// StatefulHandler is an interface that can save and restore state
type StatefulHandler interface {
	// GetState returns the current state
	GetState() interface{}
	// RestoreState restores the given state
	RestoreState(state interface{}) error
}

// handlerState stores the state information of the handler
type handlerState struct {
	pkgPath  string
	function string
	state    any
	handler  ParameterHandler
}

// stateManager manages the state
type stateManager struct {
	states map[string]*handlerState
}

func newStateManager() *stateManager {
	return &stateManager{
		states: make(map[string]*handlerState),
	}
}

func (sm *stateManager) SaveState(pkgPath, function string, handler ParameterHandler) error {
	key := makeHandlerKey(pkgPath, function)

	// check if the handler implements the StatefulHandler interface
	if sh, ok := handler.(StatefulHandler); ok {
		sm.states[key] = &handlerState{
			pkgPath:  pkgPath,
			function: function,
			state:    sh.GetState(),
			handler:  handler,
		}
		return nil
	}

	// handlers that do not implement the StatefulHandler interface do not need to save state
	return nil
}

func (sm *stateManager) RestoreStates() error {
	for _, state := range sm.states {
		if sh, ok := state.handler.(StatefulHandler); ok {
			if err := sh.RestoreState(state.state); err != nil {
				return ufmt.Errorf("failed to restore state for %s: %v",
					makeHandlerKey(state.pkgPath, state.function), err)
			}
		}
	}
	return nil
}

// FunctionHandler wraps existing functions as a ParameterHandler interface.
type FunctionHandler struct {
	fn func(params []string) error
}

func NewFunctionHandler(fn func(params []string) error) *FunctionHandler {
	return &FunctionHandler{fn: fn}
}

var _ ParameterHandler = (*FunctionHandler)(nil)

func (h *FunctionHandler) Execute(params []string) error        { return h.fn(params) }
func (h *FunctionHandler) RestoreState(state interface{}) error { return nil }
func (h *FunctionHandler) GetState() interface{}                { return nil }
