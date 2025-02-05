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
	// getDepth retrieves the depth value of the entity.
	getDepth() int

	// setID assigns a unique ID to the entity from a tree perspective.
	setID(id int)
	// getID retrieves the assigned entity ID.
	getID() int

	// Returns a printable string representation of the fields in the XXXEnt structure.
	String() string
	// Score-related methods
	// GetScore retrieves the entity's score, which is used by the mutator to
	// determine the next entity to mutate.
	GetScore() uint64
	// SetScore assigns a score to the entity.
	SetScore(s uint64)
	PlusScore(delta int)

	// Mutation strategy counter methods
	// getStrategyCount retrieves the score of a specific mutation strategy.
	// This helps the mutator decide the next strategy to apply.
	getStrategyCount(code MutationCode) uint64
	// plusStrategyCount increments the score of a given mutation strategy.
	plusStrategyCount(code MutationCode, delta int)
	// InitStrategyCount initializes all strategy keys with a value of 1.
	InitStrategyCount()

	// Strategy map retrieval methods
	// getStrategy2count retrieves the overall scores for each mutation strategy.
	getStrategy2count() map[MutationCode]uint64

	// Caching methods for mutation strategies
	// getStrategyCache retrieves an array of mutation strategies along with
	// their cumulative scores. The mutator uses this to select a strategy based
	// on its score distribution.
	getStrategyCache() ([]MutationCode, []uint64)
	// updateStrategyCache updates the cumulative score array for mutation strategies.
	updateStrategyCache()
}

// LetterEnt represents an entity storing alphabetic data such as "apple" or "hello"
// from strings or byte slices.
type LetterEnt struct {
	Data  []byte // Stores the string data.
	Depth int    // Represents the depth of the entity in a tree structure.

	Id int // Unique node ID assigned to each entity in the tree.

	Score uint64 // Score associated with the entity.

	// Strategy usage log
	Strategy2count map[MutationCode]uint64 // Stores scores for each applied mutation strategy.

	// Cache fields
	cachedStrategyCodes []MutationCode // Array of mutation codes.
	cachedStrategySums  []uint64       // Cumulative sum array mapping to mutation codes.
	needsUpdateChache   bool           // Tracks whether the cumulative sum array needs an update.
}

// Implements Ent interface.
func (l *LetterEnt) setDepth(d int) { l.Depth = d }
func (l *LetterEnt) getDepth() int  { return l.Depth }

func (l *LetterEnt) setID(id int) { l.Id = id }
func (l *LetterEnt) getID() int   { return l.Id }

func (l *LetterEnt) String() string {
	return fmt.Sprintf("LetterEnt{Data:%q, Depth:%d, Id:%d, Score: %d}", l.Data, l.Depth, l.Id, l.Score)
}
func (l *LetterEnt) SetScore(s uint64)   { l.Score = s }
func (l *LetterEnt) GetScore() uint64    { return l.Score }
func (l *LetterEnt) PlusScore(delta int) { l.Score += uint64(delta) }

func (l *LetterEnt) getStrategyCount(code MutationCode) uint64 {
	if l.Strategy2count == nil {
		l.InitStrategyCount()
	}
	if count, ok := l.Strategy2count[code]; ok {
		return count
	}
	return 1
}

func (l *LetterEnt) plusStrategyCount(code MutationCode, delta int) {
	if l.Strategy2count == nil {
		l.InitStrategyCount()
	}
	l.needsUpdateChache = true
	l.Strategy2count[code] += uint64(delta)
}

func (l *LetterEnt) InitStrategyCount() {
	l.Strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		l.Strategy2count[code] = 1
	}
}

func (l *LetterEnt) getStrategy2count() map[MutationCode]uint64 {
	return l.Strategy2count
}

