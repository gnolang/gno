package fuzzing

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Ent represents a structure used for analyzing string and []byte values.
// Strings are parsed into a tree structure using Ent as an entity.
type Ent interface {
	// setDepth assigns a depth value to the entity from a tree perspective.
	setDepth(d int)

	// setID assigns a unique ID to the entity from a tree perspective.
	setID(id int)
	// ID retrieves the assigned entity ID.
	ID() int

	// Returns a printable string representation of the fields in the XXXEnt structure.
	String() string
	// Score-related methods
	// Score retrieves the entity's score, which is used by the mutator to
	// determine the next entity to mutate.
	Score() uint64
	// SetScore assigns a score to the entity.
	SetScore(s uint64)
	PlusScore(delta int)

	// Mutation strategy counter methods
	// StrategyCount retrieves the score of a specific mutation strategy.
	// This helps the mutator decide the next strategy to apply.
	StrategyCount(code MutationCode) uint64
	// plusStrategyCount increments the score of a given mutation strategy.
	plusStrategyCount(code MutationCode, delta int)
	// InitStrategyCount initializes all strategy keys with a value of 1.
	InitStrategyCount()

	// Strategy map retrieval methods
	// Strategy2count retrieves the overall scores for each mutation strategy.
	Strategy2count() map[MutationCode]uint64

	// Caching methods for mutation strategies
	// StrategyCache retrieves an array of mutation strategies along with
	// their cumulative scores. The mutator uses this to select a strategy based
	// on its score distribution.
	StrategyCache() ([]MutationCode, []uint64)
	// updateStrategyCache updates the cumulative score array for mutation strategies.
	updateStrategyCache()
}

type PrimitiveEnt interface {
	Ent
	PrimitiveData() []byte
	SetPrimitiveData([]byte)
}

// LetterEnt represents an entity storing alphabetic data such as "apple" or "hello"
// from strings or byte slices.
type LetterEnt struct {
	data  []byte // Stores the string data.
	depth int    // Represents the depth of the entity in a tree structure.

	id int // Unique node ID assigned to each entity in the tree.

	score uint64 // Score associated with the entity.

	// Strategy usage log
	strategy2count map[MutationCode]uint64 // Stores scores for each applied mutation strategy.

	// Cache fields
	cachedStrategyCodes []MutationCode // Array of mutation codes.
	cachedStrategySums  []uint64       // Cumulative sum array mapping to mutation codes.
	needsUpdateChache   bool           // Tracks whether the cumulative sum array needs an update.
}

func (l *LetterEnt) PrimitiveData() []byte                 { return l.data }
func (l *LetterEnt) SetPrimitiveData(primitiveData []byte) { l.data = primitiveData }

// Implements Ent interface.
func (l *LetterEnt) setDepth(d int) { l.depth = d }

func (l *LetterEnt) setID(id int) { l.id = id }
func (l *LetterEnt) ID() int      { return l.id }

func (l *LetterEnt) String() string {
	return fmt.Sprintf("LetterEnt{Data:%q, Depth:%d, Id:%d, Score: %d}", l.data, l.depth, l.id, l.score)
}
func (l *LetterEnt) SetScore(s uint64)   { l.score = s }
func (l *LetterEnt) Score() uint64       { return l.score }
func (l *LetterEnt) PlusScore(delta int) { l.score += uint64(delta) }

func (l *LetterEnt) StrategyCount(code MutationCode) uint64 {
	if l.strategy2count == nil {
		l.InitStrategyCount()
	}
	if count, ok := l.strategy2count[code]; ok {
		return count
	}
	return 1
}

func (l *LetterEnt) plusStrategyCount(code MutationCode, delta int) {
	if l.strategy2count == nil {
		l.InitStrategyCount()
	}
	l.needsUpdateChache = true
	l.strategy2count[code] += uint64(delta)
}

func (l *LetterEnt) InitStrategyCount() {
	l.strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		l.strategy2count[code] = 1
	}
}

func (l *LetterEnt) Strategy2count() map[MutationCode]uint64 {
	return l.strategy2count
}

func (l *LetterEnt) StrategyCache() ([]MutationCode, []uint64) {
	// Updates the cache if it is not already initialized.

	if l.cachedStrategyCodes == nil || l.cachedStrategySums == nil {
		l.updateStrategyCache()
	}
	return l.cachedStrategyCodes, l.cachedStrategySums
}

func (l *LetterEnt) updateStrategyCache() {
	if !l.needsUpdateChache && l.cachedStrategyCodes != nil {
		return
	}
	if l.strategy2count == nil {
		l.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(l.strategy2count)
	l.cachedStrategyCodes = keys
	l.cachedStrategySums = cum
	l.needsUpdateChache = false
}

// buildCumulativeStrategyUint64 generates a cumulative sum slice (uint64)
// and a corresponding key slice from the Strategy2count map.
func buildCumulativeStrategyUint64(strategy map[MutationCode]uint64) ([]MutationCode, []uint64) {
	keys := make([]MutationCode, 0, len(strategy))
	for code := range strategy {
		keys = append(keys, code)
	}
	quickSortMutationCodes(keys)

	cum := make([]uint64, len(keys))
	var total uint64 = 0
	for i, code := range keys {
		total += strategy[code]
		cum[i] = total
	}
	return keys, cum
}

// quickSortMutationCodes sorts a slice of MutationCode values in ascending order
// using the Quicksort algorithm.
func quickSortMutationCodes(keys []MutationCode) {
	if len(keys) < 2 {
		return
	}
	pivotIndex := len(keys) / 2
	pivot := keys[pivotIndex]
	keys[pivotIndex], keys[len(keys)-1] = keys[len(keys)-1], keys[pivotIndex]

	storeIndex := 0
	for i := 0; i < len(keys)-1; i++ {
		if keys[i] < pivot {
			keys[i], keys[storeIndex] = keys[storeIndex], keys[i]
			storeIndex++
		}
	}
	keys[storeIndex], keys[len(keys)-1] = keys[len(keys)-1], keys[storeIndex]

	quickSortMutationCodes(keys[:storeIndex])
	quickSortMutationCodes(keys[storeIndex+1:])
}

// NumEnt represents an entity that stores numerical data extracted from
// strings or byte slices, such as "12", "-1", "0x12", or "1e13".
type NumEnt struct {
	data []byte
	Hex  bool // Hex indicates whether the data is represented in hexadecimal notation.

	depth int

	id    int
	score uint64

	strategy2count map[MutationCode]uint64

	cachedStrategyCodes []MutationCode
	cachedStrategySums  []uint64
	needsUpdateChache   bool
}

func (n *NumEnt) PrimitiveData() []byte                 { return n.data }
func (n *NumEnt) SetPrimitiveData(primitiveData []byte) { n.data = primitiveData }

// Implements Ent interface.
func (n *NumEnt) setDepth(d int) { n.depth = d }

func (n *NumEnt) setID(id int) { n.id = id }
func (n *NumEnt) ID() int      { return n.id }
func (n *NumEnt) String() string {
	return fmt.Sprintf("NumEnt{Data:%q, Hex:%v, Depth:%d, Id:%d, Score: %d}", n.data, n.Hex, n.depth, n.id, n.score)
}
func (n *NumEnt) SetScore(s uint64)   { n.score = s }
func (n *NumEnt) Score() uint64       { return n.score }
func (n *NumEnt) PlusScore(delta int) { n.score += uint64(delta) }

func (n *NumEnt) StrategyCount(code MutationCode) uint64 {
	if n.strategy2count == nil {
		n.InitStrategyCount()
	}
	if count, ok := n.strategy2count[code]; ok {
		return count
	}
	return 1
}

func (n *NumEnt) plusStrategyCount(code MutationCode, delta int) {
	if n.strategy2count == nil {
		n.InitStrategyCount()
	}
	n.needsUpdateChache = true
	n.strategy2count[code] += uint64(delta)
}

func (n *NumEnt) InitStrategyCount() {
	n.strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		n.strategy2count[code] = 1
	}
}

