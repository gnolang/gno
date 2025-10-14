package ics23

import (
	"bytes"
	"errors"

	"gno.land/p/nt/ufmt"
)

type CommitmentProof_Exist struct {
	Exist *ExistenceProof
}

// ExistenceProof takes a key and a value and a set of steps to perform on it.
// The result of peforming all these steps will provide a "root hash", which can
// be compared to the value in a header.
//
// Since it is computationally infeasible to produce a hash collission for any of the used
// cryptographic hash functions, if someone can provide a series of operations to transform
// a given key and value into a root hash that matches some trusted root, these key and values
// must be in the referenced merkle tree.
//
// The only possible issue is maliablity in LeafOp, such as providing extra prefix data,
// which should be controlled by a spec. Eg. with lengthOp as NONE,
// prefix = FOO, key = BAR, value = CHOICE
// and
// prefix = F, key = OOBAR, value = CHOICE
// would produce the same value.
//
// With LengthOp this is tricker but not impossible. Which is why the "leafPrefixEqual" field
// in the ProofSpec is valuable to prevent this mutability. And why all trees should
// length-prefix the data before hashing it.
type ExistenceProof struct {
	Key   []byte
	Value []byte
	Leaf  *LeafOp
	Path  []*InnerOp
}

// LeafOp represents the raw key-value data we wish to prove, and
// must be flexible to represent the internal transformation from
// the original key-value pairs into the basis hash, for many existing
// merkle trees.
//
// key and value are passed in. So that the signature of this operation is:
// leafOp(key, value) -> output
//
// To process this, first prehash the keys and values if needed (ANY means no hash in this case):
// hkey = prehashKey(key)
// hvalue = prehashValue(value)
//
// Then combine the bytes, and hash it
// output = hash(prefix || length(hkey) || hkey || length(hvalue) || hvalue)
type LeafOp struct {
	Hash         HashOp
	PrehashKey   HashOp
	PrehashValue HashOp
	Length       LengthOp
	// prefix is a fixed bytes that may optionally be included at the beginning to differentiate
	// a leaf node from an inner node.
	Prefix []byte
}

func (m *LeafOp) GetHash() HashOp {
	if m != nil {
		return m.Hash
	}
	return HashOp_NO_HASH
}

func (m *LeafOp) GetPrefix() []byte {
	return m.Prefix
}

// InnerOp represents a merkle-proof step that is not a leaf.
// It represents concatenating two children and hashing them to provide the next result.
//
// The result of the previous step is passed in, so the signature of this op is:
// innerOp(child) -> output
//
// The result of applying InnerOp should be:
// output = op.hash(op.prefix || child || op.suffix)
//
// where the || operator is concatenation of binary data,
// and child is the result of hashing all the tree below this step.
//
// Any special data, like prepending child with the length, or prepending the entire operation with
// some value to differentiate from leaf nodes, should be included in prefix and suffix.
// If either of prefix or suffix is empty, we just treat it as an empty string
type InnerOp struct {
	Hash   HashOp
	Prefix []byte
	Suffix []byte
}

func (m *InnerOp) GetHash() HashOp {
	if m != nil {
		return m.Hash
	}
	return HashOp_NO_HASH
}

func (m *InnerOp) GetPrefix() []byte {
	return m.Prefix
}

// Verify does all checks to ensure this proof proves this key, value -> root
// and matches the spec.
func (p *ExistenceProof) Verify(spec *ProofSpec, root CommitmentRoot, key []byte, value []byte) error {
	if err := p.CheckAgainstSpec(spec); err != nil {
		return err
	}

	if !bytes.Equal(key, p.Key) {
		return ufmt.Errorf("provided key doesn't match proof")
	}
	if !bytes.Equal(value, p.Value) {
		return ufmt.Errorf("provided value doesn't match proof")
	}

	calc, err := p.calculate(spec)
	if err != nil {
		return ufmt.Errorf("error calculating root, %w", err)
	}
	if !bytes.Equal(root, calc) {
		return ufmt.Errorf("calculated root doesn't match provided root")
	}

	return nil
}

// Calculate determines the root hash that matches the given proof.
// You must validate the result is what you have in a header.
// Returns error if the calculations cannot be performed.
func (p *ExistenceProof) Calculate() (CommitmentRoot, error) {
	return p.calculate(nil)
}

func (p *ExistenceProof) calculate(spec *ProofSpec) (CommitmentRoot, error) {
	if p.Leaf == nil {
		return nil, errors.New("existence Proof needs defined LeafOp")
	}

	// leaf step takes the key and value as input
	res, err := p.Leaf.Apply(p.Key, p.Value)
	if err != nil {
		return nil, ufmt.Errorf("leaf, %w", err)
	}

	// the rest just take the output of the last step (reducing it)
	for _, step := range p.Path {
		res, err = step.Apply(res)
		if err != nil {
			return nil, ufmt.Errorf("inner, %w", err)
		}
		if spec != nil {
			if len(res) > int(spec.InnerSpec.ChildSize) && int(spec.InnerSpec.ChildSize) >= 32 {
				return nil, ufmt.Errorf("inner, %w", err)
			}
		}
	}
	return res, nil
}