func (l *LetterEnt) getStrategyCache() ([]MutationCode, []uint64) {
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
	if l.Strategy2count == nil {
		l.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(l.Strategy2count)
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
	Data []byte
	Hex  bool // Hex indicates whether the data is represented in hexadecimal notation.

	Depth int

	Id    int
	Score uint64

	Strategy2count map[MutationCode]uint64

	cachedStrategyCodes []MutationCode
	cachedStrategySums  []uint64
	needsUpdateChache   bool
}

// Implements Ent interface.
func (n *NumEnt) setDepth(d int) { n.Depth = d }
func (n *NumEnt) getDepth() int  { return n.Depth }

func (n *NumEnt) setID(id int) { n.Id = id }
func (n *NumEnt) getID() int   { return n.Id }
func (n *NumEnt) String() string {
	return fmt.Sprintf("NumEnt{Data:%q, Hex:%v, Depth:%d, Id:%d, Score: %d}", n.Data, n.Hex, n.Depth, n.Id, n.Score)
}
func (n *NumEnt) SetScore(s uint64)   { n.Score = s }
func (n *NumEnt) GetScore() uint64    { return n.Score }
func (n *NumEnt) PlusScore(delta int) { n.Score += uint64(delta) }

func (n *NumEnt) getStrategyCount(code MutationCode) uint64 {
	if n.Strategy2count == nil {
		n.InitStrategyCount()
	}
	if count, ok := n.Strategy2count[code]; ok {
		return count
	}
	return 1
}

func (n *NumEnt) plusStrategyCount(code MutationCode, delta int) {
	if n.Strategy2count == nil {
		n.InitStrategyCount()
	}
	n.needsUpdateChache = true
	n.Strategy2count[code] += uint64(delta)
}

func (n *NumEnt) InitStrategyCount() {
	n.Strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		n.Strategy2count[code] = 1
	}
}

func (n *NumEnt) getStrategy2count() map[MutationCode]uint64 {
	return n.Strategy2count
}

func (n *NumEnt) getStrategyCache() ([]MutationCode, []uint64) {
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
	keys, cum := buildCumulativeStrategyUint64(n.Strategy2count)
	n.cachedStrategyCodes = keys
	n.cachedStrategySums = cum
	n.needsUpdateChache = false
}

// WsEnt represents an entity that stores whitespace characters, such as " ",
// extracted from strings or byte slices.
type WsEnt struct {
	Data  []byte
	Depth int

	Id    int
	Score uint64

	Strategy2count map[MutationCode]uint64

	cachedStrategyCodes []MutationCode
	cachedStrategySums  []uint64
	needsUpdateChache   bool
}

// Implements Ent interface.
func (w *WsEnt) setDepth(d int) { w.Depth = d }
func (w *WsEnt) getDepth() int  { return w.Depth }

func (w *WsEnt) setID(id int) { w.Id = id }
func (w *WsEnt) getID() int   { return w.Id }

func (w *WsEnt) String() string {
	return fmt.Sprintf("WsEnt{Data:%q, Depth:%d, Id:%d, Score: %d}", w.Data, w.Depth, w.Id, w.Score)
}
func (w *WsEnt) SetScore(s uint64)   { w.Score = s }
func (w *WsEnt) GetScore() uint64    { return w.Score }
func (w *WsEnt) PlusScore(delta int) { w.Score += uint64(delta) }

func (w *WsEnt) getStrategyCount(code MutationCode) uint64 {
	if w.Strategy2count == nil {
		w.InitStrategyCount()
	}
	if count, ok := w.Strategy2count[code]; ok {
		return count
	}
	return 1
}

func (w *WsEnt) plusStrategyCount(code MutationCode, delta int) {
	if w.Strategy2count == nil {
		w.InitStrategyCount()
	}
	w.needsUpdateChache = true
	w.Strategy2count[code] += uint64(delta)
}

func (w *WsEnt) InitStrategyCount() {
	w.Strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		w.Strategy2count[code] = 1
	}
}

func (w *WsEnt) getStrategy2count() map[MutationCode]uint64 {
	return w.Strategy2count
}

func (w *WsEnt) getStrategyCache() ([]MutationCode, []uint64) {
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
	keys, cum := buildCumulativeStrategyUint64(w.Strategy2count)
	w.cachedStrategyCodes = keys
	w.cachedStrategySums = cum
	w.needsUpdateChache = false
}

// MarkEnt represents an entity that stores special values such as "!", "HTTP",
// or "AND", extracted from strings or byte slices.
type MarkEnt struct {
	Data  []byte
	Depth int

	Id    int
	Score uint64

	Strategy2count map[MutationCode]uint64

	// 캐시 필드
	cachedStrategyCodes []MutationCode
	cachedStrategySums  []uint64
	needsUpdateChache   bool
}

