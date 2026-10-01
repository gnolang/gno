package benchops

// Enabled is true when any benchmarking build tag is set.
// Used to gate all timing collection. The specific flags
// (OpsEnabled, StorageEnabled, NativeEnabled) gate what
// gets exported.
const Enabled = OpsEnabled || StorageEnabled || NativeEnabled
