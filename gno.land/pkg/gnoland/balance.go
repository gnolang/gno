package gnoland

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Balance represents a genesis account balance with an optional vesting schedule.
type Balance struct {
	Address bft.Address          `json:"address" yaml:"address"`
	Amount  std.Coins            `json:"amount" yaml:"amount"`
	Vesting *std.VestingSchedule `json:"vesting,omitempty" yaml:"vesting,omitempty"`
}

// IsVesting returns true if this balance entry creates a vesting account.
func (b Balance) IsVesting() bool {
	return b.Vesting != nil && !b.Vesting.IsZero()
}

func (b *Balance) Verify() error {
	if b.Address.IsZero() {
		return ErrBalanceEmptyAddress
	}

	if b.Amount.Len() == 0 {
		return ErrBalanceEmptyAmount
	}

	if b.Vesting != nil {
		if err := b.Vesting.Validate(); err != nil {
			return fmt.Errorf("invalid vesting schedule: %w", err)
		}
		if !b.Amount.IsAllGTE(b.Vesting.OriginalVesting) {
			return fmt.Errorf(
				"original vesting amount (%s) exceeds total balance (%s)",
				b.Vesting.OriginalVesting, b.Amount,
			)
		}
	}

	return nil
}

func (b *Balance) Parse(entry string) error {
	// Format: <address>=<coins> [;vesting=<coins>,<start_time>,<end_time>]
	// The vesting suffix is optional.
	parts := strings.SplitN(strings.TrimSpace(entry), ";", 2)
	balancePart := parts[0]

	kv := strings.SplitN(balancePart, "=", 2)
	if len(kv) != 2 {
		return fmt.Errorf("malformed entry: %q", entry)
	}

	var err error

	b.Address, err = crypto.AddressFromBech32(kv[0])
	if err != nil {
		return fmt.Errorf("invalid address %q: %w", kv[0], err)
	}

	b.Amount, err = std.ParseCoins(kv[1])
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", kv[1], err)
	}

	// Parse optional vesting suffix.
	if len(parts) == 2 {
		vestingPart := parts[1]
		if !strings.HasPrefix(vestingPart, "vesting=") {
			return fmt.Errorf("malformed vesting option: %q", vestingPart)
		}
		vestingValue := strings.TrimPrefix(vestingPart, "vesting=")

		// vesting=<coins>,<start_time>,<end_time>
		fields := strings.SplitN(vestingValue, ",", 3)
		if len(fields) != 3 {
			return fmt.Errorf("malformed vesting schedule: expected <coins>,<start>,<end>, got %q", vestingValue)
		}

		var schedule std.VestingSchedule
		schedule.OriginalVesting, err = std.ParseCoins(fields[0])
		if err != nil {
			return fmt.Errorf("invalid vesting amount %q: %w", fields[0], err)
		}
		schedule.StartTime, err = strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid vesting start time %q: %w", fields[1], err)
		}
		schedule.EndTime, err = strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid vesting end time %q: %w", fields[2], err)
		}
		b.Vesting = &schedule
	}

	return nil
}

func (b *Balance) UnmarshalAmino(rep string) error {
	return b.Parse(rep)
}

func (b Balance) MarshalAmino() (string, error) {
	return b.String(), nil
}

func (b Balance) String() string {
	s := fmt.Sprintf("%s=%s", b.Address.String(), b.Amount.String())
	if b.Vesting != nil && !b.Vesting.IsZero() {
		s += fmt.Sprintf(";vesting=%s,%d,%d",
			b.Vesting.OriginalVesting.String(),
			b.Vesting.StartTime,
			b.Vesting.EndTime,
		)
	}
	return s
}

type Balances map[crypto.Address]Balance

func NewBalances() Balances {
	return make(Balances)
}

func (bs Balances) Set(address crypto.Address, amount std.Coins) {
	bs[address] = Balance{
		Address: address,
		Amount:  amount,
	}
}

func (bs Balances) Get(address crypto.Address) (balance Balance, ok bool) {
	balance, ok = bs[address]
	return
}

// List returns a slice of balances, sorted by Balance.Address
// in lexicographic order.
func (bs Balances) List() []Balance {
	list := make([]Balance, 0, len(bs))
	for _, balance := range bs {
		list = append(list, balance)
	}

	SortBalances(list)
	return list
}

// SortBalances sorts balances in lexicographic order, compared by .Address instead of .Amount
// because .Amount's type is Coins that requires a deeper comparison by .Denom and
// .Amount which are unnecessarily complex yet by the nature of each Balance in Balances,
// each entry will be keyed by the same Address in a map.
func SortBalances(list []Balance) {
	slices.SortFunc(list, func(a, b Balance) int {
		return a.Address.Compare(b.Address)
	})
}

// LeftMerge left-merges the two maps
func (bs Balances) LeftMerge(from Balances) {
	for key, bVal := range from {
		if _, present := (bs)[key]; !present {
			(bs)[key] = bVal
		}
	}
}

func GetBalancesFromEntries(entries ...string) (Balances, error) {
	balances := NewBalances()
	return balances, balances.LoadFromEntries(entries...)
}

// LoadFromEntries extracts the balance entries in the form of <address>=<amount>
func (bs Balances) LoadFromEntries(entries ...string) error {
	for _, entry := range entries {
		var balance Balance
		if err := balance.Parse(entry); err != nil {
			return fmt.Errorf("unable to parse balance entry: %w", err)
		}
		bs[balance.Address] = balance
	}

	return nil
}

func GetBalancesFromSheet(sheet io.Reader) (Balances, error) {
	balances := NewBalances()
	return balances, balances.LoadFromSheet(sheet)
}

// LoadFromSheet extracts the balance sheet from the passed in
// balance sheet file, that has the format of <address>=<amount>ugnot
func (bs Balances) LoadFromSheet(sheet io.Reader) error {
	// Parse the balances
	scanner := bufio.NewScanner(sheet)

	for scanner.Scan() {
		entry := scanner.Text()

		// Remove comments
		entry = strings.Split(entry, "#")[0]
		entry = strings.TrimSpace(entry)

		// Skip empty lines
		if entry == "" {
			continue
		}

		if err := bs.LoadFromEntries(entry); err != nil {
			return fmt.Errorf("unable to load entries: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error encountered while scanning, %w", err)
	}

	return nil
}