func (n *NumEnt) Strategy2count() map[MutationCode]uint64 {
	return n.strategy2count
}

func (n *NumEnt) StrategyCache() ([]MutationCode, []uint64) {
	if n.cachedStrategyCodes == nil || n.cachedStrategySums == nil {
		n.updateStrategyCache()
	}
	return n.cachedStrategyCodes, n.cachedStrategySums
}

func (n *NumEnt) updateStrategyCache() {
	if !n.needsUpdateChache && n.cachedStrategyCodes != nil {
		return
	}
	if n.Strategy2count == nil {
		n.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(n.strategy2count)
	n.cachedStrategyCodes = keys
	n.cachedStrategySums = cum
	n.needsUpdateChache = false
}

// WsEnt represents an entity that stores whitespace characters, such as " ",
// extracted from strings or byte slices.
type WsEnt struct {
	data  []byte
	depth int

	id    int
	score uint64

	strategy2count map[MutationCode]uint64

	cachedStrategyCodes []MutationCode
	cachedStrategySums  []uint64
	needsUpdateChache   bool
}

func (w *WsEnt) PrimitiveData() []byte                 { return w.data }
func (w *WsEnt) SetPrimitiveData(primitiveData []byte) { w.data = primitiveData }

// Implements Ent interface.
func (w *WsEnt) setDepth(d int) { w.depth = d }

func (w *WsEnt) setID(id int) { w.id = id }
func (w *WsEnt) ID() int      { return w.id }

func (w *WsEnt) String() string {
	return fmt.Sprintf("WsEnt{Data:%q, Depth:%d, Id:%d, Score: %d}", w.data, w.depth, w.id, w.score)
}
func (w *WsEnt) SetScore(s uint64)   { w.score = s }
func (w *WsEnt) Score() uint64       { return w.score }
func (w *WsEnt) PlusScore(delta int) { w.score += uint64(delta) }

func (w *WsEnt) StrategyCount(code MutationCode) uint64 {
	if w.strategy2count == nil {
		w.InitStrategyCount()
	}
	if count, ok := w.strategy2count[code]; ok {
		return count
	}
	return 1
}

func (w *WsEnt) plusStrategyCount(code MutationCode, delta int) {
	if w.strategy2count == nil {
		w.InitStrategyCount()
	}
	w.needsUpdateChache = true
	w.strategy2count[code] += uint64(delta)
}

func (w *WsEnt) InitStrategyCount() {
	w.strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		w.strategy2count[code] = 1
	}
}

func (w *WsEnt) Strategy2count() map[MutationCode]uint64 {
	return w.strategy2count
}

func (w *WsEnt) StrategyCache() ([]MutationCode, []uint64) {
	if w.cachedStrategyCodes == nil || w.cachedStrategySums == nil {
		w.updateStrategyCache()
	}
	return w.cachedStrategyCodes, w.cachedStrategySums
}

func (w *WsEnt) updateStrategyCache() {
	if !w.needsUpdateChache && w.cachedStrategyCodes != nil {
		return
	}
	if w.Strategy2count == nil {
		w.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(w.strategy2count)
	w.cachedStrategyCodes = keys
	w.cachedStrategySums = cum
	w.needsUpdateChache = false
}

// MarkEnt represents an entity that stores special values such as "!", "HTTP",
// or "AND", extracted from strings or byte slices.
type MarkEnt struct {
	data  []byte
	depth int

	id    int
	score uint64

	strategy2count map[MutationCode]uint64

	// build cache
	cachedStrategyCodes []MutationCode
	cachedStrategySums  []uint64
	needsUpdateChache   bool
}

func (m *MarkEnt) PrimitiveData() []byte                 { return m.data }
func (m *MarkEnt) SetPrimitiveData(primitiveData []byte) { m.data = primitiveData }

// Implements Ent interface.
func (m *MarkEnt) setDepth(d int) { m.depth = d }

func (m *MarkEnt) setID(id int) { m.id = id }
func (m *MarkEnt) ID() int      { return m.id }

func (m *MarkEnt) String() string {
	return fmt.Sprintf("MarkEnt{Data:%q, Depth:%d, Id:%d, Score: %d}", m.data, m.depth, m.id, m.score)
}
func (m *MarkEnt) SetScore(s uint64)   { m.score = s }
func (m *MarkEnt) Score() uint64       { return m.score }
func (m *MarkEnt) PlusScore(delta int) { m.score += uint64(delta) }

func (m *MarkEnt) StrategyCount(code MutationCode) uint64 {
	if m.strategy2count == nil {
		m.InitStrategyCount()
	}
	if count, ok := m.strategy2count[code]; ok {
		return count
	}
	return 1
}

func (m *MarkEnt) plusStrategyCount(code MutationCode, delta int) {
	if m.strategy2count == nil {
		m.InitStrategyCount()
	}
	m.needsUpdateChache = true
	m.strategy2count[code] += uint64(delta)
}

func (m *MarkEnt) InitStrategyCount() {
	m.strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		m.strategy2count[code] = 1
	}
}

func (m *MarkEnt) Strategy2count() map[MutationCode]uint64 {
	return m.strategy2count
}

func (m *MarkEnt) StrategyCache() ([]MutationCode, []uint64) {
	if m.cachedStrategyCodes == nil || m.cachedStrategySums == nil {
		m.updateStrategyCache()
	}
	return m.cachedStrategyCodes, m.cachedStrategySums
}

