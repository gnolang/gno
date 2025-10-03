# `avl/list` - Dynamic List with AVL Tree Backend

A dynamic list data structure backed by an AVL tree, providing O(log n) operations while maintaining order stability. Combines the benefits of arrays (indexed access) with the efficiency of balanced trees.

## Features

- **O(log n) operations**: Efficient insert, delete, get, and set operations
- **Order preservation**: Maintains insertion order and supports index-based access
- **Dynamic resizing**: Grows and shrinks as needed
- **Generic values**: Stores any type using `any` interface
- **Range operations**: Support for range queries and iteration
- **Memory efficient**: Only allocates memory as needed

## Usage

```go
import "gno.land/p/nt/avl/list"

// Create new list
var l list.List

// Add elements
l.Append(1, 2, 3)
l.Append("hello", "world")

// Access by index
value := l.Get(1)    // returns 2
l.Set(1, 42)         // updates index 1 to 42

// Insert at specific position
l.Insert(0, "first") // insert at beginning

// Delete elements
l.Delete(2)          // delete element at index 2

// Size and capacity
size := l.Size()     // current number of elements

// Check if empty
isEmpty := l.IsEmpty()
```

## Iteration

```go
// Iterate over all elements
l.ForEach(func(index int, value any) bool {
    fmt.Printf("index %d: %v\n", index, value)
    return false // continue iteration (return true to stop)
})

// Iterate over range
l.ForEachInRange(1, 3, func(index int, value any) bool {
    fmt.Printf("index %d: %v\n", index, value)
    return false
})

// Reverse iteration
l.ReverseForEach(func(index int, value any) bool {
    fmt.Printf("index %d: %v\n", index, value)
    return false
})
```

## API

```go
type List struct {
    // private fields
}

// Construction
func New() *List
func NewWithSize(size int) *List

// Element access
func (l *List) Get(index int) any
func (l *List) Set(index int, value any) bool
func (l *List) Append(values ...any)
func (l *List) Insert(index int, value any) bool
func (l *List) Delete(index int) bool

// Queries
func (l *List) Size() int
func (l *List) IsEmpty() bool
func (l *List) IndexOf(value any) int
func (l *List) Contains(value any) bool

// Iteration
func (l *List) ForEach(fn func(int, any) bool)
func (l *List) ForEachInRange(start, end int, fn func(int, any) bool)
func (l *List) ReverseForEach(fn func(int, any) bool)

// Conversion
func (l *List) ToSlice() []any
```

## Advanced Examples

### Task Queue Implementation

```go
type TaskQueue struct {
    tasks *list.List
}

func NewTaskQueue() *TaskQueue {
    return &TaskQueue{
        tasks: list.New(),
    }
}

func (tq *TaskQueue) Enqueue(task Task) {
    tq.tasks.Append(task)
}

func (tq *TaskQueue) Dequeue() (Task, bool) {
    if tq.tasks.IsEmpty() {
        return nil, false
    }
    
    task := tq.tasks.Get(0)
    tq.tasks.Delete(0)
    return task.(Task), true
}

func (tq *TaskQueue) Peek() (Task, bool) {
    if tq.tasks.IsEmpty() {
        return nil, false
    }
    return tq.tasks.Get(0).(Task), true
}

func (tq *TaskQueue) Size() int {
    return tq.tasks.Size()
}
```

### Dynamic Buffer

```go
type Buffer struct {
    data *list.List
    maxSize int
}

func NewBuffer(maxSize int) *Buffer {
    return &Buffer{
        data: list.New(),
        maxSize: maxSize,
    }
}

func (b *Buffer) Add(item any) {
    b.data.Append(item)
    
    // Remove oldest items if over capacity
    for b.data.Size() > b.maxSize {
        b.data.Delete(0)
    }
}

func (b *Buffer) GetRecent(count int) []any {
    size := b.data.Size()
    if count > size {
        count = size
    }
    
    result := make([]any, 0, count)
    start := size - count
    
    for i := start; i < size; i++ {
        result = append(result, b.data.Get(i))
    }
    
    return result
}
```

### Undo/Redo System

```go
type UndoRedoSystem struct {
    actions *list.List
    currentIndex int
}

func NewUndoRedoSystem() *UndoRedoSystem {
    return &UndoRedoSystem{
        actions: list.New(),
        currentIndex: -1,
    }
}

func (urs *UndoRedoSystem) Execute(action Action) {
    // Remove any actions after current position
    size := urs.actions.Size()
    for i := urs.currentIndex + 1; i < size; i++ {
        urs.actions.Delete(urs.currentIndex + 1)
    }
    
    // Add new action
    urs.actions.Append(action)
    urs.currentIndex++
    
    // Execute the action
    action.Execute()
}

func (urs *UndoRedoSystem) Undo() bool {
    if urs.currentIndex < 0 {
        return false
    }
    
    action := urs.actions.Get(urs.currentIndex).(Action)
    action.Undo()
    urs.currentIndex--
    
    return true
}

func (urs *UndoRedoSystem) Redo() bool {
    if urs.currentIndex >= urs.actions.Size()-1 {
        return false
    }
    
    urs.currentIndex++
    action := urs.actions.Get(urs.currentIndex).(Action)
    action.Execute()
    
    return true
}
```

## Performance Characteristics

- **Get/Set by index**: O(log n)
- **Append**: O(log n)
- **Insert at position**: O(log n)
- **Delete at position**: O(log n)
- **Size**: O(1)
- **Iteration**: O(n) for n elements

## Use Cases

- **Dynamic arrays**: When you need efficient indexed access with dynamic sizing
- **Queues and stacks**: Implement various queue types with O(log n) operations
- **Undo/redo systems**: Maintain operation history with efficient access
- **Buffer management**: Circular buffers and sliding windows
- **Event logging**: Maintain ordered event lists with efficient insertion

## Comparison with Standard Arrays

**Advantages:**
- Dynamic sizing without memory reallocation
- Efficient insertion/deletion at any position
- Memory usage grows only as needed

**Trade-offs:**
- O(log n) vs O(1) for indexed access
- More memory overhead per element
- More complex implementation

This package is ideal when you need the flexibility of dynamic arrays with efficient operations at any position, especially for applications with frequent insertions and deletions.
