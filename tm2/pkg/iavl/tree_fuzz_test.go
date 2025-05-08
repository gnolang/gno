package iavl

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/random"
)

// This file implement fuzz testing by generating programs and then running
// them. If an error occurs, the program that had the error is printed.

// A program is a list of instructions.
type program struct {
	Instructions []instruction `json:"instructions"`
}

func (p *program) Execute(tree *MutableTree) (err error) {
	var errLine int

	defer func() {
		r := recover()
		if r == nil {
			return
		}

		// These are simply input errors and shouldn't be reported as actual logical issues.
		if containsAny(fmt.Sprint(r), "Unrecognized op:", "Attempt to store nil value at key") {
			return
		}

		var str string

		for i, instr := range p.Instructions {
			prefix := "   "
			if i == errLine {
				prefix = ">> "
			}
			str += prefix + instr.String() + "\n"
		}
		err = fmt.Errorf("Program panicked with: %s\n%s", r, str)
	}()

	for i, instr := range p.Instructions {
		errLine = i
		instr.Execute(tree)
	}
	return
}

func (p *program) addInstruction(i instruction) {
	p.Instructions = append(p.Instructions, i)
}

func (p *program) size() int {
	return len(p.Instructions)
}

type instruction struct {
	Op      string
	K, V    []byte
	Version int64
}

func (i instruction) Execute(tree *MutableTree) {
	switch i.Op {
	case "SET":
		tree.Set(i.K, i.V)
	case "REMOVE":
		tree.Remove(i.K)
	case "SAVE":
		tree.SaveVersion()
	case "DELETE":
		tree.DeleteVersion(i.Version)
	default:
		panic("Unrecognized op: " + i.Op)
	}
}

func (i instruction) String() string {
	if i.Version > 0 {
		return fmt.Sprintf("%-8s %-8s %-8s %-8d", i.Op, i.K, i.V, i.Version)
	}
	return fmt.Sprintf("%-8s %-8s %-8s", i.Op, i.K, i.V)
}

// Generate a random program of the given size.
func genRandomProgram(size int) *program {
	p := &program{}
	nextVersion := 1

	for p.size() < size {
		k, v := []byte(random.RandStr(1)), []byte(random.RandStr(1))

		switch rand.Int() % 7 {
		case 0, 1, 2:
			p.addInstruction(instruction{Op: "SET", K: k, V: v})
		case 3, 4:
			p.addInstruction(instruction{Op: "REMOVE", K: k})
		case 5:
			p.addInstruction(instruction{Op: "SAVE", Version: int64(nextVersion)})
			nextVersion++
		case 6:
			if rv := rand.Int() % nextVersion; rv < nextVersion && rv > 0 {
				p.addInstruction(instruction{Op: "DELETE", Version: int64(rv)})
			}
		}
	}
	return p
}

// Generate many programs and run them.
func TestMutableTreeFuzz(t *testing.T) {
	t.Parallel()

	runThenGenerateMutableTreeFuzzSeeds(t, false)
}

var pathForMutableTreeProgramSeeds = filepath.Join("testdata", "corpora", "mutable_tree_programs")

func runThenGenerateMutableTreeFuzzSeeds(tb testing.TB, writeSeedsToFileSystem bool) {
	tb.Helper()

	if testing.Short() {
		tb.Skip("Running in -short mode")
	}

	maxIterations := testFuzzIterations
	progsPerIteration := 100000
	iterations := 0

	if writeSeedsToFileSystem {
		if err := os.MkdirAll(pathForMutableTreeProgramSeeds, 0o755); err != nil {
			tb.Fatal(err)
		}
	}

	for size := 5; iterations < maxIterations; size++ {
		for i := range progsPerIteration / size {
			tree := NewMutableTree(memdb.NewMemDB(), 0)
			program := genRandomProgram(size)
			err := program.Execute(tree)
			if err != nil {
				tb.Fatalf("Error after %d iterations (size %d): %s\n%s", iterations, size, err.Error(), tree.String())
			}
			iterations++

			if !writeSeedsToFileSystem {
				continue
			}

			// Otherwise write them to the testdata/corpra directory.
			programJSON, err := json.Marshal(program)
			if err != nil {
				tb.Fatal(err)
			}
			path := filepath.Join(pathForMutableTreeProgramSeeds, fmt.Sprintf("%d", i+1))
			if err := os.WriteFile(path, programJSON, 0o755); err != nil {
				tb.Fatal(err)
			}
		}
	}
}

type treeRange struct {
	Start   []byte
	End     []byte
	Forward bool
}

var basicRecords = []struct {
	key, value string
}{
	{"abc", "123"},
	{"low", "high"},
	{"fan", "456"},
	{"foo", "a"},
	{"foobaz", "c"},
	{"good", "bye"},
	{"foobang", "d"},
	{"foobar", "b"},
	{"food", "e"},
	{"foml", "f"},
}

// Allows hooking into Go's fuzzers and then for continuous fuzzing
// enriched with coverage guided mutations, instead of naive mutations.
func FuzzIterateRange(f *testing.F) {
	if testing.Short() {
		f.Skip("Skipping in -short mode")
	}

	// 1. Add the seeds.
	seeds := []*treeRange{
		{[]byte("foo"), []byte("goo"), true},
		{[]byte("aaa"), []byte("abb"), true},
		{nil, []byte("flap"), true},
		{[]byte("foob"), nil, true},
		{[]byte("very"), nil, true},
		{[]byte("very"), nil, false},
		{[]byte("fooba"), []byte("food"), true},
		{[]byte("fooba"), []byte("food"), false},
		{[]byte("g"), nil, false},
	}
	for _, seed := range seeds {
		blob, err := json.Marshal(seed)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(blob)
	}

	db := memdb.NewMemDB()
	tree := NewMutableTree(db, 0)
	for _, br := range basicRecords {
		tree.Set([]byte(br.key), []byte(br.value))
	}

	var trav traverser

	// 2. Run the fuzzer.
	f.Fuzz(func(t *testing.T, rangeJSON []byte) {
		tr := new(treeRange)
		if err := json.Unmarshal(rangeJSON, tr); err != nil {
			return
		}

		tree.IterateRange(tr.Start, tr.End, tr.Forward, trav.view)
	})
}

func containsAny(s string, anyOf ...string) bool {
	for _, q := range anyOf {
		if strings.Contains(s, q) {
			return true
		}
	}
	return false
}

func FuzzMutableTreeInstructions(f *testing.F) {
	if testing.Short() {
		f.Skip("Skipping in -short mode")
	}

	// 0. Generate then add the seeds.
	runThenGenerateMutableTreeFuzzSeeds(f, true)

	// 1. Add the seeds.
	dir := os.DirFS("testdata")
	err := fs.WalkDir(dir, ".", func(path string, de fs.DirEntry, err error) error {
		if de.IsDir() {
			return err
		}

		ff, err := dir.Open(path)
		if err != nil {
			return err
		}
		defer ff.Close()

		blob, err := io.ReadAll(ff)
		if err != nil {
			return err
		}
		f.Add(blob)
		return nil
	})
	if err != nil {
		f.Fatal(err)
	}

	// 2. Run the fuzzer.
	f.Fuzz(func(t *testing.T, programJSON []byte) {
		program := new(program)
		if err := json.Unmarshal(programJSON, program); err != nil {
			return
		}

		tree := NewMutableTree(memdb.NewMemDB(), 0)
		err := program.Execute(tree)
		if err != nil {
			t.Fatal(err)
		}
	})
}
