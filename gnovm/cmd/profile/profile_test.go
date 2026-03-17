package profile

import (
	"io"
	"math"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Full-VM benchmarks for pprof profiling and before/after comparisons.
// Parse and preprocess happen once; only Machine.RunMain() is measured.
//
// Usage:
//   go test -bench=BenchmarkVM -benchtime=10s -count=10 . > bench.txt
//   benchstat before.txt after.txt
//
// Profiling:
//   go test -bench=BenchmarkVM_GasMetered -benchtime=10s -cpuprofile=cpu.prof -memprofile=mem.prof .
//   go tool pprof -http=:8080 cpu.prof

// programSynthetic exercises the main VM paths: arithmetic, comparisons,
// function calls, recursion, loops, slice/map operations, struct method
// dispatch, closures, interface dispatch, type switches, pointers, and
// string allocation.
const programSynthetic = `
package test

func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func sieve(limit int) []bool {
	s := make([]bool, limit)
	for i := 2; i < limit; i++ {
		s[i] = true
	}
	for i := 2; i*i < limit; i++ {
		if s[i] {
			for j := i * i; j < limit; j += i {
				s[j] = false
			}
		}
	}
	return s
}

type Point struct {
	X, Y int
}

func (p Point) Dist() int {
	return p.X*p.X + p.Y*p.Y
}

func mapWork(n int) int {
	m := make(map[int]int)
	for i := 0; i < n; i++ {
		m[i] = i * i
	}
	sum := 0
	for _, v := range m {
		sum += v
	}
	return sum
}

func closureWork(n int) int {
	acc := 0
	add := func(x int) {
		acc += x
	}
	for i := 0; i < n; i++ {
		add(i)
	}
	return acc
}

func stringWork(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "a"
	}
	return s
}

// Interface dispatch
type Stringer interface {
	String() string
}

type Named struct {
	Name string
}

func (n Named) String() string {
	return n.Name
}

type Numbered struct {
	N int
}

func (n Numbered) String() string {
	if n.N == 0 {
		return "zero"
	}
	return "nonzero"
}

func interfaceWork(n int) int {
	items := make([]Stringer, n)
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			items[i] = Named{Name: "hello"}
		} else {
			items[i] = Numbered{N: i}
		}
	}
	count := 0
	for _, item := range items {
		if item.String() != "" {
			count++
		}
	}
	return count
}

// Type switch
func typeSwitch(vals []interface{}) int {
	sum := 0
	for _, v := range vals {
		switch v := v.(type) {
		case int:
			sum += v
		case string:
			sum += len(v)
		case bool:
			if v {
				sum++
			}
		}
	}
	return sum
}

// Pointer and nested struct access
type Inner struct {
	Value int
}

type Outer struct {
	A, B *Inner
	Name string
}

func pointerWork(n int) int {
	sum := 0
	for i := 0; i < n; i++ {
		o := Outer{
			A:    &Inner{Value: i},
			B:    &Inner{Value: i * 2},
			Name: "test",
		}
		sum += o.A.Value + o.B.Value
	}
	return sum
}

// Slice append and iteration
func sliceWork(n int) int {
	s := make([]int, 0)
	for i := 0; i < n; i++ {
		s = append(s, i*i)
	}
	sum := 0
	for _, v := range s {
		sum += v
	}
	return sum
}

func main() {
	// Fibonacci (heavy recursion + arithmetic)
	_ = fib(20)

	// Sieve of Eratosthenes (loops + slice ops + conditionals)
	_ = sieve(1000)

	// Struct method calls
	sum := 0
	for i := 0; i < 500; i++ {
		p := Point{X: i, Y: i + 1}
		sum += p.Dist()
	}
	_ = sum

	// Map operations
	_ = mapWork(500)

	// Closures
	_ = closureWork(1000)

	// String concatenation (allocation-heavy)
	_ = stringWork(200)

	// Interface dispatch
	_ = interfaceWork(200)

	// Type switches
	vals := make([]interface{}, 300)
	for i := 0; i < 300; i++ {
		switch i % 3 {
		case 0:
			vals[i] = i
		case 1:
			vals[i] = "hello"
		case 2:
			vals[i] = true
		}
	}
	_ = typeSwitch(vals)

	// Pointer and nested struct access
	_ = pointerWork(500)

	// Slice append
	_ = sliceWork(500)
}
`

// programContract simulates a simple token contract with balance tracking,
// transfers, and allowances — representative of real realm workloads.
const programContract = `
package test

type Token struct {
	balances   map[string]int
	allowances map[string]map[string]int
	totalSupply int
}

func NewToken(supply int) *Token {
	t := &Token{
		balances:   make(map[string]int),
		allowances: make(map[string]map[string]int),
	}
	t.balances["creator"] = supply
	t.totalSupply = supply
	return t
}

func (t *Token) BalanceOf(addr string) int {
	return t.balances[addr]
}

func (t *Token) Transfer(from, to string, amount int) bool {
	if t.balances[from] < amount {
		return false
	}
	t.balances[from] -= amount
	t.balances[to] += amount
	return true
}

func (t *Token) Approve(owner, spender string, amount int) {
	if t.allowances[owner] == nil {
		t.allowances[owner] = make(map[string]int)
	}
	t.allowances[owner][spender] = amount
}

func (t *Token) TransferFrom(spender, from, to string, amount int) bool {
	if t.allowances[from] == nil {
		return false
	}
	allowed := t.allowances[from][spender]
	if allowed < amount || t.balances[from] < amount {
		return false
	}
	t.allowances[from][spender] -= amount
	t.balances[from] -= amount
	t.balances[to] += amount
	return true
}

func main() {
	tok := NewToken(1000000)

	// Simulate 200 direct transfers
	addrs := []string{"alice", "bob", "carol", "dave", "eve"}
	for i := 0; i < 200; i++ {
		from := addrs[i % len(addrs)]
		to := addrs[(i+1) % len(addrs)]
		if tok.BalanceOf(from) > 0 {
			tok.Transfer(from, to, 10)
		}
	}

	// Set up allowances and do delegated transfers
	for _, addr := range addrs {
		tok.Approve("creator", addr, 5000)
	}
	for i := 0; i < 100; i++ {
		spender := addrs[i % len(addrs)]
		to := addrs[(i+2) % len(addrs)]
		tok.TransferFrom(spender, "creator", to, 50)
	}

	// Check balances
	total := 0
	total += tok.BalanceOf("creator")
	for _, addr := range addrs {
		total += tok.BalanceOf(addr)
	}
	if total != 1000000 {
		panic("total supply mismatch")
	}
}
`

func setupMachine(program string, gasMeter store.GasMeter) *gnolang.Machine {
	m := gnolang.NewMachineWithOptions(gnolang.MachineOptions{
		PkgPath:  "test",
		Output:   io.Discard,
		GasMeter: gasMeter,
	})
	nn := m.MustParseFile("main.go", program)
	m.RunFiles(nn)
	return m
}

func setupMachineWithStdlibs(program string, gasMeter store.GasMeter) *gnolang.Machine {
	rootDir, err := filepath.Abs("../../../")
	if err != nil {
		panic(err)
	}
	_, testStore := test.StoreWithOptions(
		rootDir, io.Discard,
		test.StoreOptions{Testing: true},
	)
	m := test.Machine(testStore, io.Discard, "test", false, gasMeter)
	nn := m.MustParseFile("main.go", program)
	// Set up a package for the test code (like filetest runner does).
	pn := gnolang.NewPackageNode("test", "test", &gnolang.FileSet{})
	pv := pn.NewPackage(m.Alloc)
	m.Store.SetBlockNode(pn)
	m.Store.SetCachePackage(pv)
	m.SetActivePackage(pv)
	m.RunFiles(nn)
	return m
}

// --- Synthetic benchmark (recursion, arithmetic, data structures) ---

func BenchmarkVM(b *testing.B) {
	m := setupMachine(programSynthetic, nil)
	m.RunMain()
	defer m.Release()

	b.ResetTimer()
	for range b.N {
		m.RunMain()
	}
}

func BenchmarkVM_GasMetered(b *testing.B) {
	m := setupMachine(programSynthetic, store.NewGasMeter(math.MaxInt64))
	m.RunMain()
	defer m.Release()

	b.ResetTimer()
	for range b.N {
		m.RunMain()
	}
}

// --- Contract benchmark (map-heavy, method calls, realistic realm workload) ---

func BenchmarkContract(b *testing.B) {
	m := setupMachine(programContract, nil)
	m.RunMain()
	defer m.Release()

	b.ResetTimer()
	for range b.N {
		m.RunMain()
	}
}

func BenchmarkContract_GasMetered(b *testing.B) {
	m := setupMachine(programContract, store.NewGasMeter(math.MaxInt64))
	m.RunMain()
	defer m.Release()

	b.ResetTimer()
	for range b.N {
		m.RunMain()
	}
}

// --- Stdlib benchmarks (use real stdlibs: strings, strconv, sort, bytes) ---

// programStringProcessing exercises strings, strconv, and bytes packages.
const programStringProcessing = `
package test

import (
	"strconv"
	"strings"
)

func main() {
	// strings.Builder usage
	var b strings.Builder
	for i := 0; i < 200; i++ {
		b.WriteString("item-")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(", ")
	}
	result := b.String()

	// String splitting and joining
	parts := strings.Split(result, ", ")
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.HasPrefix(p, "item-1") {
			filtered = append(filtered, strings.ToUpper(p))
		}
	}
	joined := strings.Join(filtered, "; ")
	_ = joined

	// String replacement and counting
	s := strings.Repeat("hello world ", 100)
	s = strings.ReplaceAll(s, "world", "gno")
	_ = strings.Count(s, "gno")

	// strconv conversions
	sum := 0
	for i := 0; i < 500; i++ {
		s := strconv.Itoa(i * 17)
		n, _ := strconv.Atoi(s)
		sum += n
	}
	_ = sum
}
`

// programSortSearch exercises sort package with various data patterns.
const programSortSearch = `
package test

import "sort"

type PairSlice []Pair
type Pair struct {
	Key   string
	Value int
}

func (s PairSlice) Len() int           { return len(s) }
func (s PairSlice) Less(i, j int) bool {
	if s[i].Key == s[j].Key {
		return s[i].Value < s[j].Value
	}
	return s[i].Key < s[j].Key
}
func (s PairSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func main() {
	// Sort integers
	nums := make(sort.IntSlice, 500)
	for i := range nums {
		nums[i] = (i * 97) % 503 // pseudo-random via modular arithmetic
	}
	sort.Sort(nums)

	// Verify sorted
	for i := 1; i < len(nums); i++ {
		if nums[i] < nums[i-1] {
			panic("not sorted")
		}
	}

	// Sort strings
	strs := make(sort.StringSlice, 200)
	for i := range strs {
		strs[i] = string(rune('A'+(i%26))) + string(rune('a'+(i*7%26))) + string(rune('a'+(i*13%26)))
	}
	sort.Sort(strs)

	// Binary search
	for i := 0; i < 100; i++ {
		target := (i * 97) % 503
		sort.SearchInts([]int(nums), target)
	}

	// Sort with custom comparator via sort.Interface
	pairs := make(PairSlice, 300)
	for i := range pairs {
		pairs[i] = Pair{
			Key:   string(rune('A'+(i%26))) + string(rune('a'+(i*3%26))),
			Value: (i * 53) % 1000,
		}
	}
	sort.Sort(pairs)
}
`

// programComplex exercises multiple stdlibs together in a realistic pattern:
// parsing structured data, transforming it, and producing output.
const programComplex = `
package test

import (
	"strconv"
	"strings"
	"sort"
)

type Record struct {
	Name  string
	Score int
	Tags  []string
}

func parseRecord(line string) Record {
	parts := strings.Split(line, "|")
	score, _ := strconv.Atoi(parts[1])
	tags := strings.Split(parts[2], ",")
	return Record{Name: parts[0], Score: score, Tags: tags}
}

func (r Record) HasTag(tag string) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

type ByScore []Record

func (s ByScore) Len() int           { return len(s) }
func (s ByScore) Less(i, j int) bool { return s[i].Score < s[j].Score }
func (s ByScore) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (r Record) String() string {
	var b strings.Builder
	b.WriteString(r.Name)
	b.WriteString("(")
	b.WriteString(strconv.Itoa(r.Score))
	b.WriteString(")")
	return b.String()
}

func main() {
	// Generate test data
	names := []string{"Alice", "Bob", "Carol", "Dave", "Eve", "Frank", "Grace", "Heidi"}
	tagSets := []string{"admin,user", "user", "admin,moderator", "user,viewer", "admin,user,moderator"}

	lines := make([]string, 0, 200)
	for i := 0; i < 200; i++ {
		name := names[i%len(names)]
		score := (i * 73 + 17) % 1000
		tags := tagSets[i%len(tagSets)]
		lines = append(lines, name+"|"+strconv.Itoa(score)+"|"+tags)
	}

	// Parse all records
	records := make([]Record, len(lines))
	for i, line := range lines {
		records[i] = parseRecord(line)
	}

	// Filter admins
	admins := make([]Record, 0)
	for _, r := range records {
		if r.HasTag("admin") {
			admins = append(admins, r)
		}
	}

	// Sort by score descending using sort.Interface
	sort.Sort(sort.Reverse(ByScore(admins)))

	// Build leaderboard string
	var lb strings.Builder
	lb.WriteString("=== Leaderboard ===\n")
	for i, r := range admins {
		if i >= 20 {
			break
		}
		lb.WriteString(strconv.Itoa(i+1))
		lb.WriteString(". ")
		lb.WriteString(r.String())
		lb.WriteString(" [")
		lb.WriteString(strings.Join(r.Tags, ", "))
		lb.WriteString("]\n")
	}
	result := lb.String()

	// Compute statistics
	totalScore := 0
	for _, r := range admins {
		totalScore += r.Score
	}
	avgScore := 0
	if len(admins) > 0 {
		avgScore = totalScore / len(admins)
	}
	_ = result
	_ = avgScore
}
`

func BenchmarkStdlibStrings(b *testing.B) {
	m := setupMachineWithStdlibs(programStringProcessing, nil)
	m.RunMain()
	defer m.Release()

	b.ResetTimer()
	for range b.N {
		m.RunMain()
	}
}

func BenchmarkStdlibStrings_GasMetered(b *testing.B) {
	m := setupMachineWithStdlibs(programStringProcessing, store.NewGasMeter(math.MaxInt64))
	m.RunMain()
	defer m.Release()

	b.ResetTimer()
	for range b.N {
		m.RunMain()
	}
}

func BenchmarkStdlibSort(b *testing.B) {
	m := setupMachineWithStdlibs(programSortSearch, nil)
	m.RunMain()
	defer m.Release()

	b.ResetTimer()
	for range b.N {
		m.RunMain()
	}
}

func BenchmarkStdlibSort_GasMetered(b *testing.B) {
	m := setupMachineWithStdlibs(programSortSearch, store.NewGasMeter(math.MaxInt64))
	m.RunMain()
	defer m.Release()

	b.ResetTimer()
	for range b.N {
		m.RunMain()
	}
}

func BenchmarkStdlibComplex(b *testing.B) {
	m := setupMachineWithStdlibs(programComplex, nil)
	m.RunMain()
	defer m.Release()

	b.ResetTimer()
	for range b.N {
		m.RunMain()
	}
}

func BenchmarkStdlibComplex_GasMetered(b *testing.B) {
	m := setupMachineWithStdlibs(programComplex, store.NewGasMeter(math.MaxInt64))
	m.RunMain()
	defer m.Release()

	b.ResetTimer()
	for range b.N {
		m.RunMain()
	}
}
