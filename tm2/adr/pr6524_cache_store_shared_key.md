# Share immutable cache-store write keys

Each metered write used three copies of the same key bytes. The gas map tracks the last charge, the value cache stores the pending write, and the dirty-key set feeds iteration.

Convert the key to a Go string once in Set and Delete, and pass that string to the cache update. The conversion still copies the caller's bytes. The three maps can share the immutable string safely. Gas charges and refunds occur in their existing order, and dirty-key tracking and checkpoint snapshots keep their existing entries and semantics.

A separate intern table would require another map lookup and a lifetime policy. Changing cache entry representation would also change checkpoint storage. Sharing the string keeps the change local to how the maps receive their keys.

The in-tree benchmark measures repeated metered writes over 1,024 prefilled 49-byte keys. Each operation saves two allocations and 128 bytes on amd64. The measured median went from 807 to 571 ns for Set and from 759 to 524 ns for Delete across seven interleaved samples per arm.

The cache package passed with the race detector. The VM gas tests and the pinned cross-realm app hash passed.
