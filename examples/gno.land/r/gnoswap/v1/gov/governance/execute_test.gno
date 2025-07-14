package governance

import (
	"errors"
	"testing"
	"time"

	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
)

func TestParameterRegistry_Register(t *testing.T) {
	testHandler := NewFunctionHandler(func(params []string) error {
		return nil
	})

	tests := []struct {
		name                     string
		pkgPath                  string
		function                 string
		handler                  ParameterHandler
		expectedHasError         bool
		expectedError            error
		expectedHandlerKey       string
		expectedHandlerKeyExists bool
	}{
		{
			name:                     "register handler is success",
			pkgPath:                  "test/pkg",
			function:                 "testFunc",
			handler:                  testHandler,
			expectedHasError:         false,
			expectedError:            nil,
			expectedHandlerKey:       "test/pkg:testFunc",
			expectedHandlerKeyExists: true,
		},
		{
			name:                     "register handler is failed by empty pkgPath", // XXX: should be failed
			pkgPath:                  "",
			function:                 "testFunc",
			handler:                  testHandler,
			expectedHasError:         true,
			expectedError:            errors.New("handler not found for :testFunc"),
			expectedHandlerKey:       ":testFunc",
			expectedHandlerKeyExists: true,
		},
		{
			name:                     "register handler is failed by empty handler",
			pkgPath:                  "non/existent",
			function:                 "missing",
			handler:                  nil,
			expectedHasError:         true,
			expectedError:            errors.New("failed to cast handler non/existent:missing to ParameterHandler"),
			expectedHandlerKey:       "non/existent:missing",
			expectedHandlerKeyExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			registry := NewParameterRegistry()
			registry.Register(tt.pkgPath, tt.function, tt.handler)

			// when
			handler, err := registry.Handler(tt.pkgPath, tt.function)

			// then
			if tt.expectedHasError {
				uassert.ErrorIs(t, err, tt.expectedError)
			} else {
				uassert.NoError(t, err)
				uassert.NotNil(t, handler)
			}

			handlerKey := makeHandlerKey(tt.pkgPath, tt.function)
			_, exists := registry.handlers.Get(handlerKey)

			uassert.Equal(t, handlerKey, tt.expectedHandlerKey)
			uassert.Equal(t, exists, tt.expectedHandlerKeyExists)
		})
	}
}