func (m *MarkEnt) updateStrategyCache() {
	if !m.needsUpdateChache && m.cachedStrategyCodes != nil {
		return
	}
	if m.strategy2count == nil {
		m.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(m.strategy2count)
	m.cachedStrategyCodes = keys
	m.cachedStrategySums = cum
	m.needsUpdateChache = false
}

// Composite Entities Begin

// KeyValEnt represents an entity that stores key-delimiter-value structures
// extracted from strings or byte slices, such as "a := 2" or "var = x".
type KeyValEnt struct {
	k         Ent // Represents the key in the key-delimiter-value structure.
	plOrNil   Ent // Represents the whitespace entity between the key and the delimiter.
	prOrNil   Ent // Represents the whitespace entity between the delimiter and the value.
	delimiter Ent // Represents the delimiter in the key-delimiter-value structure.
	v         Ent // Represents the value in the key-delimiter-value structure.
	depth     int
	id        int
	score     uint64

	strategy2count map[MutationCode]uint64

	cachedStrategyCodes []MutationCode
	cachedStrategySums  []uint64
	needsUpdateChache   bool
}

// Implements Ent interface.
func (kv *KeyValEnt) setDepth(d int) { kv.depth = d }

func (kv *KeyValEnt) setID(id int) { kv.id = id }
func (kv *KeyValEnt) ID() int      { return kv.id }
func (kv *KeyValEnt) String() string {
	return fmt.Sprintf("KeyValEnt{Depth:%d, Id:%d, K=%v, Delim=%v, V=%v, Score: %d}",
		kv.depth, kv.id, kv.k, kv.delimiter, kv.v, kv.score)
}
func (kv *KeyValEnt) SetScore(s uint64)   { kv.score = s }
func (kv *KeyValEnt) Score() uint64       { return kv.score }
func (kv *KeyValEnt) PlusScore(delta int) { kv.score += uint64(delta) }

func (kv *KeyValEnt) StrategyCount(code MutationCode) uint64 {
	if kv.strategy2count == nil {
		kv.InitStrategyCount()
	}
	if count, ok := kv.strategy2count[code]; ok {
		return count
	}
	return 1
}

func (kv *KeyValEnt) plusStrategyCount(code MutationCode, delta int) {
	if kv.strategy2count == nil {
		kv.InitStrategyCount()
	}
	kv.needsUpdateChache = true
	kv.strategy2count[code] += uint64(delta)
}

func (kv *KeyValEnt) InitStrategyCount() {
	kv.strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		kv.strategy2count[code] = 1
	}
}

func (kv *KeyValEnt) Strategy2count() map[MutationCode]uint64 {
	return kv.strategy2count
}

func (kv *KeyValEnt) StrategyCache() ([]MutationCode, []uint64) {
	if kv.cachedStrategyCodes == nil || kv.cachedStrategySums == nil {
		kv.updateStrategyCache()
	}
	return kv.cachedStrategyCodes, kv.cachedStrategySums
}

func (kv *KeyValEnt) updateStrategyCache() {
	if !kv.needsUpdateChache && kv.cachedStrategyCodes != nil {
		return
	}
	if kv.strategy2count == nil {
		kv.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(kv.strategy2count)
	kv.cachedStrategyCodes = keys
	kv.cachedStrategySums = cum
	kv.needsUpdateChache = false
}

// GroupEnt represents an entity that stores grouped structures extracted from
// strings or byte slices, such as "(a)" or "[1,2,3]".
type GroupEnt struct {
	data      []Ent // Stores entities grouped together in an array.
	depth     int
	IsPattern bool // Indicates whether the group follows a specific pattern.

	id    int
	score uint64

	strategy2count map[MutationCode]uint64

	cachedStrategyCodes []MutationCode
	cachedStrategySums  []uint64
	needsUpdateChache   bool
}

// Implements Ent interface.
func (g *GroupEnt) String() string {
	return fmt.Sprintf("GroupEnt{Len:%d, Depth:%d, Id:%d, IsPattern:%v, Score: %d}",
		len(g.data), g.depth, g.id, g.IsPattern, g.score)
}
func (g *GroupEnt) setDepth(d int) { g.depth = d }

func (g *GroupEnt) setID(id int)        { g.id = id }
func (g *GroupEnt) ID() int             { return g.id }
func (g *GroupEnt) SetScore(s uint64)   { g.score = s }
func (g *GroupEnt) Score() uint64       { return g.score }
func (g *GroupEnt) PlusScore(delta int) { g.score += uint64(delta) }

func (g *GroupEnt) StrategyCount(code MutationCode) uint64 {
	if g.strategy2count == nil {
		g.InitStrategyCount()
	}
	if count, ok := g.strategy2count[code]; ok {
		return count
	}
	return 1
}

func (g *GroupEnt) plusStrategyCount(code MutationCode, delta int) {
	if g.strategy2count == nil {
		g.InitStrategyCount()
	}
	g.needsUpdateChache = true
	g.strategy2count[code] += uint64(delta)
}

func (g *GroupEnt) InitStrategyCount() {
	g.strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		g.strategy2count[code] = 1
	}
}

func (g *GroupEnt) Strategy2count() map[MutationCode]uint64 {
	return g.strategy2count
}

func (g *GroupEnt) StrategyCache() ([]MutationCode, []uint64) {
	if g.cachedStrategyCodes == nil || g.cachedStrategySums == nil {
		g.updateStrategyCache()
	}
	return g.cachedStrategyCodes, g.cachedStrategySums
}

func (g *GroupEnt) updateStrategyCache() {
	if !g.needsUpdateChache && g.cachedStrategyCodes != nil {
		return
	}
	if g.strategy2count == nil {
		g.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(g.strategy2count)
	g.cachedStrategyCodes = keys
	g.cachedStrategySums = cum
	g.needsUpdateChache = false
}

// RootEnt represents the root entity that stores, manages, and analyzes
// an entire string or byte sequence.
type RootEnt struct {
	data  []Ent // Stores all entities within the root.
	depth int

	IsPattern bool   // Indicates whether the string follows a specific pattern.
	Pattern   string // Stores the expected pattern of the string.

	id int // Unique ID of the root entity. (it's value is 0)

	Id2Ent map[int]Ent // Maps each contained entity's ID for easy access.
	score  uint64

	strategy2count map[MutationCode]uint64

	// Cache fields
	cachedStrategyCodes       []MutationCode
	cachedStrategySums        []uint64
	needsUpdateStrategyChache bool // Tracks if the strategy cache needs updating.

	// Score-based cache fields
	cachedScoreIds         []int    // Caches an array of entity IDs.
	cachedScoreCums        []uint64 // Caches the cumulative score array for entity IDs.
	needsUpdateScoreChache bool     // Tracks if the ID-score cache needs updating.
}

// Implements Ent interface.
func (r *RootEnt) setDepth(d int) { r.depth = d }

func (r *RootEnt) setID(id int) { r.id = id }
func (r *RootEnt) ID() int      { return r.id }

func (r *RootEnt) String() string {
	return fmt.Sprintf("RootEnt{Len:%d, Depth:%d, Pattern:%q, Id:%d, Score: %d}",
		len(r.data), r.depth, r.Pattern, r.id, r.score)
}
func (r *RootEnt) SetScore(s uint64)   { r.score = s }
func (r *RootEnt) Score() uint64       { return r.score }
func (r *RootEnt) PlusScore(delta int) { r.score += uint64(delta) }

func (r *RootEnt) InitStrategyCount() {
	r.strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		r.strategy2count[code] = 1
	}
}

func (r *RootEnt) StrategyCount(code MutationCode) uint64 {
	if r.strategy2count == nil {
		r.InitStrategyCount()
	}
	if count, ok := r.strategy2count[code]; ok {
		return count
	}
	return 1
}

func (r *RootEnt) plusStrategyCount(code MutationCode, delta int) {
	if r.strategy2count == nil {
		r.InitStrategyCount()
	}
	r.needsUpdateStrategyChache = true
	r.strategy2count[code] += uint64(delta)
}