type CommitmentProof_Nonexist struct {
	Nonexist *NonExistenceProof
}

// NonExistenceProof takes a proof of two neighbors, one left of the desired key,
// one right of the desired key. If both proofs are valid AND they are neighbors,
// then there is no valid proof for the given key.
type NonExistenceProof struct {
	Key   []byte
	Left  *ExistenceProof
	Right *ExistenceProof
}

// Calculate determines the root hash that matches the given nonexistence proof.
// You must validate the result is what you have in a header.
// Returns error if the calculations cannot be performed.
func (p *NonExistenceProof) Calculate() (CommitmentRoot, error) {
	// A Nonexist proof may have left or right proof nil
	switch {
	case p.Left != nil:
		return p.Left.Calculate()
	case p.Right != nil:
		return p.Right.Calculate()
	default:
		return nil, errors.New("nonexistence proof has empty Left and Right proof")
	}
}

// CheckAgainstSpec will verify the leaf and all path steps are in the format defined in spec
func (p *ExistenceProof) CheckAgainstSpec(spec *ProofSpec) error {
	leaf := p.Leaf
	if leaf == nil {
		return errors.New("existence Proof needs defined LeafOp")
	}
	if err := leaf.CheckAgainstSpec(spec); err != nil {
		return ufmt.Errorf("leaf, %w", err)
	}
	if spec.MinDepth > 0 && len(p.Path) < int(spec.MinDepth) {
		return ufmt.Errorf("innerOps depth too short: %d", len(p.Path))
	}

	maxDepth := spec.MaxDepth
	if maxDepth == 0 {
		maxDepth = 128
	}

	if len(p.Path) > int(maxDepth) {
		return ufmt.Errorf("innerOps depth too long: %d", len(p.Path))
	}

	layerNum := 1

	for _, inner := range p.Path {
		if err := inner.CheckAgainstSpec(spec, layerNum); err != nil {
			return ufmt.Errorf("inner, %w", err)
		}
		layerNum++
	}
	return nil
}

// If we should prehash the key before comparison, do so; otherwise, return the key. Prehashing
// changes lexical comparison, so we do so before comparison if the spec sets
// `PrehashKeyBeforeComparison`.
func keyForComparison(spec *ProofSpec, key []byte) []byte {
	if !spec.PrehashKeyBeforeComparison {
		return key
	}
	hash, _ := doHashOrNoop(spec.LeafSpec.PrehashKey, key)
	return hash
}

// Verify does all checks to ensure the proof has valid non-existence proofs,
// and they ensure the given key is not in the CommitmentState
func (p *NonExistenceProof) Verify(spec *ProofSpec, root CommitmentRoot, key []byte) error {
	// ensure the existence proofs are valid
	var leftKey, rightKey []byte
	if p.Left != nil {
		if err := p.Left.Verify(spec, root, p.Left.Key, p.Left.Value); err != nil {
			return ufmt.Errorf("left proof, %w", err)
		}
		leftKey = p.Left.Key
	}
	if p.Right != nil {
		if err := p.Right.Verify(spec, root, p.Right.Key, p.Right.Value); err != nil {
			return ufmt.Errorf("right proof, %w", err)
		}
		rightKey = p.Right.Key
	}

	// If both proofs are missing, this is not a valid proof
	if leftKey == nil && rightKey == nil {
		return errors.New("both left and right proofs missing")
	}

	// Ensure in valid range
	if rightKey != nil {
		if bytes.Compare(keyForComparison(spec, key), keyForComparison(spec, rightKey)) >= 0 {
			return errors.New("key is not left of right proof")
		}
	}

	if leftKey != nil {
		if bytes.Compare(keyForComparison(spec, key), keyForComparison(spec, leftKey)) <= 0 {
			return errors.New("key is not right of left proof")
		}
	}

	switch {
	case leftKey == nil:
		isLeftMost, err := IsLeftMost(spec.InnerSpec, p.Right.Path)
		if err != nil {
			return err
		}

		if !isLeftMost {
			return errors.New("left proof missing, right proof must be left-most")
		}
	case rightKey == nil:
		isRightMost, err := IsRightMost(spec.InnerSpec, p.Left.Path)
		if err != nil {
			return err
		}

		if !isRightMost {
			return errors.New("right proof missing, left proof must be right-most")
		}
	default:
		isLeftNeighbor, err := IsLeftNeighbor(spec.InnerSpec, p.Left.Path, p.Right.Path)
		if err != nil {
			return err
		}

		if !isLeftNeighbor {
			return errors.New("right proof missing, left proof must be right-most")
		}
	}

	return nil
}

