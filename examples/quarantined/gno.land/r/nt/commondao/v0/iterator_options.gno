package commondao

// IteratorOption configures iterators.
type IteratorOption func(*Iterator)

// WithOffset assigns a initial offset to start iterating DAOs.
// Offset can be used to reduce the number of iterated DAOs for example during pagination.
func WithOffset(offset uint) IteratorOption {
	return func(it *Iterator) {
		it.current = int(offset)
	}
}

// WithCount assigns a number to limit the number of iterated DAOs.
// A zero count means to iterates until the last DAO.
func WithCount(count uint) IteratorOption {
	return func(it *Iterator) {
		it.count = int(count)
	}
}
