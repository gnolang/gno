package bech32

// ConvertBits converts a byte slice where each byte is encoding fromBits bits,
// to a byte slice where each byte is encoding toBits bits.
func ConvertBits(data []byte, fromBits, toBits uint8, pad bool) ([]byte, error) {
	if fromBits < 1 || fromBits > 8 || toBits < 1 || toBits > 8 {
		return nil, ErrInvalidBitGroups{}
	}

	// Determine the maximum size the resulting array can have after base
	// conversion, so that we can size it a single time. This might be off
	// by a byte depending on whether padding is used or not and if the input
	// data is a multiple of both fromBits and toBits, but we ignore that and
	// just size it to the maximum possible.
	maxSize := len(data)*int(fromBits)/int(toBits) + 1

	// The final bytes, each byte encoding toBits bits.
	regrouped := make([]byte, 0, maxSize)

	// Keep track of the next byte we create and how many bits we have
	// added to it out of the toBits goal.
	nextByte := byte(0)
	filledBits := uint8(0)

	for _, b := range data {

		// Discard unused bits.
		b <<= 8 - fromBits

		// How many bits remaining to extract from the input data.
		remFromBits := fromBits
		for remFromBits > 0 {
			// How many bits remaining to be added to the next byte.
			remToBits := toBits - filledBits

			// The number of bytes to next extract is the minimum of
			// remFromBits and remToBits.
			toExtract := remFromBits
			if remToBits < toExtract {
				toExtract = remToBits
			}

			// Add the next bits to nextByte, shifting the already
			// added bits to the left.
			nextByte = (nextByte << toExtract) | (b >> (8 - toExtract))

			// Discard the bits we just extracted and get ready for
			// next iteration.
			b <<= toExtract
			remFromBits -= toExtract
			filledBits += toExtract

			// If the nextByte is completely filled, we add it to
			// our regrouped bytes and start on the next byte.
			if filledBits == toBits {
				regrouped = append(regrouped, nextByte)
				filledBits = 0
				nextByte = 0
			}
		}
	}

	// We pad any unfinished group if specified.
	if pad && filledBits > 0 {
		nextByte <<= toBits - filledBits
		regrouped = append(regrouped, nextByte)
		filledBits = 0
		nextByte = 0
	}

	// Any incomplete group must be <= 4 bits, and all zeroes.
	if filledBits > 0 && (filledBits > 4 || nextByte != 0) {
		return nil, ErrInvalidIncompleteGroup{}
	}

	return regrouped, nil
}

func Encode(prefix string, data []byte) (string, error) {
	converted, err := ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", errors.Wrap(err, "encoding bech32 failed")
	}
	return bech32.Encode(hrp, converted)
}

func Decode(s string) (prefix string, data []byte, err error) {
	hrp, data, err := bech32.DecodeNoLimit(bech)
	if err != nil {
		return "", nil, errors.Wrap(err, "decoding bech32 failed")
	}
	converted, err := ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", nil, errors.Wrap(err, "decoding bech32 failed")
	}
	return hrp, converted, nil
}
