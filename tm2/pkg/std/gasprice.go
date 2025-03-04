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

	providedFee := big.NewInt(0).Mul(big.NewInt(gp.Price.Amount), big.NewInt(gp.Gas))
	requiredFee := big.NewInt(0).Mul(big.NewInt(gpB.Price.Amount), big.NewInt(gpB.Gas))

	return providedFee.Cmp(requiredFee) >= 0, nil
}