// Implements Ent interface.
func (m *MarkEnt) setDepth(d int) { m.Depth = d }
func (m *MarkEnt) getDepth() int  { return m.Depth }

func (m *MarkEnt) setID(id int) { m.Id = id }
func (m *MarkEnt) getID() int   { return m.Id }

func (m *MarkEnt) String() string {
	return fmt.Sprintf("MarkEnt{Data:%q, Depth:%d, Id:%d, Score: %d}", m.Data, m.Depth, m.Id, m.Score)
}
func (m *MarkEnt) SetScore(s uint64)   { m.Score = s }
func (m *MarkEnt) GetScore() uint64    { return m.Score }
func (m *MarkEnt) PlusScore(delta int) { m.Score += uint64(delta) }

func (m *MarkEnt) getStrategyCount(code MutationCode) uint64 {
	if m.Strategy2count == nil {
		m.InitStrategyCount()
	}
	if count, ok := m.Strategy2count[code]; ok {
		return count
	}
	return 1
}

func (m *MarkEnt) plusStrategyCount(code MutationCode, delta int) {
	if m.Strategy2count == nil {
		m.InitStrategyCount()
	}
	m.needsUpdateChache = true
	m.Strategy2count[code] += uint64(delta)
}

func (m *MarkEnt) InitStrategyCount() {
	m.Strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		m.Strategy2count[code] = 1
	}
}

func (m *MarkEnt) getStrategy2count() map[MutationCode]uint64 {
	return m.Strategy2count
}

func (m *MarkEnt) getStrategyCache() ([]MutationCode, []uint64) {
	if m.cachedStrategyCodes == nil || m.cachedStrategySums == nil {
		m.updateStrategyCache()
	}
	return m.cachedStrategyCodes, m.cachedStrategySums
}

func (m *MarkEnt) updateStrategyCache() {
	if !m.needsUpdateChache && m.cachedStrategyCodes != nil {
		return
	}
	if m.Strategy2count == nil {
		m.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(m.Strategy2count)
	m.cachedStrategyCodes = keys
	m.cachedStrategySums = cum
	m.needsUpdateChache = false
}

// Composite Entities Begin

// KeyValEnt represents an entity that stores key-delimiter-value structures
// extracted from strings or byte slices, such as "a := 2" or "var = x".
type KeyValEnt struct {
	Depth     int
	K         Ent // Represents the key in the key-delimiter-value structure.
	PlOrNil   Ent // Represents the whitespace entity between the key and the delimiter.
	PrOrNil   Ent // Represents the whitespace entity between the delimiter and the value.
	Delimiter Ent // Represents the delimiter in the key-delimiter-value structure.
	V         Ent // Represents the value in the key-delimiter-value structure.
	Id        int
	Score     uint64

	Strategy2count map[MutationCode]uint64

	cachedStrategyCodes []MutationCode
	cachedStrategySums  []uint64
	needsUpdateChache   bool
}

// Implements Ent interface.
func (kv *KeyValEnt) setDepth(d int) { kv.Depth = d }
func (kv *KeyValEnt) getDepth() int  { return kv.Depth }

func (kv *KeyValEnt) setID(id int) { kv.Id = id }
func (kv *KeyValEnt) getID() int   { return kv.Id }
func (kv *KeyValEnt) String() string {
	return fmt.Sprintf("KeyValEnt{Depth:%d, Id:%d, K=%v, Delim=%v, V=%v, Score: %d}",
		kv.Depth, kv.Id, kv.K, kv.Delimiter, kv.V, kv.Score)
}
func (kv *KeyValEnt) SetScore(s uint64)   { kv.Score = s }
func (kv *KeyValEnt) GetScore() uint64    { return kv.Score }
func (kv *KeyValEnt) PlusScore(delta int) { kv.Score += uint64(delta) }

func (kv *KeyValEnt) getStrategyCount(code MutationCode) uint64 {
	if kv.Strategy2count == nil {
		kv.InitStrategyCount()
	}
	if count, ok := kv.Strategy2count[code]; ok {
		return count
	}
	return 1
}

func (kv *KeyValEnt) plusStrategyCount(code MutationCode, delta int) {
	if kv.Strategy2count == nil {
		kv.InitStrategyCount()
	}
	kv.needsUpdateChache = true
	kv.Strategy2count[code] += uint64(delta)
}

func (kv *KeyValEnt) InitStrategyCount() {
	kv.Strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		kv.Strategy2count[code] = 1
	}
}

