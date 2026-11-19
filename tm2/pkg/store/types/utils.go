package types

// Iterator over all the keys with a certain prefix in ascending order.
// gctx may be nil for gas-free iteration (node startup, test helpers).
func PrefixIterator(gctx *GasContext, kvs Store, prefix []byte) Iterator {
	return kvs.Iterator(gctx, prefix, PrefixEndBytes(prefix))
}

// Iterator over all the keys with a certain prefix in descending order.
// gctx may be nil for gas-free iteration (node startup, test helpers).
func ReversePrefixIterator(gctx *GasContext, kvs Store, prefix []byte) Iterator {
	return kvs.ReverseIterator(gctx, prefix, PrefixEndBytes(prefix))
}

// PrefixEndBytes returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func PrefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for {
		if end[len(end)-1] != byte(255) {
			end[len(end)-1]++
			break
		} else {
			end = end[:len(end)-1]
			if len(end) == 0 {
				end = nil
				break
			}
		}
	}
	return end
}

// ----------------------------------------
func Cp(bz []byte) (ret []byte) {
	if bz == nil {
		return nil
	}
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}
