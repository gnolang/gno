package params

func setPrefixedString(keeperPrefix, key string, val string)
func setPrefixedBool(keeperPrefix, key string, val bool)
func setPrefixedInt64(keeperPrefix, key string, val int64)
func setPrefixedUint64(keeperPrefix, key string, val uint64)
func setPrefixedBytes(keeperPrefix, key string, val []byte)

// SetPrefixedXXX(module, k, v) is used for module parameters and can be called only from the gno.land/r/sys/params realm.
// It modifies an existing parameter predefined in the module and affects system behavior.

func SetPrefixedString(keeperPrefix, key string, val string) {
	setPrefixedString(keeperPrefix, key, val)
}
func SetPrefixedBool(keeperPrefix, key string, val bool)   { setPrefixedBool(keeperPrefix, key, val) }
func SetPrefixedInt64(keeperPrefix, key string, val int64) { setPrefixedInt64(keeperPrefix, key, val) }
func SetPrefixedUint64(keeperPrefix, key string, val uint64) {
	setPrefixedUint64(keeperPrefix, key, val)
}
func SetPrefixedBytes(keeperPrefix, key string, val []byte) { setPrefixedBytes(keeperPrefix, key, val) }