// IsLeftMost returns true if this is the left-most path in the tree, excluding placeholder (empty child) nodes
func IsLeftMost(spec *InnerSpec, path []*InnerOp) (bool, error) {
	minPrefix, maxPrefix, suffix, err := getPadding(spec, 0)
	if err != nil {
		return false, err
	}

	// ensure every step has a prefix and suffix defined to be leftmost, unless it is a placeholder node
	for _, step := range path {
		if !hasPadding(step, minPrefix, maxPrefix, suffix) {
			leftBranchesAreEmpty, err := leftBranchesAreEmpty(spec, step)
			if err != nil {
				return false, err
			}

			if !leftBranchesAreEmpty {
				return false, nil
			}
		}
	}
	return true, nil
}

// IsRightMost returns true if this is the right-most path in the tree, excluding placeholder (empty child) nodes
func IsRightMost(spec *InnerSpec, path []*InnerOp) (bool, error) {
	last := len(spec.ChildOrder) - 1
	minPrefix, maxPrefix, suffix, err := getPadding(spec, int32(last))
	if err != nil {
		return false, err
	}

	// ensure every step has a prefix and suffix defined to be rightmost, unless it is a placeholder node
	for _, step := range path {
		if !hasPadding(step, minPrefix, maxPrefix, suffix) {
			rightBranchesAreEmpty, err := rightBranchesAreEmpty(spec, step)
			if err != nil {
				return false, err
			}

			if !rightBranchesAreEmpty {
				return false, nil
			}
		}
	}
	return true, nil
}

// IsLeftNeighbor returns true if `right` is the next possible path right of `left`
//
//	Find the common suffix from the Left.Path and Right.Path and remove it. We have LPath and RPath now, which must be neighbors.
//	Validate that LPath[len-1] is the left neighbor of RPath[len-1]
//	For step in LPath[0..len-1], validate step is right-most node
//	For step in RPath[0..len-1], validate step is left-most node
func IsLeftNeighbor(spec *InnerSpec, left []*InnerOp, right []*InnerOp) (bool, error) {
	// count common tail (from end, near root)
	left, topleft := left[:len(left)-1], left[len(left)-1]
	right, topright := right[:len(right)-1], right[len(right)-1]
	for bytes.Equal(topleft.Prefix, topright.Prefix) && bytes.Equal(topleft.Suffix, topright.Suffix) {
		left, topleft = left[:len(left)-1], left[len(left)-1]
		right, topright = right[:len(right)-1], right[len(right)-1]
	}

	// now topleft and topright are the first divergent nodes
	// make sure they are left and right of each other
	if !isLeftStep(spec, topleft, topright) {
		return false, nil
	}

	// left and right are remaining children below the split,
	// ensure left child is the rightmost path, and visa versa
	isRightMost, err := IsRightMost(spec, left)
	if err != nil {
		return false, err
	}
	if !isRightMost {
		return false, nil
	}

	isLeftMost, err := IsLeftMost(spec, right)
	if err != nil {
		return false, err
	}
	if !isLeftMost {
		return false, nil
	}
	return true, nil
}

// isLeftStep assumes left and right have common parents
// checks if left is exactly one slot to the left of right
func isLeftStep(spec *InnerSpec, left *InnerOp, right *InnerOp) bool {
	leftidx, err := orderFromPadding(spec, left)
	if err != nil {
		panic(err)
	}
	rightidx, err := orderFromPadding(spec, right)
	if err != nil {
		panic(err)
	}

	// TODO: is it possible there are empty (nil) children???
	return rightidx == leftidx+1
}

// checks if an op has the expected padding
func hasPadding(op *InnerOp, minPrefix, maxPrefix, suffix int) bool {
	if len(op.Prefix) < minPrefix {
		return false
	}
	if len(op.Prefix) > maxPrefix {
		return false
	}
	return len(op.Suffix) == suffix
}

// getPadding determines prefix and suffix with the given spec and position in the tree
func getPadding(spec *InnerSpec, branch int32) (minPrefix, maxPrefix, suffix int, err error) {
	idx, err := getPosition(spec.ChildOrder, branch)
	if err != nil {
		return 0, 0, 0, err
	}

	// count how many children are in the prefix
	prefix := idx * int(spec.ChildSize)
	minPrefix = prefix + int(spec.MinPrefixLength)
	maxPrefix = prefix + int(spec.MaxPrefixLength)

	// count how many children are in the suffix
	suffix = (len(spec.ChildOrder) - 1 - idx) * int(spec.ChildSize)
	return minPrefix, maxPrefix, suffix, nil
}