func (r *RootEnt) Strategy2count() map[MutationCode]uint64 {
	return r.strategy2count
}

func (r *RootEnt) StrategyCache() ([]MutationCode, []uint64) {
	if r.cachedStrategyCodes == nil || r.cachedStrategySums == nil {
		r.updateStrategyCache()
	}
	return r.cachedStrategyCodes, r.cachedStrategySums
}

func (r *RootEnt) updateStrategyCache() {
	if !r.needsUpdateStrategyChache && r.cachedStrategyCodes != nil {
		return
	}
	if r.strategy2count == nil {
		r.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(r.strategy2count)
	r.cachedStrategyCodes = keys
	r.cachedStrategySums = cum
	r.needsUpdateStrategyChache = false
}

// ReflectMutationInfo updates mutation strategy counts and scores for a mutated entity.
// If mutInfo.IsMutated is true, it retrieves the entity using EntId from RootEnt's Id2Ent map,
// increments the strategy count for each recorded strategy in UsedStrategyMap by 5 per usage,
// and increases the entity's score by 5.
//
// TODO: The increment value (currently 5) could be dynamic. Improve by adjusting
// score increments based on the overall context of the cluster
func (r *RootEnt) ReflectMutationInfo(mutInfo MutationInfo) {
	if !mutInfo.IsMutated || !mutInfo.IsStringOrByteArray {
		return
	}

	ent, exists := r.Id2Ent[mutInfo.EntId]
	if !exists {
		panic("No EntID satisfying reflect exists.")
	}

	// Calls plusStrategyCount for each strategy in UsedStrategyMap, increasing the count by 5 for each usage.
	for code, count := range mutInfo.UsedStrategyMap {
		ent.plusStrategyCount(code, 5*count)
	}

	// Also increments the Score of the entity by 5.
	ent.PlusScore(5)
}

// ScoreById retrieves the score of an entity with the specified ID.
func (r *RootEnt) ScoreById(id int) uint64 {
	if ent, exists := r.Id2Ent[id]; exists {
		return ent.Score()
	}
	return 0
}

// SetScoreById sets the score for an entity with the specified ID.
func (r *RootEnt) SetScoreById(id int, score uint64) {
	if ent, exists := r.Id2Ent[id]; exists {
		r.needsUpdateScoreChache = true
		ent.SetScore(score)
	}
}

// PlusScoreById increases the score of an entity with the specified ID by delta.
func (r *RootEnt) PlusScoreById(id int, delta int) {
	if ent, exists := r.Id2Ent[id]; exists {
		r.needsUpdateScoreChache = true
		ent.PlusScore(delta)
	}
}

// updateScoreCache computes and stores cumulative scores from RootEnt's Id2Ent map.
// It gathers all node IDs, sorts them in ascending order, and calculates the cumulative sum.
func (r *RootEnt) updateScoreCache() {
	if !r.needsUpdateScoreChache && r.cachedScoreIds != nil {
		return
	}
	// Collects all IDs.
	keys := make([]int, 0, len(r.Id2Ent))
	for id := range r.Id2Ent {
		keys = append(keys, id)
	}
	// Sorts them in ascending order (used for binary search via sort.Search).
	sort.Ints(keys)

	cum := make([]uint64, len(keys))
	var total uint64 = 0
	for i, id := range keys {
		// Converts Score() return value from int to uint64.
		score := uint64(r.Id2Ent[id].Score())
		total += score
		cum[i] = total
	}
	r.cachedScoreIds = keys
	r.cachedScoreCums = cum
	r.needsUpdateScoreChache = false
}

// ScoreCache returns cached score-related slices.
// Updates the cache if it is uninitialized.
func (r *RootEnt) ScoreCache() ([]int, []uint64) {
	return r.cachedScoreIds, r.cachedScoreCums
}

// ParseToRoot parses a []byte input into a RootEnt structure.
func ParseToRoot(b []byte) *RootEnt {
	// 1) Tokenizes and lexes the input string.
	lx := NewLexer([]byte(b))
	toks := lx.Lex()
	// 2) Parses key-value structures into KeyVal entities.
	toks2 := parseKeyVals(toks)
	// 3) Parses bracketed expressions into Group entities.
	toks3 := parseBrackets(toks2)
	// 4) Constructs the RootEnt entity.
	r := &RootEnt{data: toks3, depth: 0, needsUpdateStrategyChache: true, needsUpdateScoreChache: true}
	// 5) Infers patterns within the root entity.
	r = applyPatternToRoot(r)
	// 6) Assigns depth values to each entity.
	finalizeDepth(r, 0)
	// 7) Assigns unique IDs to each entity.
	finalizeId(r)
	// 8) Initializes the score of each entity to 1.
	initializeScore(r)
	// 9) Initializes the strategy-score dictionary for each entity with a score of 1.
	initializeStrategyCount(r)

	return r
}

// Lexer preserves the original input as a []byte sequence.
type Lexer struct {
	input []byte
	pos   int
}

// NewLexer creates a new Lexer instance from a []byte input.
func NewLexer(input []byte) *Lexer {
	return &Lexer{input: input, pos: 0}
}

// eof checks if the lexer has reached the end of input.
func (lx *Lexer) eof() bool {
	return lx.pos >= len(lx.input)
}

// peek returns the []byte representation of the UTF-8 encoded character at the current position.
func (lx *Lexer) peek() []byte {
	if lx.eof() {
		return nil
	}
	// Decodes UTF-8 from the current position.
	_, size := utf8.DecodeRune(lx.input[lx.pos:])
	return lx.input[lx.pos : lx.pos+size]
}

// next returns the []byte representation of the UTF-8 encoded character at the current position
// and advances the internal position by the character's byte length.
func (lx *Lexer) next() []byte {
	if lx.eof() {
		return nil
	}
	_, size := utf8.DecodeRune(lx.input[lx.pos:])
	b := lx.input[lx.pos : lx.pos+size]
	lx.pos += size
	return b
}

// matchString compares the given string with the byte sequence starting at the current position.
func (lx *Lexer) matchString(s string) bool {
	bs := []byte(s)
	if lx.pos+len(bs) > len(lx.input) {
		return false
	}
	for i, b := range bs {
		if lx.input[lx.pos+i] != b {
			return false
		}
	}
	return true
}

// sortMarksByLengthDesc sorts a slice of marker strings in descending order by length.
func sortMarksByLengthDesc(marks []string) []string {
	cp := make([]string, len(marks))
	copy(cp, marks)
	for i := 0; i < len(cp)-1; i++ {
		for j := 0; j < len(cp)-1-i; j++ {
			if len(cp[j]) < len(cp[j+1]) {
				cp[j], cp[j+1] = cp[j+1], cp[j]
			}
		}
	}
	return cp
}

// Lex scans the entire []byte input and returns a list of tokens (Ent).
func (lx *Lexer) Lex() []Ent {
	var tokens []Ent
	for !lx.eof() {
		t := lx.nextToken()
		if t != nil {
			t.setDepth(0) // Initialize depth to 0 during the lexing stage.
			tokens = append(tokens, t)
		}
	}
	return tokens
}

// Global settings (mark list, bracket)
// ParserMark is sorted in descending order of length by the init() function.
var ParserMark = []string{
	"\r\n", "\n", "\t",
	"/", "://", "//",
	".", "?", "@", "!", "#", "$", "%", "^",
	"&&", "&", "||", "|",
	"or", "OR", "and", "AND",
	",", ";",
	":=", "=", ":",
}

// nextToken extracts a single token (Ent) from the input stream.
func (lx *Lexer) nextToken() Ent {
	// Retrieves the []byte at the current position and decodes it as a UTF-8 character.
	pb := lx.peek()
	if pb == nil {
		return nil
	}
	r, _ := utf8.DecodeRune(pb)

	// 1) Handles newline characters.
	if r == '\n' {
		lx.next()
		return &MarkEnt{data: []byte("\n")}
	}

	// 2) Detects and groups whitespace characters (excluding newlines).
	if unicode.IsSpace(r) {
		var sb strings.Builder
		for !lx.eof() {
			pb = lx.peek()
			if pb == nil {
				break
			}
			r, _ = utf8.DecodeRune(pb)
			if !unicode.IsSpace(r) || r == '\n' {
				break
			}
			sb.WriteRune(r)
			lx.next()
		}
		return &WsEnt{data: []byte(sb.String())}
	}

	// 3) Parses numbers, including negative, decimal, hexadecimal, and scientific notation.
	if unicode.IsDigit(r) || r == '-' {
		return lx.lexNumber()
	}

	// 4) Extracts alphabetic sequences (letters, digits, and underscores).
	if unicode.IsLetter(r) {
		var sb strings.Builder
		for !lx.eof() {
			pb = lx.peek()
			if pb == nil {
				break
			}
			r, _ = utf8.DecodeRune(pb)
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				break
			}
			sb.WriteRune(r)
			lx.next()
		}
		return &LetterEnt{data: []byte(sb.String())}
	}

	// 5) Identifies multi-character markers.
	for _, mk := range ParserMark {
		if lx.matchString(mk) {
			lx.pos += len([]byte(mk))
			return &MarkEnt{data: []byte(mk)}
		}
	}

	// 6) Defaults to treating unrecognized single characters as markers.
	b := lx.next()
	return &MarkEnt{data: b}
}