func TestExecuteParameterChange(t *testing.T) {
	tests := []struct {
		name             string
		inputMsgs        []string
		expectedHasError bool
		expectedError    error
	}{
		{
			name: "execute parameter change is success",
			inputMsgs: []string{
				"test/pkg*EXE*TestFunc*EXE*param1,param2",
			},
			expectedHasError: false,
			expectedError:    nil,
		},
		{
			name: "execute parameter change is success by multiple messages",
			inputMsgs: []string{
				"test/pkg*EXE*TestFunc*EXE*param1,param2",
				"test/pkg*EXE*TestFunc*EXE*param2,param3",
			},
			expectedHasError: false,
			expectedError:    nil,
		},
		{
			name: "execute parameter change is failed by missing separator",
			inputMsgs: []string{
				"test/pkg*EXE*TestFunc",
			},
			expectedHasError: true,
			expectedError:    errors.New("failed to parse message: test/pkg*EXE*TestFunc"),
		},
		{
			name: "execute parameter change is failed by non-existent handler",
			inputMsgs: []string{
				"unknown/pkg*EXE*UnknownFunc*EXE*param1",
			},
			expectedHasError: true,
			expectedError:    errors.New("failed to parse message: test/pkg*EXE*TestFunc"),
		},
		{
			name: "execute parameter change is failed by not enough parameters",
			inputMsgs: []string{
				"test/pkg*EXE*TestFunc*EXE*param1",
			},
			expectedHasError: true,
			expectedError:    errors.New("failed to parse message: test/pkg*EXE*TestFunc"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			registry := NewParameterRegistry()
			registry.Register("test/pkg", "TestFunc", NewFunctionHandler(func(params []string) error {
				if len(params) != 2 {
					return errors.New("invalid params length")
				}
				return nil
			}))

			// when
			err := executeParameterChange(cross, tt.inputMsgs, registry)

			// then
			if tt.expectedHasError {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestValidateProposalState(t *testing.T) {
	tests := []struct {
		name     string
		proposal *ProposalInfo
		expected ExecutionValidator
	}{
		{
			name: "validate proposal state is success by text proposal",
			proposal: &ProposalInfo{
				ProposalType: Text,
				State: ProposalState{
					Executed: false,
					Canceled: false,
					Rejected: false,
					Passed:   true,
				},
			},
			expected: ExecutionValidator{
				isTextProposal:    true,
				isAlreadyExecuted: false,
				isAlreadyCanceled: false,
				isAlreadyRejected: false,
				hasPassed:         true,
			},
		},
		{
			name: "validate proposal state is success by already executed",
			proposal: &ProposalInfo{
				ProposalType: ParameterChange,
				State: ProposalState{
					Executed: true,
					Canceled: false,
					Rejected: false,
					Passed:   true,
				},
			},
			expected: ExecutionValidator{
				isTextProposal:    false,
				isAlreadyExecuted: true,
				isAlreadyCanceled: false,
				isAlreadyRejected: false,
				hasPassed:         true,
			},
		},
		{
			name: "validate proposal state is success by canceled",
			proposal: &ProposalInfo{
				ProposalType: ParameterChange,
				State: ProposalState{
					Executed: false,
					Canceled: true,
					Rejected: false,
					Passed:   false,
				},
			},
			expected: ExecutionValidator{
				isTextProposal:    false,
				isAlreadyExecuted: false,
				isAlreadyCanceled: true,
				isAlreadyRejected: false,
				hasPassed:         false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			proposalState := validateProposalState(tt.proposal)

			// then
			uassert.Equal(t, proposalState.isTextProposal, tt.expected.isTextProposal)
			uassert.Equal(t, proposalState.isAlreadyExecuted, tt.expected.isAlreadyExecuted)
			uassert.Equal(t, proposalState.isAlreadyCanceled, tt.expected.isAlreadyCanceled)
			uassert.Equal(t, proposalState.isAlreadyRejected, tt.expected.isAlreadyRejected)
			uassert.Equal(t, proposalState.hasPassed, tt.expected.hasPassed)
		})
	}
}

func TestCheckProposalValidation(t *testing.T) {
	// TODO: change error message after error code is defined
	tests := []struct {
		name      string
		validator ExecutionValidator
		wantErr   bool
		errMsg    string
	}{
		{
			name: "check proposal validation is success by valid proposal",
			validator: ExecutionValidator{
				isTextProposal:    false,
				isAlreadyExecuted: false,
				isAlreadyCanceled: false,
				isAlreadyRejected: false,
				hasPassed:         true,
			},
			wantErr: false,
		},
		{
			name: "check proposal validation is failed by text proposal is not executable",
			validator: ExecutionValidator{
				isTextProposal:    true,
				isAlreadyExecuted: false,
				isAlreadyCanceled: false,
				isAlreadyRejected: false,
				hasPassed:         true,
			},
			wantErr: true,
			errMsg:  "[GNOSWAP-GOVERNANCE-011] can not execute text proposal",
		},
		{
			name: "check proposal validation is failed by already executed, canceled, or rejected",
			validator: ExecutionValidator{
				isTextProposal:    false,
				isAlreadyExecuted: true,
				isAlreadyCanceled: false,
				isAlreadyRejected: false,
				hasPassed:         true,
			},
			wantErr: true,
			errMsg:  "proposal already executed, canceled, or rejected",
		},
		{
			name: "check proposal validation is failed by proposal has not passed",
			validator: ExecutionValidator{
				isTextProposal:    false,
				isAlreadyExecuted: false,
				isAlreadyCanceled: false,
				isAlreadyRejected: false,
				hasPassed:         false,
			},
			wantErr: true,
			errMsg:  "[GNOSWAP-GOVERNANCE-016] proposal not passed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			err := checkProposalValidation(tt.validator)

			// then
			if tt.wantErr {
				uassert.Error(t, err, tt.errMsg)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestParameterRegistry_Execute(t *testing.T) {
	tests := []struct {
		name                    string
		pkgPath                 string
		function                string
		params                  []string
		setupMock               func(*ParameterRegistry)
		expectedHasHandlerError bool
		expectedHasExecuteError bool
	}{
		{
			name:     "execute parameter change is success by valid handler",
			pkgPath:  consts.POOL_PATH,
			function: "SetFeeProtocol",
			params:   []string{"1", "2"},
			setupMock: func(r *ParameterRegistry) {
				r.Register(consts.POOL_PATH, "SetFeeProtocol", NewFunctionHandler(func(p []string) error {
					return nil
				}))
			},
			expectedHasHandlerError: false,
			expectedHasExecuteError: false,
		},
		{
			name:                    "execute parameter change is failed by non-existent handler",
			pkgPath:                 consts.POOL_PATH,
			function:                "SetFeeProtocol",
			params:                  []string{},
			setupMock:               func(r *ParameterRegistry) {},
			expectedHasHandlerError: true,
			expectedHasExecuteError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			registry := NewParameterRegistry()
			tt.setupMock(registry)

			// when
			handler, err := registry.Handler(tt.pkgPath, tt.function)
			if tt.expectedHasHandlerError {
				uassert.Error(t, err)
				return
			}

			err = handler.Execute(tt.params)

			// then
			if tt.expectedHasExecuteError {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestRollbackSystem(t *testing.T) {
	// trace the state change
	type stateChange struct {
		pkgPath  string
		function string
		value    any
	}

	tests := []struct {
		name        string
		msgs        []string
		setupMock   func(*ParameterRegistry) ([]stateChange, func())
		expectError bool
		validate    func(*testing.T, []stateChange)
	}{
		{
			name: "Should rollback when second message fails",
			msgs: []string{
				"gno.land/r/gnoswap/v1/gns*EXE*SetAvgBlockTimeInMs*EXE*123",
				"gno.land/r/gnoswap/v1/community_pool*EXE*TransferToken*EXE*gno.land/r/gnoswap/v1/gns,g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d,905",
			},
			setupMock: func(r *ParameterRegistry) ([]stateChange, func()) {
				initialStates := []stateChange{
					{
						pkgPath:  "gno.land/r/gnoswap/v1/gns",
						function: "SetAvgBlockTimeInMs",
						value:    int64(0),
					},
				}

				// track the state change
				blockTime := int64(0)

				// first message should succeed
				r.Register("gno.land/r/gnoswap/v1/gns", "SetAvgBlockTimeInMs", NewFunctionHandler(func(params []string) error {
					blockTime = int64(123)
					initialStates[0].value = blockTime
					return nil
				}))
				// second message should fail
				r.Register("gno.land/r/gnoswap/v1/community_pool", "TransferToken", NewFunctionHandler(func(params []string) error {
					return errors.New("insufficient balance")
				}))

				// rollback the state
				rollback := func() {
					blockTime = int64(0)
					initialStates[0].value = blockTime
				}

				return initialStates, rollback
			},
			expectError: true,
			validate: func(t *testing.T, states []stateChange) {
				// should be rolled back to initial value
				uassert.Equal(t, states[0].value, int64(0))
			},
		},
		{
			name: "Should not rollback when all messages succeed",
			msgs: []string{
				"gno.land/r/gnoswap/v1/gns*EXE*SetAvgBlockTimeInMs*EXE*123",
				"gno.land/r/gnoswap/v1/community_pool*EXE*TransferToken*EXE*gno.land/r/gnoswap/v1/gns,g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d,905",
			},
			setupMock: func(r *ParameterRegistry) ([]stateChange, func()) {
				initialStates := []stateChange{
					{
						pkgPath:  "gno.land/r/gnoswap/v1/gns",
						function: "SetAvgBlockTimeInMs",
						value:    int64(0),
					},
					{
						pkgPath:  "gno.land/r/gnoswap/v1/community_pool",
						function: "TransferToken",
						value:    uint64(0),
					},
				}

				// track the state change
				blockTime := int64(0)
				transferAmount := uint64(0)

				// all messages should succeed
				r.Register("gno.land/r/gnoswap/v1/gns", "SetAvgBlockTimeInMs", NewFunctionHandler(func(params []string) error {
					blockTime = int64(123)
					initialStates[0].value = blockTime
					return nil
				}))
				r.Register("gno.land/r/gnoswap/v1/community_pool", "TransferToken", NewFunctionHandler(func(params []string) error {
					transferAmount = uint64(905)
					initialStates[1].value = transferAmount
					return nil
				}))

				// rollback function (not used in this case)
				rollback := func() {
					blockTime = int64(0)
					transferAmount = uint64(0)
					initialStates[0].value = blockTime
					initialStates[1].value = transferAmount
				}

				return initialStates, rollback
			},
			expectError: false,
			validate: func(t *testing.T, states []stateChange) {
				// should be changed
				uassert.Equal(t, states[0].value, int64(123))
				uassert.Equal(t, states[1].value, uint64(905))
			},
		},
		{
			name: "Should not rollback when first message fails",
			msgs: []string{
				"gno.land/r/gnoswap/v1/gns*EXE*SetAvgBlockTimeInMs*EXE*invalid",
				"gno.land/r/gnoswap/v1/community_pool*EXE*TransferToken*EXE*gno.land/r/gnoswap/v1/gns,g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d,905",
			},
			setupMock: func(r *ParameterRegistry) ([]stateChange, func()) {
				initialStates := []stateChange{
					{
						pkgPath:  "gno.land/r/gnoswap/v1/gns",
						function: "SetAvgBlockTimeInMs",
						value:    int64(0),
					},
					{
						pkgPath:  "gno.land/r/gnoswap/v1/community_pool",
						function: "TransferToken",
						value:    uint64(0),
					},
				}

				// track the state change
				blockTime := int64(0)
				transferAmount := uint64(0)

				// first message should fail
				r.Register("gno.land/r/gnoswap/v1/gns", "SetAvgBlockTimeInMs", NewFunctionHandler(func(params []string) error {
					return errors.New("invalid parameter")
				}))
				// second message should not be executed
				r.Register("gno.land/r/gnoswap/v1/community_pool", "TransferToken", NewFunctionHandler(func(params []string) error {
					transferAmount = uint64(905)
					initialStates[1].value = transferAmount
					return nil
				}))

				// rollback function (not used in this case)
				rollback := func() {
					blockTime = int64(0)
					transferAmount = uint64(0)
					initialStates[0].value = blockTime
					initialStates[1].value = transferAmount
				}

				return initialStates, rollback
			},
			expectError: true,
			validate: func(t *testing.T, states []stateChange) {
				// should not be changed
				uassert.Equal(t, states[0].value, int64(0))
				uassert.Equal(t, states[1].value, uint64(0))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a test registry
			registry := NewParameterRegistry()
			initialStates, rollback := tt.setupMock(registry)

			// create a test proposal
			proposal := &ProposalInfo{
				ProposalType: ParameterChange,
				State: ProposalState{
					Passed: true,
				},
				Execution: ExecutionInfo{
					Msgs: tt.msgs,
				},
			}

			// create an execution context
			ctx := &ExecutionContext{
				ProposalId: 1,
				Now:        uint64(time.Now().Unix()),
				Config:     &Config{},
				Proposal:   proposal,
			}

			err := executeProposal(cross, ctx, registry)

			// if error occurs, execute rollback
			if err != nil {
				rollback()
			}

			if tt.expectError {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}

			// validate state
			if tt.validate != nil {
				tt.validate(t, initialStates)
			}
		})
	}
}
