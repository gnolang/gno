package governance

import (
	"errors"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
)

func TestParameterRegistry_Register(t *testing.T) {
	registry := NewParameterRegistry()

	testHandler := func(params []string) error {
		return nil
	}

	tests := []struct {
		name     string
		pkgPath  string
		function string
		handler  ParameterHandler
	}{
		{
			name:     "pass",
			pkgPath:  "test/pkg",
			function: "testFunc",
			handler:  testHandler,
		},
		{
			name:     "empty pass",
			pkgPath:  "",
			function: "testFunc",
			handler:  testHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry.Register(tt.pkgPath, tt.function, tt.handler)

			handler, err := registry.Handler(tt.pkgPath, tt.function)
			uassert.NoError(t, err)

			if handler == nil {
				t.Errorf("handler is nil")
			}

			expectedKey := makeHandlerKey(tt.pkgPath, tt.function)
			if _, exists := registry.handlers.Get(expectedKey); !exists {
				t.Errorf("expected key %s not found", expectedKey)
			}
		})
	}
}

func TestParameterRegistry(t *testing.T) {
	tests := []struct {
		name       string
		pkgPath    string
		function   string
		handler    ParameterHandler
		wantErr    bool
		errMessage string
	}{
		{
			name:     "should register and retrieve handler successfully",
			pkgPath:  "test/pkg",
			function: "testFunc",
			handler: func(params []string) error {
				return nil
			},
			wantErr:    false,
			errMessage: "",
		},
		{
			name:       "should return error for non-existent handler",
			pkgPath:    "non/existent",
			function:   "missing",
			handler:    nil,
			wantErr:    true,
			errMessage: "handler not found for non/existent:missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewParameterRegistry()

			// Register handler if provided
			if tt.handler != nil {
				registry.Register(tt.pkgPath, tt.function, tt.handler)
			}

			// Try to retrieve handler
			handler, err := registry.Handler(tt.pkgPath, tt.function)

			if tt.wantErr {
				uassert.Error(t, err, tt.errMessage)
				uassert.Equal(t, err.Error(), tt.errMessage)
			} else {
				uassert.NoError(t, err)
				if handler == nil {
					t.Error("expected handler to be non-nil")
				}
			}
		})
	}
}

func TestExecuteParameterChange(t *testing.T) {
	registry := NewParameterRegistry()

	registry.Register("test/pkg", "TestFunc", func(params []string) error {
		if len(params) != 2 {
			return errors.New("invalid params length")
		}
		return nil
	})

	tests := []struct {
		name    string
		msgs    []string
		wantErr bool
	}{
		{
			name: "Pass: Valid message",
			msgs: []string{
				"test/pkg*EXE*TestFunc*EXE*param1,param2",
			},
			wantErr: false,
		},
		{
			name: "Fail: Missing separator",
			msgs: []string{
				"test/pkg*EXE*TestFunc",
			},
			wantErr: true,
		},
		{
			name: "Fail: Non-existent handler",
			msgs: []string{
				"unknown/pkg*EXE*UnknownFunc*EXE*param1",
			},
			wantErr: true,
		},
		{
			name: "Fail: Not enough parameters",
			msgs: []string{
				"test/pkg*EXE*TestFunc*EXE*param1",
			},
			wantErr: true,
		},
		{
			name: "handle multiple messages",
			msgs: []string{
				"test/pkg*EXE*TestFunc*EXE*param1,param2",
				"test/pkg*EXE*TestFunc*EXE*param2,param3",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeParameterChange(tt.msgs, registry)
			if !tt.wantErr {
				uassert.NoError(t, err)
			} else {
				uassert.Error(t, err)
			}
		})
	}
}

func TestValidateProposalState(t *testing.T) {
	tests := []struct {
		name     string
		proposal *ProposalInfo
		want     ExecutionValidator
	}{
		{
			name: "Pass: Text proposal",
			proposal: &ProposalInfo{
				ProposalType: Text,
				State: ProposalState{
					Executed: false,
					Canceled: false,
					Rejected: false,
					Passed:   true,
				},
			},
			want: ExecutionValidator{
				isTextProposal:    true,
				isAlreadyExecuted: false,
				isAlreadyCanceled: false,
				isAlreadyRejected: false,
				hasPassed:         true,
			},
		},
		{
			name: "Pass: Already executed",
			proposal: &ProposalInfo{
				ProposalType: ParameterChange,
				State: ProposalState{
					Executed: true,
					Canceled: false,
					Rejected: false,
					Passed:   true,
				},
			},
			want: ExecutionValidator{
				isTextProposal:    false,
				isAlreadyExecuted: true,
				isAlreadyCanceled: false,
				isAlreadyRejected: false,
				hasPassed:         true,
			},
		},
		{
			name: "Pass: Canceled",
			proposal: &ProposalInfo{
				ProposalType: ParameterChange,
				State: ProposalState{
					Executed: false,
					Canceled: true,
					Rejected: false,
					Passed:   false,
				},
			},
			want: ExecutionValidator{
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
			got := validateProposalState(tt.proposal)
			if got != tt.want {
				t.Errorf("validateProposalState() = %v, want %v", got, tt.want)
			}
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
			name: "Pass: Valid proposal",
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
			name: "Fail: Text proposal is not executable",
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
			name: "Fail: Already executed, canceled, or rejected",
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
			name: "Fail: Proposal has not passed",
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
			err := checkProposalValidation(tt.validator)
			if tt.wantErr {
				uassert.Error(t, err, tt.errMsg)
				uassert.Equal(t, err.Error(), tt.errMsg)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestParameterRegistry2(t *testing.T) {
	tests := []struct {
		name        string
		pkgPath     string
		function    string
		params      []string
		setupMock   func(*ParameterRegistry)
		expectError bool
	}{
		{
			name:     "valid handler",
			pkgPath:  consts.POOL_PATH,
			function: "SetFeeProtocol",
			params:   []string{"1", "2"},
			setupMock: func(r *ParameterRegistry) {
				r.Register(consts.POOL_PATH, "SetFeeProtocol", func(p []string) error {
					return nil
				})
			},
			expectError: false,
		},
		{
			name:        "invalid handler",
			pkgPath:     "invalid",
			function:    "invalid",
			params:      []string{},
			setupMock:   func(r *ParameterRegistry) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewParameterRegistry()
			tt.setupMock(registry)

			handler, err := registry.Handler(tt.pkgPath, tt.function)
			if tt.expectError {
				uassert.Error(t, err)
				return
			}

			err = handler(tt.params)
			if err != nil {
				t.Errorf("handler returned error: %v", err)
			}
		})
	}
}
