package logos

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/stretchr/testify/require"
)

// Tests whether widthOf() and nextCharacter() do the same thing.
func TestStringWidthSlow(t *testing.T) {
	for n := 1; n < 4; n++ {
		bz := make([]byte, n)
		for {
			width1 := widthOf(string(bz))
			width2 := widthOfSlow(string(bz))
			if width1 == 0 {
				if isRepeatedWZJ(bz) {
					// these bytes encode one or more U+200D WZJ as UTF8.
				} else {
					require.Fail(t, fmt.Sprintf("unexpected zero width string for bytes %X", bz))
				}
			} else {
				require.True(t, 0 < width1, "got zero width for bytes %X", bz)
			}
			require.Equal(t, width1, width2)
			if !incBuffer(bz) {
				break
			}
		}
	}
}

// Same as above but for longer pseudo-random strings.
func TestStringWidthRandom(t *testing.T) {
	max := 10 * 1024 * 1024
	for i := 0; i < max; i++ {
		if i%(max/80) == 0 {
			fmt.Print(".")
		}
		bz := random.RandBytes(12)
		width1 := widthOf(string(bz))
		width2 := widthOfSlow(string(bz))
		if width1 == 0 {
			if isRepeatedWZJ(bz) {
				// these bytes encode one or more U+200D WZJ as UTF8.
			} else {
				require.Fail(t, "unexpected zero width string")
			}
		} else {
			require.True(t, 0 < width1, "got zero width for bytes %X", bz)
		}
		require.Equal(t, width1, width2,
			"want %d but got %d the slow way: %X",
			width1, width2, bz)
	}
}

// For debugging.
func TestStringWidthDummy(t *testing.T) {
	bz := []byte{0x0C, 0x5B, 0x0D, 0xCF, 0xC5, 0xE2, 0x80, 0x8D, 0xC1, 0x32, 0x69, 0x41}
	width1 := widthOf(string(bz))
	width2 := widthOfSlow(string(bz))
	if width1 == 0 {
		if isRepeatedWZJ(bz) {
			// these bytes encode one or more U+200D WZJ as UTF8.
		} else {
			require.Fail(t, "unexpected zero width string")
		}
	} else {
		require.True(t, 0 < width1, "got zero width for bytes %X", bz)
	}
	require.Equal(t, width1, width2,
		"want %d but got %d the slow way: %X",
		width1, width2, bz)
}

// For debugging.
func TestStringWidthDummy2(t *testing.T) {
	// NOTE: this is broken in the OSX terminal.  This should print a USA flag
	// and have width 2, or possibly default to two block letters "U" and "S",
	// but my terminal prints a flag of width 1.
	bz := []byte("\U0001f1fa\U0001f1f8")
	width1 := widthOf(string(bz))
	width2 := widthOfSlow(string(bz))
	require.Equal(t, 1, width1)
	require.Equal(t, width1, width2,
		"want %d but got %d the slow way: %X",
		width1, width2, bz)
}

func isRepeatedWZJ(bz []byte) bool {
	if len(bz)%3 != 0 {
		return false
	}
	// this is U+200D is UTF8.
	for i := 0; i < len(bz); i += 3 {
		if bz[i] != 0xE2 {
			return false
		}
		if bz[i+1] != 0x80 {
			return false
		}
		if bz[i+2] != 0x8D {
			return false
		}
	}
	return true
}

// get the width of a string using nextCharacter().
func widthOfSlow(s string) (w int) {
	rz := toRunes(s)
	for 0 < len(rz) {
		_, w2, n := nextCharacter(rz)
		if n == 0 {
			panic("should not happen")
		}
		w += w2
		rz = rz[n:]
	}
	return
}

//----------------------------------------
// incBuffer for testing

// If overflow, bz becomes zero and returns false.
func incBuffer(bz []byte) bool {
	for i := 0; i < len(bz); i++ {
		if bz[i] == 0xFF {
			bz[i] = 0x00
		} else {
			bz[i]++
			return true
		}
	}
	return false
}

func TestIncBuffer1(t *testing.T) {
	bz := []byte{0x00}
	for i := 0; i < (1<<(1*8))-1; i++ {
		require.Equal(t, true, incBuffer(bz))
		require.Equal(t, byte(i+1), bz[0])
	}
	require.Equal(t, false, incBuffer(bz))
	require.Equal(t, byte(0x00), bz[0])
}

func TestIncBuffer2(t *testing.T) {
	bz := []byte{0x00, 0x00}
	for i := 0; i < (1<<(2*8))-1; i++ {
		require.Equal(t, true, incBuffer(bz))
		require.Equal(t, byte(((i+1)>>0)%256), bz[0])
		require.Equal(t, byte(((i+1)>>8)%256), bz[1])
	}
	require.Equal(t, []byte{0xFF, 0xFF}, bz)
	require.Equal(t, false, incBuffer(bz))
	require.Equal(t, byte(0x00), bz[0])
	require.Equal(t, byte(0x00), bz[1])
}