func (kv *KeyValEnt) getStrategy2count() map[MutationCode]uint64 {
	return kv.Strategy2count
}

func (kv *KeyValEnt) getStrategyCache() ([]MutationCode, []uint64) {
	if kv.cachedStrategyCodes == nil || kv.cachedStrategySums == nil {
		kv.updateStrategyCache()
	}
	return kv.cachedStrategyCodes, kv.cachedStrategySums
}

func (kv *KeyValEnt) updateStrategyCache() {
	if !kv.needsUpdateChache && kv.cachedStrategyCodes != nil {
		return
	}
	if kv.Strategy2count == nil {
		kv.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(kv.Strategy2count)
	kv.cachedStrategyCodes = keys
	kv.cachedStrategySums = cum
	kv.needsUpdateChache = false
}

// GroupEnt represents an entity that stores grouped structures extracted from
// strings or byte slices, such as "(a)" or "[1,2,3]".
type GroupEnt struct {
	Data      []Ent // Stores entities grouped together in an array.
	Depth     int
	IsPattern bool // Indicates whether the group follows a specific pattern.

	Id    int
	Score uint64

	Strategy2count map[MutationCode]uint64

	cachedStrategyCodes []MutationCode
	cachedStrategySums  []uint64
	needsUpdateChache   bool
}

// Implements Ent interface.
func (g *GroupEnt) String() string {
	return fmt.Sprintf("GroupEnt{Len:%d, Depth:%d, Id:%d, IsPattern:%v, Score: %d}",
		len(g.Data), g.Depth, g.Id, g.IsPattern, g.Score)
}
func (g *GroupEnt) setDepth(d int) { g.Depth = d }
func (g *GroupEnt) getDepth() int  { return g.Depth }

func (g *GroupEnt) setID(id int)        { g.Id = id }
func (g *GroupEnt) getID() int          { return g.Id }
func (g *GroupEnt) SetScore(s uint64)   { g.Score = s }
func (g *GroupEnt) GetScore() uint64    { return g.Score }
func (g *GroupEnt) PlusScore(delta int) { g.Score += uint64(delta) }

func (g *GroupEnt) getStrategyCount(code MutationCode) uint64 {
	if g.Strategy2count == nil {
		g.InitStrategyCount()
	}
	if count, ok := g.Strategy2count[code]; ok {
		return count
	}
	return 1
}

func (g *GroupEnt) plusStrategyCount(code MutationCode, delta int) {
	if g.Strategy2count == nil {
		g.InitStrategyCount()
	}
	g.needsUpdateChache = true
	g.Strategy2count[code] += uint64(delta)
}

func (g *GroupEnt) InitStrategyCount() {
	g.Strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		g.Strategy2count[code] = 1
	}
}

func (g *GroupEnt) getStrategy2count() map[MutationCode]uint64 {
	return g.Strategy2count
}

func (g *GroupEnt) getStrategyCache() ([]MutationCode, []uint64) {
	if g.cachedStrategyCodes == nil || g.cachedStrategySums == nil {
		g.updateStrategyCache()
	}
	return g.cachedStrategyCodes, g.cachedStrategySums
}

