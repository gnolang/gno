package integration

// ClusterOverrides are the cluster settings named on the command line. Each is
// nil unless it was actually given, which is what separates a flag left at its
// default from one set to that same value.
//
// A named setting wins over what a script declares, and the caller owns the
// result: running a scenario that declared three validators against two is
// allowed, and that scenario may go red. That is the point of an override.
type ClusterOverrides struct {
	Validators           *int
	Oracle               *bool
	CodeSubmissionPolicy *string
	// PkgApprover is a bech32 address here, where a script names a role.
	// ClusterSpec.ApplyTo accepts either.
	PkgApprover *string
	BlockMaxGas *int64
}

// Apply returns the cluster a script will actually run against, leaving the
// declaration itself untouched. A setting named on the command line wins over
// the one the script declared.
func (o ClusterOverrides) Apply(spec ClusterSpec) ClusterSpec {
	if o.Validators != nil {
		spec.Validators = *o.Validators
	}
	if o.Oracle != nil {
		spec.Oracle = *o.Oracle
	}
	if o.CodeSubmissionPolicy != nil {
		spec.CodeSubmissionPolicy = *o.CodeSubmissionPolicy
	}
	if o.PkgApprover != nil {
		spec.PkgApprover = *o.PkgApprover
	}
	if o.BlockMaxGas != nil {
		spec.BlockMaxGas = *o.BlockMaxGas
	}
	return spec
}
