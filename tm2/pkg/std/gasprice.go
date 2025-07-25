package std

import (
	"fmt"
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
		return GasPrice{}, errors.Wrapf(err, "invalid gas price: %s (invalid price)", gasprice)
	}
	gas, err := ParseCoin(parts[1])
	if err != nil {
		return GasPrice{}, errors.Wrapf(err, "invalid gas price: %s (invalid gas denom)", gasprice)
	}
	if gas.Denom != "gas" {
		return GasPrice{}, errors.New("invalid gas price: %s (invalid gas denom)", gasprice)
	}

	if gas.Amount <= 0 {
		return GasPrice{}, errors.New("invalid gas price: %s (invalid gas amount)", gasprice)
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
			return nil, errors.Wrapf(err, "invalid gas prices: %s", gasprices)
		}
	}
	return res, nil
}

// IsGTE compares the GasPrice with another gas price B. If the coin denom matches AND the fee per gas
// is greater than or equal to gas price B, return true; otherwise, return false.
func (gp GasPrice) IsGTE(gpB GasPrice) (bool, error) {
	if gp.Price.Denom != gpB.Price.Denom {
		return false, errors.New("Gas price denominations should be equal; %s, %s", gp.Price.Denom, gpB.Price.Denom)
	}
	if gp.Gas == 0 || gpB.Gas == 0 {
		return false, errors.New("GasPrice.Gas cannot be zero; %+v, %+v", gp, gpB)
	}

	gpg := big.NewInt(gp.Gas)
	gpa := big.NewInt(gp.Price.Amount)

	gpBg := big.NewInt(gpB.Gas)
	gpBa := big.NewInt(gpB.Price.Amount)

	prod1 := big.NewInt(0).Mul(gpa, gpBg) // gp's price amount * gpB's gas
	prod2 := big.NewInt(0).Mul(gpg, gpBa) // gpB's gas * pg's price amount
	// This is equivalent to checking
	// That the Fee / GasWanted ratio is greater than or equal to the minimum GasPrice per gas.
	// This approach helps us avoid dealing with configurations where the value of
	// the minimum gas price is set to 0.00001ugnot/gas.
	return prod1.Cmp(prod2) >= 0, nil
}

func (gp GasPrice) String() string {
	return fmt.Sprintf("%s/%dgas", gp.Price.String(), gp.Gas)
}