// lexNumber parses numeric values, including decimal, hexadecimal,
// and scientific notation, and returns a NumEnt entity.
func (lx *Lexer) lexNumber() Ent {
	sign := ""
	// Handles negative sign if present.
	pb := lx.peek()
	if pb != nil {
		r, _ := utf8.DecodeRune(pb)
		if r == '-' {
			sign = "-"
			lx.next()
		}
	}
	// Processes hexadecimal numbers: If the number starts with "0x" or "0X"
	pb = lx.peek()
	if pb != nil {
		r, _ := utf8.DecodeRune(pb)
		// consumes '0' first, then consumes 'x' or 'X'.
		if r == '0' {
			if lx.pos+1 < len(lx.input) {
				nb := lx.input[lx.pos+1:]
				nr, _ := utf8.DecodeRune(nb)
				if nr == 'x' || nr == 'X' {
					lx.next() // consume '0'
					lx.next() // consume 'x' or 'X'

					var hexBuilder strings.Builder
					for !lx.eof() {
						pb = lx.peek()
						if pb == nil {
							break
						}
						r, _ = utf8.DecodeRune(pb)
						if isHexDigit[r] {
							hexBuilder.WriteRune(r)

							lx.next()
						} else {
							break
						}
					}
					hexDigits := hexBuilder.String()
					if len(hexDigits) == 0 {
						return &NumEnt{
							data: []byte(sign + "0" + string(nr)),
							Hex:  false,
						}
					}

					return &NumEnt{
						data: []byte(sign + "0" + string(nr) + hexDigits),
						Hex:  true,
					}
				}
			}
		}
	}
	// Processes decimal numbers.
	digits := ""
	for !lx.eof() {
		pb = lx.peek()
		if pb == nil {
			break
		}
		r, _ := utf8.DecodeRune(pb)
		if !unicode.IsDigit(r) {
			break
		}
		digits += string(r)
		lx.next()
	}
	if digits == "" {
		return &MarkEnt{data: []byte(sign)}
	}
	// Handles scientific notation: If 'e' or 'E' is encountered,
	expo := ""
	pb = lx.peek()
	if pb != nil {
		r, _ := utf8.DecodeRune(pb)
		if r == 'e' || r == 'E' {
			expLetter := string(lx.next()) // it is used directly as part of the exponent representation.
			expSign := ""
			pb = lx.peek()
			if pb != nil {
				r, _ = utf8.DecodeRune(pb)
				if r == '+' || r == '-' {
					expSign = string(lx.next())
				}
			}
			expDigits := ""
			for !lx.eof() {
				pb = lx.peek()
				if pb == nil {
					break
				}
				r, _ = utf8.DecodeRune(pb)
				if !unicode.IsDigit(r) {
					break
				}
				expDigits += string(r)
				lx.next()
			}
			if expDigits == "" {
				return &NumEnt{data: []byte(sign + digits + expLetter + expSign), Hex: false}
			}
			expo = expLetter + expSign + expDigits
		}
	}
	return &NumEnt{data: []byte(sign + digits + expo), Hex: false}
}

// parseKeyVals identifies key-delimiter-value structures in the token stream.
// It detects and extracts key-value pairs, including optional whitespace
// surrounding delimiters like ":", "=", and ":=".
func parseKeyVals(tokens []Ent) []Ent {
	var result []Ent
	i := 0
	for i < len(tokens) {
		// 1) Identifies a potential key (letter or number).
		left := tokens[i]
		if !isAlphaOrNum(left) {
			result = append(result, left)
			i++
			continue
		}
		pos := i + 1
		// 2) Optionally captures whitespace between the key and delimiter.
		var leftPadding Ent
		if pos < len(tokens) {
			if w, ok := tokens[pos].(*WsEnt); ok {
				leftPadding = w
				pos++
			}
		}

		// 3) Checks if the next token is a valid delimiter (":", "=", or ":=").
		if pos >= len(tokens) {
			// No delimiter position found => failure.
			result = append(result, left)
			if leftPadding != nil {
				result = append(result, leftPadding)
				// Consume left + padding.
				i += 2
			} else {
				i += 1
			}
			continue
		}

		delimiter := tokens[pos]
		md, ok2 := delimiter.(*MarkEnt)
		if !ok2 || !(bytes.Equal(md.data, []byte(":")) || bytes.Equal(md.data, []byte("=")) || bytes.Equal(md.data, []byte(":="))) {
			// Not a delimiter => failure.
			result = append(result, left)
			if leftPadding != nil {
				result = append(result, leftPadding)
				i += 2
			} else {
				i += 1
			}
			continue
		}
		pos++ // Consume delimiter.

		// 4) Optionally captures whitespace between the delimiter and the value.
		var rightPadding Ent
		if pos < len(tokens) {
			if w2, ok3 := tokens[pos].(*WsEnt); ok3 {
				rightPadding = w2
				pos++
			}
		}

		if pos >= len(tokens) {
			// No space for value => failure.
			// => left + leftPadding + delimiter + rightPadding
			result = append(result, left)
			if leftPadding != nil {
				result = append(result, leftPadding)
			}
			result = append(result, delimiter)
			if rightPadding != nil {
				result = append(result, rightPadding)
			}
			// Consumed `pos` characters so far.

			i = pos
			continue
		}

		// 5) Identifies a potential value (letter or number).
		right := tokens[pos]
		if !isAlphaOrNum(right) {
			// If not a value candidate => failure

			result = append(result, left)
			if leftPadding != nil {
				result = append(result, leftPadding)
			}
			result = append(result, delimiter)
			if rightPadding != nil {
				result = append(result, rightPadding)
			}
			// i = pos => Consumed up to delimiter/padding.

			i = pos
			continue
		}
		// Success => Create keyValEnt.

		kv := &KeyValEnt{
			depth:     0,
			k:         left,
			plOrNil:   leftPadding,
			delimiter: delimiter,
			prOrNil:   rightPadding,
			v:         right,
		}
		result = append(result, kv)

		// Consume key, left padding, delimiter, right padding, and value.

		i = pos + 1
	} // end for
	return result
}