// 캐시 관련 메서드 구현
func (g *GroupEnt) updateStrategyCache() {
	if !g.needsUpdateChache && g.cachedStrategyCodes != nil {
		return
	}
	if g.Strategy2count == nil {
		g.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(g.Strategy2count)
	g.cachedStrategyCodes = keys
	g.cachedStrategySums = cum
	g.needsUpdateChache = false
}

// RootEnt represents the root entity that stores, manages, and analyzes
// an entire string or byte sequence.
type RootEnt struct {
	Data  []Ent // Stores all entities within the root.
	Depth int

	IsPattern bool   // Indicates whether the string follows a specific pattern.
	Pattern   string // Stores the expected pattern of the string.

	Id int // Unique ID of the root entity. (it's value is 0)

	Id2Ent map[int]Ent // Maps each contained entity's ID for easy access.
	Score  uint64

	Strategy2count map[MutationCode]uint64

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
func (r *RootEnt) setDepth(d int) { r.Depth = d }
func (r *RootEnt) getDepth() int  { return r.Depth }

func (r *RootEnt) setID(id int) { r.Id = id }
func (r *RootEnt) getID() int   { return r.Id }

func (r *RootEnt) String() string {
	return fmt.Sprintf("RootEnt{Len:%d, Depth:%d, Pattern:%q, Id:%d, Score: %d}",
		len(r.Data), r.Depth, r.Pattern, r.Id, r.Score)
}
func (r *RootEnt) SetScore(s uint64)   { r.Score = s }
func (r *RootEnt) GetScore() uint64    { return r.Score }
func (r *RootEnt) PlusScore(delta int) { r.Score += uint64(delta) }

func (r *RootEnt) InitStrategyCount() {
	r.Strategy2count = make(map[MutationCode]uint64, len(entMutators))
	for code := range entMutators {
		r.Strategy2count[code] = 1
	}
}

func (r *RootEnt) getStrategyCount(code MutationCode) uint64 {
	if r.Strategy2count == nil {
		r.InitStrategyCount()
	}
	if count, ok := r.Strategy2count[code]; ok {
		return count
	}
	return 1
}

func (r *RootEnt) plusStrategyCount(code MutationCode, delta int) {
	if r.Strategy2count == nil {
		r.InitStrategyCount()
	}
	r.needsUpdateStrategyChache = true
	r.Strategy2count[code] += uint64(delta)
}

func (r *RootEnt) getStrategy2count() map[MutationCode]uint64 {
	return r.Strategy2count
}

func (r *RootEnt) getStrategyCache() ([]MutationCode, []uint64) {
	if r.cachedStrategyCodes == nil || r.cachedStrategySums == nil {
		r.updateStrategyCache()
	}
	return r.cachedStrategyCodes, r.cachedStrategySums
}

func (r *RootEnt) updateStrategyCache() {
	if !r.needsUpdateStrategyChache && r.cachedStrategyCodes != nil {
		return
	}
	if r.Strategy2count == nil {
		r.InitStrategyCount()
	}
	keys, cum := buildCumulativeStrategyUint64(r.Strategy2count)
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
		panic("logic error. reflect를 만족하는 EntID가 존재하지 않음")
	}

	// Calls plusStrategyCount for each strategy in UsedStrategyMap, increasing the count by 5 for each usage.
	for code, count := range mutInfo.UsedStrategyMap {
		ent.plusStrategyCount(code, 5*count)
	}

	// Also increments the Score of the entity by 5.
	ent.PlusScore(5)
}

// GetScoreById retrieves the score of an entity with the specified ID.
func (r *RootEnt) GetScoreById(id int) uint64 {
	if ent, exists := r.Id2Ent[id]; exists {
		return ent.GetScore()
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
		// Converts GetScore() return value from int to uint64.
		score := uint64(r.Id2Ent[id].GetScore())
		total += score
		cum[i] = total
	}
	r.cachedScoreIds = keys
	r.cachedScoreCums = cum
	r.needsUpdateScoreChache = false
}

// getScoreCache returns cached score-related slices.
// Updates the cache if it is uninitialized.
func (r *RootEnt) getScoreCache() ([]int, []uint64) {
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
	r := &RootEnt{Data: toks3, Depth: 0, needsUpdateStrategyChache: true, needsUpdateScoreChache: true}
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

// 전역 설정 (mark 목록, bracket)
var parserMark = []string{
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
		return &MarkEnt{Data: []byte("\n")}
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
			lx.next() // 소비
		}
		return &WsEnt{Data: []byte(sb.String())}
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
		return &LetterEnt{Data: []byte(sb.String())}
	}

	// 5) Identifies multi-character markers.
	sorted := sortMarksByLengthDesc(parserMark)
	for _, mk := range sorted {
		if lx.matchString(mk) {
			lx.pos += len([]byte(mk))
			return &MarkEnt{Data: []byte(mk)}
		}
	}

	// 6) Defaults to treating unrecognized single characters as markers.
	b := lx.next()
	return &MarkEnt{Data: b}
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
					lx.next()
					xCharBytes := lx.next()
					xChar, _ := utf8.DecodeRune(xCharBytes)
					hexDigits := ""
					for !lx.eof() {
						pb = lx.peek()
						if pb == nil {
							break
						}
						r, _ = utf8.DecodeRune(pb)
						if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
							hexDigits += string(r)
							lx.next()
						} else {
							break
						}
					}
					if hexDigits == "" {
						return &NumEnt{
							Data: []byte(sign + "0" + string(xChar)),
							Hex:  false,
						}
					}
					return &NumEnt{
						Data: []byte(sign + "0" + string(xChar) + hexDigits),
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
		return &MarkEnt{Data: []byte(sign)}
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
				return &NumEnt{Data: []byte(sign + digits + expLetter + expSign), Hex: false}
			}
			expo = expLetter + expSign + expDigits
		}
	}
	return &NumEnt{Data: []byte(sign + digits + expo), Hex: false}
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
			// (키 후보 아님) => 그대로 결과에 넣고 다음
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
		if !ok2 || !(bytes.Equal(md.Data, []byte(":")) || bytes.Equal(md.Data, []byte("=")) || bytes.Equal(md.Data, []byte(":="))) {
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
			// 값 후보가 아니면 => 실패
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
			Depth:     0,
			K:         left,
			PlOrNil:   leftPadding,
			Delimiter: delimiter,
			PrOrNil:   rightPadding,
			V:         right,
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
		if mk, ok := tok.(*MarkEnt); ok && len(mk.Data) == 1 {
			openRune := rune(mk.Data[0])
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
	closeRune := bracketPairs[rune(open.Data[0])]
	g := &GroupEnt{
		Depth: 0,
		Data:  []Ent{open},
	}
	i := from + 1
	for i < len(tokens) {
		tok := tokens[i]
		if mk, ok := tok.(*MarkEnt); ok && len(mk.Data) == 1 {
			if rune(mk.Data[0]) == closeRune {
				g.Data = append(g.Data, mk)
				return g, i + 1
			}
			// Handle nested parentheses.
			if _, has := bracketPairs[rune(mk.Data[0])]; has {
				sub, consumed := parseBracketGroup(tokens, i)
				g.Data = append(g.Data, sub)
				i = consumed
				continue
			}
		}
		g.Data = append(g.Data, tok)
		i++
	}
	return g, i
}

// applyPatternToRoot assigns a pattern to the root entity.
func applyPatternToRoot(r *RootEnt) *RootEnt {
	p, ok := detectPattern(r.Data)
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
			low := bytes.ToLower(t.Data)
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
			d := bytes.ToLower(t.Data) // modified: []byte 변환 후 소문자 처리
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
			d := t.Delimiter
			var dStr []byte
			if delim, ok := d.(*MarkEnt); ok {
				dStr = delim.Data
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
			if len(gg.Data) > 0 {
				out = append(out, gg.Data[0]) // open bracket
				if len(gg.Data) > 2 {
					body := gg.Data[1 : len(gg.Data)-1]
					out = append(out, flattenTokens(body)...)
				}
				out = append(out, gg.Data[len(gg.Data)-1]) // close bracket
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
	switch kv := e.(type) {
	case *KeyValEnt:
		// - Applies depth propagation for composite entities like KeyValEnt, GroupEnt, and RootEnt.
		finalizeDepth(kv.K, currentDepth+1)
		if kv.PlOrNil != nil {
			finalizeDepth(kv.PlOrNil, currentDepth+1)
		}
		finalizeDepth(kv.Delimiter, currentDepth+1)
		if kv.PrOrNil != nil {
			finalizeDepth(kv.PrOrNil, currentDepth+1)
		}
		finalizeDepth(kv.V, currentDepth+1)
	case *GroupEnt:
		// - Applies depth propagation for composite entities like KeyValEnt, GroupEnt, and RootEnt.
		for _, c := range kv.Data {
			finalizeDepth(c, currentDepth+1)
		}
	case *RootEnt:
		// - Applies depth propagation for composite entities like KeyValEnt, GroupEnt, and RootEnt.
		for _, c := range kv.Data {
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
		if node.K != nil {
			assignId(node.K, currentID, root)
		}
		if node.PlOrNil != nil {
			assignId(node.PlOrNil, currentID, root)
		}
		if node.Delimiter != nil {
			assignId(node.Delimiter, currentID, root)
		}
		if node.PrOrNil != nil {
			assignId(node.PrOrNil, currentID, root)
		}
		if node.V != nil {
			assignId(node.V, currentID, root)
		}

	case *GroupEnt:
		for _, c := range node.Data {
			assignId(c, currentID, root)
		}

	case *RootEnt:
		for _, c := range node.Data {
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
	switch kv := e.(type) {

	case *KeyValEnt:
		if kv.K != nil {
			initializeScore(kv.K)
		}
		if kv.PlOrNil != nil {
			initializeScore(kv.PlOrNil)
		}
		if kv.Delimiter != nil {
			initializeScore(kv.Delimiter)
		}
		if kv.PrOrNil != nil {
			initializeScore(kv.PrOrNil)
		}
		if kv.V != nil {
			initializeScore(kv.V)
		}

	case *GroupEnt:
		for _, c := range kv.Data {
			initializeScore(c)
		}

	case *RootEnt:
		for _, c := range kv.Data {
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
		if node.K != nil {
			initializeStrategyCount(node.K)
		}
		if node.PlOrNil != nil {
			initializeStrategyCount(node.PlOrNil)
		}
		if node.Delimiter != nil {
			initializeStrategyCount(node.Delimiter)
		}
		if node.PrOrNil != nil {
			initializeStrategyCount(node.PrOrNil)
		}
		if node.V != nil {
			initializeStrategyCount(node.V)
		}
	case *GroupEnt:
		for _, child := range node.Data {
			initializeStrategyCount(child)
		}
	case *RootEnt:
		for _, child := range node.Data {
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
		if excludedIDs[token.getID()] {
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
		out = append(out, SerializeToTokens(t.K)...)
		// plOrNil
		if t.PlOrNil != nil {
			out = append(out, SerializeToTokens(t.PlOrNil)...)
		}
		// delimiter
		out = append(out, SerializeToTokens(t.Delimiter)...)
		// prOrNil
		if t.PrOrNil != nil {
			out = append(out, SerializeToTokens(t.PrOrNil)...)
		}
		// v
		out = append(out, SerializeToTokens(t.V)...)
		return out

	// 3) groupEnt (first child: opening bracket, last child: closing bracket, middle: content)
	case *GroupEnt:
		var out []Ent
		if len(t.Data) > 0 {
			// opening bracket
			out = append(out, SerializeToTokens(t.Data[0])...)
			// data except opening,closing breacket
			for i := 1; i < len(t.Data)-1; i++ {
				out = append(out, SerializeToTokens(t.Data[i])...)
			}
			// closing bracket
			if len(t.Data) > 1 {
				out = append(out, SerializeToTokens(t.Data[len(t.Data)-1])...)
			}
		}
		return out

	// 4) rootEnt
	case *RootEnt:
		// Iterate through the data within rootEnt.
		var out []Ent
		for _, c := range t.Data {
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
		return string(t.Data)
	case *NumEnt:
		return string(t.Data)
	case *WsEnt:
		return string(t.Data)
	case *MarkEnt:
		return string(t.Data)
	case nil:
		return ""
	default:
		return ""
	}
}

// collectIDs recursively gathers the IDs of an entity and its children,
// storing them in the provided map.
func collectIDs(e Ent, m map[int]bool) {
	m[e.getID()] = true
	switch v := e.(type) {
	case *KeyValEnt:
		if v.K != nil {
			collectIDs(v.K, m)
		}
		if v.PlOrNil != nil {
			collectIDs(v.PlOrNil, m)
		}
		if v.Delimiter != nil {
			collectIDs(v.Delimiter, m)
		}
		if v.PrOrNil != nil {
			collectIDs(v.PrOrNil, m)
		}
		if v.V != nil {
			collectIDs(v.V, m)
		}
	case *GroupEnt:
		for _, child := range v.Data {
			collectIDs(child, m)
		}
	case *RootEnt:
		for _, child := range v.Data {
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
		newData := make([]byte, len(v.Data))
		copy(newData, v.Data)
		return &LetterEnt{
			Data:  newData,
			Depth: v.Depth,
		}

	case *NumEnt:
		newData := make([]byte, len(v.Data))
		copy(newData, v.Data)
		return &NumEnt{
			Data:  newData,
			Depth: v.Depth,
		}

	case *WsEnt:
		newData := make([]byte, len(v.Data))
		copy(newData, v.Data)
		return &WsEnt{
			Data:  newData,
			Depth: v.Depth,
		}

	case *MarkEnt:
		newData := make([]byte, len(v.Data))
		copy(newData, v.Data)
		return &MarkEnt{
			Data:  newData,
			Depth: v.Depth,
		}

	case *KeyValEnt:

		var k, pl, delim, pr, vv Ent
		if v.K != nil {
			k = cloneEnt(v.K)
		}
		if v.PlOrNil != nil {
			pl = cloneEnt(v.PlOrNil)
		}
		if v.Delimiter != nil {
			delim = cloneEnt(v.Delimiter)
		}
		if v.PrOrNil != nil {
			pr = cloneEnt(v.PrOrNil)
		}
		if v.V != nil {
			vv = cloneEnt(v.V)
		}
		return &KeyValEnt{
			Depth:     v.Depth,
			K:         k,
			PlOrNil:   pl,
			Delimiter: delim,
			PrOrNil:   pr,
			V:         vv,
		}

	case *GroupEnt:
		newData := make([]Ent, len(v.Data))
		for i, child := range v.Data {
			newData[i] = cloneEnt(child)
		}
		return &GroupEnt{
			Depth: v.Depth,
			Data:  newData,
		}

	case *RootEnt:
		newData := make([]Ent, len(v.Data))
		for i, child := range v.Data {
			newData[i] = cloneEnt(child)
		}
		return &RootEnt{
			Depth: v.Depth,
			Data:  newData,
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
		idInfo = fmt.Sprintf("id:%d", e.getID()) // 추가}
	}
	switch t := e.(type) {
	case *LetterEnt:
		fmt.Printf("%sletterEnt(%q, depth:%d, %s)\n", ind, t.Data, t.Depth, idInfo)
	case *NumEnt:
		fmt.Printf("%snumEnt(%q, hex:%v, depth:%d, %s)\n", ind, t.Data, t.Hex, t.Depth, idInfo)
	case *WsEnt:
		fmt.Printf("%swsEnt(%q, depth:%d, %s)\n", ind, t.Data, t.Depth, idInfo)
	case *MarkEnt:
		fmt.Printf("%smarkEnt(%q, depth:%d, %s)\n", ind, t.Data, t.Depth, idInfo)
	case *KeyValEnt:
		fmt.Printf("%skeyValEnt(depth:%d, %s)[\n", ind, t.Depth, idInfo)
		fmt.Printf("%s  k:\n", ind)
		PrintTree(t.K, indent+2)
		fmt.Printf("%s  pl:\n", ind)
		PrintTree(t.PlOrNil, indent+2)
		fmt.Printf("%s  delimiter:\n", ind)
		PrintTree(t.Delimiter, indent+2)
		fmt.Printf("%s  pr:\n", ind)
		PrintTree(t.PrOrNil, indent+2)
		fmt.Printf("%s  v:\n", ind)
		PrintTree(t.V, indent+2)
		fmt.Printf("%s]\n", ind)
	case *GroupEnt:
		fmt.Printf("%sgroupEnt(depth:%d, %s)[\n", ind, t.Depth, idInfo)
		for _, c := range t.Data {
			PrintTree(c, indent+1)
		}
		fmt.Printf("%s]\n", ind)
	case *RootEnt:
		if t.IsPattern {
			fmt.Printf("%srootEnt(depth:%d, pattern:%q, %s)[\n", ind, t.Depth, t.Pattern, idInfo)
		} else {
			fmt.Printf("%srootEnt(depth:%d, %s)[\n", ind, t.Depth, idInfo)
		}
		for _, c := range t.Data {
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
	// If the input contains invalid UTF-8 sequences, they are removed by "rune()"" function.
	return string([]rune(s))
}
