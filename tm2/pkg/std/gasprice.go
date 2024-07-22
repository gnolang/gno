package std

import (
	"math/big"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

// minimum gas price is Price/Gas per gas unit.
type GasPrice struct {
	Gas   int64 `json:"gas"`
	Price Coin  `json:"price"`
}

func ParseGasPrice(gasprice string) (GasPrice, error) {
	parts := strings.Split(gasprice, "/")
	if len(parts) != 2 {
		return GasPrice{}, errors.New("invalid gas price: %s", gasprice)
	}
	price, err := ParseCoin(parts[0])
	if err != nil {
		return GasPrice{}, errors.Wrap(err, "invalid gas price: %s (invalid price)", gasprice)
	}
	gas, err := ParseCoin(parts[1])
	if err != nil {
		return GasPrice{}, errors.Wrap(err, "invalid gas price: %s (invalid gas denom)", gasprice)
	}
	if gas.Denom != "gas" {
		return GasPrice{}, errors.New("invalid gas price: %s (invalid gas denom)", gasprice)
	}
	return GasPrice{
		Gas:   gas.Amount,
		Price: price,
	}, nil
}

func ParseGasPrices(gasprices string) (res []GasPrice, err error) {
	parts := strings.Split(gasprices, ";")
	if len(parts) == 0 {
		return nil, errors.New("invalid gas prices: %s", gasprices)
	}
	res = make([]GasPrice, len(parts))
	for i, part := range parts {
		res[i], err = ParseGasPrice(part)
		if err != nil {
			return nil, errors.Wrap(err, "invalid gas prices: %s", gasprices)
		}
	}
	return res, nil
}

// IsGTE compares the GasPrice with another gas price B. If coin denom matches AND fee per gas is
// greater or equal to the gas price B return true, other wise return false,
func (gp GasPrice) IsGTE(gpB GasPrice) bool {
	gpg := big.NewInt(gp.Gas)
	gpa := big.NewInt(gp.Price.Amount)
	gpd := gp.Price.Denom

	gpBg := big.NewInt(gpB.Gas)
	gpBa := big.NewInt(gpB.Price.Amount)
	gpBd := gpB.Price.Denom

	if gpd == gpBd {
		prod1 := big.NewInt(0).Mul(gpa, gpBg) // gp's price amount * gpB's gas
		prod2 := big.NewInt(0).Mul(gpg, gpBa) // gpB's gas * pg's price amount
		// This is equivalent to checking
		// That the Fee / GasWanted ratio is greater than or equal to the minimum GasPrice per gas.
		// This approach helps us avoid dealing with configurations where the value of
		// the minimum gas price is set to 0.00001ugnot/gas.
		if prod1.Cmp(prod2) >= 0 {
			return true
		} else {
			return false
		}
	}
	return false

}