// isAlphaOrNum checks whether the Ent is either a *letterEnt or a *NumEnt.
func isAlphaOrNum(e Ent) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt:
		return true
	}
	return false
}

// parseBrackets recursively parses bracketed expressions and groups entities.
func parseBrackets(tokens []Ent) []Ent {
	var result []Ent
	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		if mk, ok := tok.(*MarkEnt); ok && len(mk.data) == 1 {
			openRune := rune(mk.data[0])
			if _, exist := bracketPairs[openRune]; exist {
				g, consumed := parseBracketGroup(tokens, i)
				result = append(result, g)
				i = consumed
				continue
			}
		}
		result = append(result, tok)
		i++
	}
	return result
}

// Characters used to recognize brackets.
// Characters like "<" are recognized as special symbols, not as brackets.
var bracketPairs = map[rune]rune{
	'(': ')',
	'{': '}',
	'[': ']',
}

// parseBracketGroup recursively parses a bracketed group starting from the given index.
func parseBracketGroup(tokens []Ent, from int) (*GroupEnt, int) {
	open, _ := tokens[from].(*MarkEnt)
	closeRune := bracketPairs[rune(open.data[0])]
	g := &GroupEnt{
		depth: 0,
		data:  []Ent{open},
	}
	i := from + 1
	for i < len(tokens) {
		tok := tokens[i]
		if mk, ok := tok.(*MarkEnt); ok && len(mk.data) == 1 {
			if rune(mk.data[0]) == closeRune {
				g.data = append(g.data, mk)
				return g, i + 1
			}
			// Handle nested parentheses.
			if _, has := bracketPairs[rune(mk.data[0])]; has {
				sub, consumed := parseBracketGroup(tokens, i)
				g.data = append(g.data, sub)
				i = consumed
				continue
			}
		}
		g.data = append(g.data, tok)
		i++
	}
	return g, i
}

// applyPatternToRoot assigns a pattern to the root entity.
func applyPatternToRoot(r *RootEnt) *RootEnt {
	p, ok := detectPattern(r.data)
	if ok {
		r.IsPattern = true
		r.Pattern = p
	}
	return r
}

func detectPattern(tokens []Ent) (string, bool) {
	flat := flattenTokens(tokens)
	total := 0
	httpCount := 0
	gnoCount := 0
	mlCount := 0
	jsonBrace := 0
	jsonColon := 0

	for _, e := range flat {
		switch t := e.(type) {
		case *WsEnt:
			continue
		case *LetterEnt:
			total++
			low := bytes.ToLower(t.data)
			if bytes.Equal(low, []byte("return")) || bytes.Equal(low, []byte("func")) ||
				bytes.Equal(low, []byte("for")) || bytes.Equal(low, []byte("if")) ||
				bytes.Equal(low, []byte("case")) || bytes.Equal(low, []byte("switch")) ||
				bytes.Equal(low, []byte("int")) {
				gnoCount++
			}
			if bytes.Equal(low, []byte("http")) || bytes.Equal(low, []byte("https")) {
				httpCount++
			}
		case *MarkEnt:
			total++
			d := bytes.ToLower(t.data)
			if bytes.Equal(d, []byte(":=")) || bytes.Equal(d, []byte("=")) {
				gnoCount++
			}
			if bytes.Equal(d, []byte("://")) {
				httpCount++
			}
			if bytes.Equal(d, []byte("<")) || bytes.Equal(d, []byte(">")) {
				mlCount++
			}
			if bytes.Equal(d, []byte("{")) || bytes.Equal(d, []byte("}")) {
				jsonBrace++
			}
			if bytes.Equal(d, []byte(":")) {
				jsonColon++
			}
		case *KeyValEnt:
			total++
			d := t.delimiter
			var dStr []byte
			if delim, ok := d.(*MarkEnt); ok {
				dStr = delim.data
			} else {
				continue
			}

			if bytes.Equal(dStr, []byte(":=")) || bytes.Equal(dStr, []byte("=")) {
				gnoCount++
			} else {
				jsonColon++
			}

		}
	}
	if total == 0 {
		return "NONE", true
	}
	if gnoCount > 3 || float64(gnoCount)/float64(total) >= 0.1 {
		return "GNO", true
	}
	if httpCount > 1 {
		return "HTTP", true
	}
	if (jsonBrace + jsonColon) > 2 {
		return "JSON", true
	}
	if mlCount > 7 || float64(mlCount)/float64(total) >= 0.1 {
		return "ML", true
	}
	return "NONE", true
}

// flattenTokens recursively expands the contents inside groupEnt.
func flattenTokens(tokens []Ent) []Ent {
	var out []Ent
	for _, e := range tokens {
		switch gg := e.(type) {
		case *GroupEnt:
			if len(gg.data) > 0 {
				out = append(out, gg.data[0]) // open bracket
				if len(gg.data) > 2 {
					body := gg.data[1 : len(gg.data)-1]
					out = append(out, flattenTokens(body)...)
				}
				out = append(out, gg.data[len(gg.data)-1]) // close bracket
			}
		default:
			out = append(out, e)
		}
	}
	return out
}

// finalizeDepth assigns depth values to entities recursively.
func finalizeDepth(e Ent, currentDepth int) {
	// - Sets the given entity's depth to currentDepth.
	e.setDepth(currentDepth)
	// - If the entity has children, assigns them a depth of (currentDepth + 1).
	switch ent := e.(type) {
	case *KeyValEnt:
		// - Applies depth propagation for composite entities like KeyValEnt, GroupEnt, and RootEnt.
		finalizeDepth(ent.k, currentDepth+1)
		if ent.plOrNil != nil {
			finalizeDepth(ent.plOrNil, currentDepth+1)
		}
		finalizeDepth(ent.delimiter, currentDepth+1)
		if ent.prOrNil != nil {
			finalizeDepth(ent.prOrNil, currentDepth+1)
		}
		finalizeDepth(ent.v, currentDepth+1)
	case *GroupEnt:
		// - Applies depth propagation for composite entities like KeyValEnt, GroupEnt, and RootEnt.
		for _, c := range ent.data {
			finalizeDepth(c, currentDepth+1)
		}
	case *RootEnt:
		// - Applies depth propagation for composite entities like KeyValEnt, GroupEnt, and RootEnt.
		for _, c := range ent.data {
			finalizeDepth(c, currentDepth+1)
		}
	}
	// alphaEnt, numEnt, wsEnt, and markEnt have no children.
}

