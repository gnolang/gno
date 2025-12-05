package gas

// Gas measured by the SDK.
type Gas = int64

// Gas consumption descriptors.
const (
	IterNextCostFlatDesc = "IterNextFlat"
	ValuePerByteDesc     = "ValuePerByte"
	WritePerByteDesc     = "WritePerByte"
	ReadPerByteDesc      = "ReadPerByte"
	WriteCostFlatDesc    = "WriteFlat"
	ReadCostFlatDesc     = "ReadFlat"
	HasDesc              = "Has"
	DeleteDesc           = "Delete"
)
