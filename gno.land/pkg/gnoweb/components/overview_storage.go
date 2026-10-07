package components

import "strconv"

// PackageStorage is what a realm keeps on chain between transactions, and the
// deposit locked to pay for it, as vm/qstorage reports them.
type PackageStorage struct {
	// Bytes is the realm's stored size.
	Bytes int64
	// Deposit is the ugnot locked for it, refunded as the data is freed.
	Deposit int64
}

// Size is Bytes in decimal units, the ones the chain prices storage in.
func (s PackageStorage) Size() string {
	b := float64(s.Bytes)
	switch {
	case s.Bytes < 1_000:
		return strconv.FormatInt(s.Bytes, 10) + " B"
	case s.Bytes < 1_000_000:
		return strconv.FormatFloat(b/1e3, 'f', 1, 64) + " KB"
	case s.Bytes < 1_000_000_000:
		return strconv.FormatFloat(b/1e6, 'f', 2, 64) + " MB"
	default:
		return strconv.FormatFloat(b/1e9, 'f', 2, 64) + " GB"
	}
}

// DepositGNOT is Deposit in GNOT with cents, so a small realm's deposit does
// not round to nothing.
func (s PackageStorage) DepositGNOT() string {
	if s.Deposit == 0 {
		return "0 GNOT"
	}
	return strconv.FormatFloat(float64(s.Deposit)/1e6, 'f', 2, 64) + " GNOT"
}