// finalizeId assigns unique IDs to all entities within RootEnt
// and stores them in RootEnt.Id2Ent.
func finalizeId(root *RootEnt) {
	root.Id2Ent = make(map[int]Ent)
	currentID := 0
	assignId(root, &currentID, root)
}

// assignId recursively assigns an ID to each entity and stores it
// in root.Id2Ent[currentID] = e.
func assignId(e Ent, currentID *int, root *RootEnt) {
	// 1) Assign an ID to e
	e.setID(*currentID)
	root.Id2Ent[*currentID] = e
	*currentID++

	// 2) Process child nodes (branching based on structure type)
	switch node := e.(type) {
	case *KeyValEnt:
		if node.k != nil {
			assignId(node.k, currentID, root)
		}
		if node.plOrNil != nil {
			assignId(node.plOrNil, currentID, root)
		}
		if node.delimiter != nil {
			assignId(node.delimiter, currentID, root)
		}
		if node.prOrNil != nil {
			assignId(node.prOrNil, currentID, root)
		}
		if node.v != nil {
			assignId(node.v, currentID, root)
		}

	case *GroupEnt:
		for _, c := range node.data {
			assignId(c, currentID, root)
		}

	case *RootEnt:
		for _, c := range node.data {
			assignId(c, currentID, root)
		}

	//    letterEnt, numEnt, wsEnt, and markEnt have no children.
	default:
		// no children
	}
}

// initializeScore recursively traverse the ent and initialize the score to 1.
func initializeScore(e Ent) {
	// Set e's score to 1
	e.SetScore(1)

	// If there are child elements, recursively finalizeScore
	switch ent := e.(type) {

	case *KeyValEnt:
		if ent.k != nil {
			initializeScore(ent.k)
		}
		if ent.plOrNil != nil {
			initializeScore(ent.plOrNil)
		}
		if ent.delimiter != nil {
			initializeScore(ent.delimiter)
		}
		if ent.prOrNil != nil {
			initializeScore(ent.prOrNil)
		}
		if ent.v != nil {
			initializeScore(ent.v)
		}

	case *GroupEnt:
		for _, c := range ent.data {
			initializeScore(c)
		}

	case *RootEnt:
		for _, c := range ent.data {
			initializeScore(c)
		}

	// LetterEnt, NumEnt, WsEnt, and MarkEnt have no children.
	default:
		// no children
	}
}

// initializeStrategyCount traverses all Ent nodes in the tree and
// initializes their strategy count by calling InitStrategyCount().
func initializeStrategyCount(e Ent) {
	// Calls initStrategyCount() for each node.
	e.InitStrategyCount()

	// Recursively calls for child nodes if they exist.
	switch node := e.(type) {
	case *KeyValEnt:
		if node.k != nil {
			initializeStrategyCount(node.k)
		}
		if node.plOrNil != nil {
			initializeStrategyCount(node.plOrNil)
		}
		if node.delimiter != nil {
			initializeStrategyCount(node.delimiter)
		}
		if node.prOrNil != nil {
			initializeStrategyCount(node.prOrNil)
		}
		if node.v != nil {
			initializeStrategyCount(node.v)
		}
	case *GroupEnt:
		for _, child := range node.data {
			initializeStrategyCount(child)
		}
	case *RootEnt:
		for _, child := range node.data {
			initializeStrategyCount(child)
		}
	default:
	}
}

// SliceWithOutID removes an entity with the given ID (including its subtree)
// and splits the original string into two parts at the exclusion point.
func SliceWithOutID(excludeID int, r *RootEnt) (string, string) {
	// 1. Find the Ent to exclude.
	excludedEnt, ok := r.Id2Ent[excludeID]
	// If the given id does not exist, return the entire entity as a single element.
	if !ok {
		return SerializeToString(r), ""
	}

	// 2. Collect the ids of the Ent to be excluded along with its descendants.
	excludedIDs := make(map[int]bool)
	collectIDs(excludedEnt, excludedIDs)

	// 3. Tokenize the entire RootEnt.
	tokens := SerializeToTokens(r)
	firstIdx, lastIdx := -1, -1
	for i, token := range tokens {
		if excludedIDs[token.ID()] {
			if firstIdx == -1 {
				firstIdx = i
			}
			lastIdx = i
		}
	}
	// 4. If there are no tokens to exclude, return the entire string.
	if firstIdx == -1 {
		return SerializeToString(r), ""
	}

	// 5. Concatenate tokens before and after the excluded entity into separate strings.
	var part1, part2 strings.Builder
	for i := 0; i < firstIdx; i++ {
		part1.WriteString(entToString(tokens[i]))
	}
	for i := lastIdx + 1; i < len(tokens); i++ {
		part2.WriteString(entToString(tokens[i]))
	}

	return part1.String(), part2.String()
}

// SerializeToString converts an AST node into a flattened string representation.
func SerializeToString(e Ent) string {
	tokens := SerializeToTokens(e)
	var sb strings.Builder
	for _, tok := range tokens {
		sb.WriteString(entToString(tok))
	}
	return sb.String()
}

// SerializeToTokens performs a DFS traversal of the AST and returns a
// list of atomic tokens (basic Ent types such as LetterEnt, NumEnt, etc.).
func SerializeToTokens(e Ent) []Ent {
	switch t := e.(type) {

	// 1) Basic tokens (letters, numbers, whitespace, symbols, etc.)
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		// Directly converted into a single element.
		return []Ent{e}

	// 2) keyValEnt (order: key, optional left padding, delimiter, optional right padding, value)
	case *KeyValEnt:
		var out []Ent
		// k
		out = append(out, SerializeToTokens(t.k)...)
		// plOrNil
		if t.plOrNil != nil {
			out = append(out, SerializeToTokens(t.plOrNil)...)
		}
		// delimiter
		out = append(out, SerializeToTokens(t.delimiter)...)
		// prOrNil
		if t.prOrNil != nil {
			out = append(out, SerializeToTokens(t.prOrNil)...)
		}
		// v
		out = append(out, SerializeToTokens(t.v)...)
		return out

	// 3) groupEnt (first child: opening bracket, last child: closing bracket, middle: content)
	case *GroupEnt:
		var out []Ent
		if len(t.data) > 0 {
			// opening bracket
			out = append(out, SerializeToTokens(t.data[0])...)
			// data except opening,closing breacket
			for i := 1; i < len(t.data)-1; i++ {
				out = append(out, SerializeToTokens(t.data[i])...)
			}
			// closing bracket
			if len(t.data) > 1 {
				out = append(out, SerializeToTokens(t.data[len(t.data)-1])...)
			}
		}
		return out

	// 4) rootEnt
	case *RootEnt:
		// Iterate through the data within rootEnt.
		var out []Ent
		for _, c := range t.data {
			out = append(out, SerializeToTokens(c)...)
		}
		return out
	case nil:
		return nil
	default:
		// If encountering an unknown or unimplemented node type, return an empty list.
		return nil
	}
}

