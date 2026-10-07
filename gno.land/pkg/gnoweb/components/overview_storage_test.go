package components

import "testing"

// Sizes are decimal, as the chain prices them (100 ugnot per byte), and the
// deposit is in GNOT with cents, so small realms do not read as zero.
func TestStorageFormatting(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		bytes, deposit int64
		size, gnot     string
	}{
		{1292654, 129265400, "1.29 MB", "129.27 GNOT"},
		{4594, 459400, "4.6 KB", "0.46 GNOT"},
		{512, 51200, "512 B", "0.05 GNOT"},
		{0, 0, "0 B", "0 GNOT"},
	} {
		s := PackageStorage{Bytes: c.bytes, Deposit: c.deposit}
		if got := s.Size(); got != c.size {
			t.Errorf("Size(%d) = %q, want %q", c.bytes, got, c.size)
		}
		if got := s.DepositGNOT(); got != c.gnot {
			t.Errorf("DepositGNOT(%d) = %q, want %q", c.deposit, got, c.gnot)
		}
	}
}
