package iavl

import (
	"bytes"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

var (
	// ErrInvalidProof is returned by Verify when a proof cannot be validated.
	ErrInvalidProof = fmt.Errorf("invalid proof")

	// ErrInvalidInputs is returned when the inputs passed to the function are invalid.
	ErrInvalidInputs = fmt.Errorf("invalid inputs")

	// ErrInvalidRoot is returned when the root passed in does not match the proof's.
	ErrInvalidRoot = fmt.Errorf("invalid root")
)

//----------------------------------------

// Contract: Left and Right can never both be set. Will result in a empty `[]` roothash
type proofInnerNode struct {
	Height  int8   `json:"height"`
	Size    int64  `json:"size"`
	Version int64  `json:"version"`
	Left    []byte `json:"left"`
	Right   []byte `json:"right"`
}

func (pin proofInnerNode) String() string {
	return pin.stringIndented("")
}

func (pin proofInnerNode) stringIndented(indent string) string {
	return fmt.Sprintf(`proofInnerNode{
%s  Height:  %v
%s  Size:    %v
%s  Version: %v
%s  Left:    %X
%s  Right:   %X
%s}`,
		indent, pin.Height,
		indent, pin.Size,
		indent, pin.Version,
		indent, pin.Left,
		indent, pin.Right,
		indent)
}

func (pin proofInnerNode) Hash(childHash []byte) []byte {
	hasher := tmhash.New()
	buf := new(bytes.Buffer)

	err := amino.EncodeVarint8(buf, pin.Height)
	if err == nil {
		err = amino.EncodeVarint(buf, pin.Size)
	}
	if err == nil {
		err = amino.EncodeVarint(buf, pin.Version)
	}

	if len(pin.Left) > 0 && len(pin.Right) > 0 {
		panic("both left and right child hashes are set")
	}

	if len(pin.Left) == 0 {
		if err == nil {
			err = amino.EncodeByteSlice(buf, childHash)
		}
		if err == nil {
			err = amino.EncodeByteSlice(buf, pin.Right)
		}
	} else {
		if err == nil {
			err = amino.EncodeByteSlice(buf, pin.Left)
		}
		if err == nil {
			err = amino.EncodeByteSlice(buf, childHash)
		}
	}
	if err != nil {
		panic(fmt.Sprintf("Failed to hash proofInnerNode: %v", err))
	}

	hasher.Write(buf.Bytes())
	return hasher.Sum(nil)
}

//----------------------------------------

type proofLeafNode struct {
	Key       []byte `json:"key"`
	ValueHash []byte `json:"value"`
	Version   int64  `json:"version"`
}

func (pln proofLeafNode) String() string {
	return pln.stringIndented("")
}

func (pln proofLeafNode) stringIndented(indent string) string {
	return fmt.Sprintf(`proofLeafNode{
%s  Key:       %v
%s  ValueHash: %X
%s  Version:   %v
%s}`,
		indent, pln.Key,
		indent, pln.ValueHash,
		indent, pln.Version,
		indent)
}

func (pln proofLeafNode) Hash() []byte {
	hasher := tmhash.New()
	buf := new(bytes.Buffer)

	err := amino.EncodeVarint8(buf, 0)
	if err == nil {
		err = amino.EncodeVarint(buf, 1)
	}
	if err == nil {
		err = amino.EncodeVarint(buf, pln.Version)
	}
	if err == nil {
		err = amino.EncodeByteSlice(buf, pln.Key)
	}
	if err == nil {
		err = amino.EncodeByteSlice(buf, pln.ValueHash)
	}
	if err != nil {
		panic(fmt.Sprintf("Failed to hash proofLeafNode: %v", err))
	}
	hasher.Write(buf.Bytes())

	return hasher.Sum(nil)
}

//----------------------------------------

// If the key does not exist, returns the path to the next leaf left of key (w/
// path), except when key is less than the least item, in which case it returns
// a path to the least item.
func (node *Node) PathToLeaf(t *ImmutableTree, key []byte) (PathToLeaf, *Node, error) {
	path := new(PathToLeaf)
	val, err := node.pathToLeaf(t, key, path)
	return *path, val, err
}

// pathToLeaf is a helper which recursively constructs the PathToLeaf.
// As an optimization the already constructed path is passed in as an argument
// and is shared among recursive calls.
func (node *Node) pathToLeaf(t *ImmutableTree, key []byte, path *PathToLeaf) (*Node, error) {
	if node.height == 0 {
		if bytes.Equal(node.key, key) {
			return node, nil
		}
		return node, errors.New("key does not exist")
	}

	if bytes.Compare(key, node.key) < 0 {
		// left side
		pin := proofInnerNode{
			Height:  node.height,
			Size:    node.size,
			Version: node.version,
			Left:    nil,
			Right:   node.getRightNode(t).hash,
		}
		*path = append(*path, pin)
		n, err := node.getLeftNode(t).pathToLeaf(t, key, path)
		return n, err
	}
	// right side
	pin := proofInnerNode{
		Height:  node.height,
		Size:    node.size,
		Version: node.version,
		Left:    node.getLeftNode(t).hash,
		Right:   nil,
	}
	*path = append(*path, pin)
	n, err := node.getRightNode(t).pathToLeaf(t, key, path)
	return n, err
}