// leftBranchesAreEmpty returns true if the padding bytes correspond to all empty siblings
// on the left side of a branch, ie. it's a valid placeholder on a leftmost path
func leftBranchesAreEmpty(spec *InnerSpec, op *InnerOp) (bool, error) {
	idx, err := orderFromPadding(spec, op)
	if err != nil {
		return false, err
	}
	// count branches to left of this
	leftBranches := int(idx)
	if leftBranches == 0 {
		return false, nil
	}
	// compare prefix with the expected number of empty branches
	actualPrefix := len(op.Prefix) - leftBranches*int(spec.ChildSize)
	if actualPrefix < 0 {
		return false, nil
	}
	for i := 0; i < leftBranches; i++ {
		idx, err := getPosition(spec.ChildOrder, int32(i))
		if err != nil {
			return false, err
		}

		from := actualPrefix + idx*int(spec.ChildSize)
		if !bytes.Equal(spec.EmptyChild, op.Prefix[from:from+int(spec.ChildSize)]) {
			return false, nil
		}
	}
	return true, nil
}

// rightBranchesAreEmpty returns true if the padding bytes correspond to all empty siblings
// on the right side of a branch, ie. it's a valid placeholder on a rightmost path
func rightBranchesAreEmpty(spec *InnerSpec, op *InnerOp) (bool, error) {
	idx, err := orderFromPadding(spec, op)
	if err != nil {
		return false, err
	}
	// count branches to right of this one
	rightBranches := len(spec.ChildOrder) - 1 - int(idx)
	if rightBranches == 0 {
		return false, nil
	}
	// compare suffix with the expected number of empty branches
	if len(op.Suffix) != rightBranches*int(spec.ChildSize) {
		return false, nil // sanity check
	}
	for i := 0; i < rightBranches; i++ {
		idx, err := getPosition(spec.ChildOrder, int32(i))
		if err != nil {
			return false, err
		}

		from := idx * int(spec.ChildSize)
		if !bytes.Equal(spec.EmptyChild, op.Suffix[from:from+int(spec.ChildSize)]) {
			return false, nil
		}
	}
	return true, nil
}

// getPosition checks where the branch is in the order and returns
// the index of this branch
func getPosition(order []int32, branch int32) (int, error) {
	if branch < 0 || int(branch) >= len(order) {
		return -1, ufmt.Errorf("invalid branch: %d", branch)
	}
	for i, item := range order {
		if branch == item {
			return i, nil
		}
	}

	return -1, ufmt.Errorf("branch %d not found in order %v", branch, order)
}

// This will look at the proof and determine which order it is...
// So we can see if it is branch 0, 1, 2 etc... to determine neighbors
func orderFromPadding(spec *InnerSpec, inner *InnerOp) (int32, error) {
	maxbranch := int32(len(spec.ChildOrder))
	for branch := int32(0); branch < maxbranch; branch++ {
		minp, maxp, suffix, err := getPadding(spec, branch)
		if err != nil {
			return 0, err
		}
		if hasPadding(inner, minp, maxp, suffix) {
			return branch, nil
		}
	}
	return 0, errors.New("cannot find any valid spacing for this node")
}

// over-declares equality, which we consider fine for now.
func (p *ProofSpec) SpecEquals(spec *ProofSpec) bool {
	// 1. Compare LeafSpecs values.
	switch {
	case (p.LeafSpec == nil) != (spec.LeafSpec == nil): // One of them is nil.
		return false

	case p.LeafSpec != nil && spec.LeafSpec != nil:
		ok := p.LeafSpec.Hash == spec.LeafSpec.Hash &&
			p.LeafSpec.PrehashKey == spec.LeafSpec.PrehashKey &&
			p.LeafSpec.PrehashValue == spec.LeafSpec.PrehashValue &&
			p.LeafSpec.Length == spec.LeafSpec.Length
		if !ok {
			return false
		}

	default: // Both are nil, hence LeafSpec values are equal.
	}

	// 2. Compare InnerSpec values.
	switch {
	case (p.InnerSpec == nil) != (spec.InnerSpec == nil): // One of them is not nil.
		return false

	case p.InnerSpec != nil && spec.InnerSpec != nil: // Both are non-nil
		ok := p.InnerSpec.Hash == spec.InnerSpec.Hash &&
			p.InnerSpec.MinPrefixLength == spec.InnerSpec.MinPrefixLength &&
			p.InnerSpec.MaxPrefixLength == spec.InnerSpec.MaxPrefixLength &&
			p.InnerSpec.ChildSize == spec.InnerSpec.ChildSize &&
			len(p.InnerSpec.ChildOrder) == len(spec.InnerSpec.ChildOrder)
		if !ok {
			return false
		}

	default: // Both are nil, hence InnerSpec values are equal.
	}

	// By this point all the above conditions pass so they are equal.
	return true
}
