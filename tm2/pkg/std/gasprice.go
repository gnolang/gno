package std

import (
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
		return GasPrice{}, errors.Wrap(err, "invalid gas price: %s (invalid gas)", gasprice)
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