// entToString converts a single token entity into a string representation.
func entToString(e Ent) string {
	switch t := e.(type) {
	case *LetterEnt:
		return string(t.data)
	case *NumEnt:
		return string(t.data)
	case *WsEnt:
		return string(t.data)
	case *MarkEnt:
		return string(t.data)
	case nil:
		return ""
	default:
		return ""
	}
}

// collectIDs recursively gathers the IDs of an entity and its children,
// storing them in the provided map.
func collectIDs(e Ent, m map[int]bool) {
	m[e.ID()] = true
	switch v := e.(type) {
	case *KeyValEnt:
		if v.k != nil {
			collectIDs(v.k, m)
		}
		if v.plOrNil != nil {
			collectIDs(v.plOrNil, m)
		}
		if v.delimiter != nil {
			collectIDs(v.delimiter, m)
		}
		if v.prOrNil != nil {
			collectIDs(v.prOrNil, m)
		}
		if v.v != nil {
			collectIDs(v.v, m)
		}
	case *GroupEnt:
		for _, child := range v.data {
			collectIDs(child, m)
		}
	case *RootEnt:
		for _, child := range v.data {
			collectIDs(child, m)
		}
	default:
	}
}

// cloneEnt creates a deep copy of an Ent, recursively duplicating all
// child entities while preserving structure.
func cloneEnt(e Ent) Ent {
	switch v := e.(type) {
	case *LetterEnt:
		newData := make([]byte, len(v.data))
		copy(newData, v.data)
		return &LetterEnt{
			data:  newData,
			depth: v.depth,
		}

	case *NumEnt:
		newData := make([]byte, len(v.data))
		copy(newData, v.data)
		return &NumEnt{
			data:  newData,
			depth: v.depth,
		}

	case *WsEnt:
		newData := make([]byte, len(v.data))
		copy(newData, v.data)
		return &WsEnt{
			data:  newData,
			depth: v.depth,
		}

	case *MarkEnt:
		newData := make([]byte, len(v.data))
		copy(newData, v.data)
		return &MarkEnt{
			data:  newData,
			depth: v.depth,
		}

	case *KeyValEnt:

		var k, pl, delim, pr, vv Ent
		if v.k != nil {
			k = cloneEnt(v.k)
		}
		if v.plOrNil != nil {
			pl = cloneEnt(v.plOrNil)
		}
		if v.delimiter != nil {
			delim = cloneEnt(v.delimiter)
		}
		if v.prOrNil != nil {
			pr = cloneEnt(v.prOrNil)
		}
		if v.v != nil {
			vv = cloneEnt(v.v)
		}
		return &KeyValEnt{
			depth:     v.depth,
			k:         k,
			plOrNil:   pl,
			delimiter: delim,
			prOrNil:   pr,
			v:         vv,
		}

	case *GroupEnt:
		newData := make([]Ent, len(v.data))
		for i, child := range v.data {
			newData[i] = cloneEnt(child)
		}
		return &GroupEnt{
			depth: v.depth,
			data:  newData,
		}

	case *RootEnt:
		newData := make([]Ent, len(v.data))
		for i, child := range v.data {
			newData[i] = cloneEnt(child)
		}
		return &RootEnt{
			depth: v.depth,
			data:  newData,
		}
	}
	return nil
}

// PrintTree recursively prints the AST structure with indentation to
// visualize entity hierarchy and depth.
func PrintTree(e Ent, indent int) {
	ind := strings.Repeat("  ", indent)
	idInfo := ""

	if e != nil {
		idInfo = fmt.Sprintf("id:%d", e.ID())
	}
	switch t := e.(type) {
	case *LetterEnt:
		fmt.Printf("%sletterEnt(%q, depth:%d, %s)\n", ind, t.data, t.depth, idInfo)
	case *NumEnt:
		fmt.Printf("%snumEnt(%q, hex:%v, depth:%d, %s)\n", ind, t.data, t.Hex, t.depth, idInfo)
	case *WsEnt:
		fmt.Printf("%swsEnt(%q, depth:%d, %s)\n", ind, t.data, t.depth, idInfo)
	case *MarkEnt:
		fmt.Printf("%smarkEnt(%q, depth:%d, %s)\n", ind, t.data, t.depth, idInfo)
	case *KeyValEnt:
		fmt.Printf("%skeyValEnt(depth:%d, %s)[\n", ind, t.depth, idInfo)
		fmt.Printf("%s  k:\n", ind)
		PrintTree(t.k, indent+2)
		fmt.Printf("%s  pl:\n", ind)
		PrintTree(t.plOrNil, indent+2)
		fmt.Printf("%s  delimiter:\n", ind)
		PrintTree(t.delimiter, indent+2)
		fmt.Printf("%s  pr:\n", ind)
		PrintTree(t.prOrNil, indent+2)
		fmt.Printf("%s  v:\n", ind)
		PrintTree(t.v, indent+2)
		fmt.Printf("%s]\n", ind)
	case *GroupEnt:
		fmt.Printf("%sgroupEnt(depth:%d, %s)[\n", ind, t.depth, idInfo)
		for _, c := range t.data {
			PrintTree(c, indent+1)
		}
		fmt.Printf("%s]\n", ind)
	case *RootEnt:
		if t.IsPattern {
			fmt.Printf("%srootEnt(depth:%d, pattern:%q, %s)[\n", ind, t.depth, t.Pattern, idInfo)
		} else {
			fmt.Printf("%srootEnt(depth:%d, %s)[\n", ind, t.depth, idInfo)
		}
		for _, c := range t.data {
			PrintTree(c, indent+1)
		}
		fmt.Printf("%s]\n", ind)
	case nil:
		fmt.Printf("%snil\n", ind)
	default:
		fmt.Printf("%sUnknown token\n", ind)
	}
}

// SafeUTF8 ensures that the given string is valid UTF-8.
func SafeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	// If the input contains invalid UTF-8 sequences, they are removed by "rune()" function.
	return string([]rune(s))
}

// isHexDigit is a lookup table mapping each byte to a boolean indicating
// whether the byte represents a valid hexadecimal digit.
var isHexDigit [256]bool

// init initializes the isHexDigit lookup table for hexadecimal digit validation.
func init() {
	// Mark '0' to '9' as valid hex digits.
	for i := byte('0'); i <= byte('9'); i++ {
		isHexDigit[i] = true
	}
	// Mark 'a' to 'f' as valid hex digits.
	for i := byte('a'); i <= byte('f'); i++ {
		isHexDigit[i] = true
	}
	// Mark 'A' to 'F' as valid hex digits.
	for i := byte('A'); i <= byte('F'); i++ {
		isHexDigit[i] = true
	}
	// By sorting ParserMark in descending order, the parser can correctly recognize them during the parsing process.
	ParserMark = sortMarksByLengthDesc(ParserMark)
}
