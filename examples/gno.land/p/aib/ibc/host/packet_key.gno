package host

import "gno.land/p/aib/encoding"

const (
	PacketCommitmentBasePrefix      = byte(1)
	PacketReceiptBasePrefix         = byte(2)
	PacketAcknowledgementBasePrefix = byte(3)
)

// PacketCommitmentPrefixKey returns the store key prefix under which packet
// commitments for a particular client are stored.
// clientID must be a generated identifier, not provided externally so key
// collisions are not possible.
func PacketCommitmentPrefixKey(clientID string) []byte {
	return append([]byte(clientID), PacketCommitmentBasePrefix)
}

// PacketCommitmentKey returns the store key of under which a packet commitment
// is stored.
// clientID must be a generated identifier, not provided externally so key
// collisions are not possible.
func PacketCommitmentKey(clientID string, sequence uint64) []byte {
	return append(PacketCommitmentPrefixKey(clientID), encoding.Uint64ToBigEndian(sequence)...)
}

// PacketReceiptPrefixKey returns the store key prefix under which packet
// receipts for a particular channel are stored.
// clientID must be a generated identifier, not provided externally so key
// collisions are not possible.
func PacketReceiptPrefixKey(clientID string) []byte {
	return append([]byte(clientID), PacketReceiptBasePrefix)
}

// PacketReceiptKey returns the store key of under which a packet receipt is
// stored.
// clientID must be a generated identifier, not provided externally so key
// collisions are not possible.
func PacketReceiptKey(clientID string, sequence uint64) []byte {
	return append(PacketReceiptPrefixKey(clientID), encoding.Uint64ToBigEndian(sequence)...)
}

// PacketAcknowledgementPrefixKey returns the store key prefix under which
// packet acknowledgements for a particular channel are stored.  clientID must
// be a generated identifier, not provided externally so key collisions are not
// possible.
func PacketAcknowledgementPrefixKey(clientID string) []byte {
	return append([]byte(clientID), PacketAcknowledgementBasePrefix)
}

// PacketAcknowledgementKey returns the store key of under which a packet
// acknowledgement is stored.
// clientID must be a generated identifier, not provided externally so key
// collisions are not possible.
func PacketAcknowledgementKey(clientID string, sequence uint64) []byte {
	return append(PacketAcknowledgementPrefixKey(clientID), encoding.Uint64ToBigEndian(sequence)...)
}
