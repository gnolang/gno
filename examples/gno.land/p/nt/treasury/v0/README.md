> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `treasury` - Coin and GRC20 treasury management

Treasury management for coin and GRC20 token transfers in Gno realms. A `Treasury` holds a set of `Banker`s, each responsible for sending a specific asset type, and records the payment history per banker.

# 1. Concepts

- **Treasury**: container that registers one or more `Banker`s and exposes a unified `Send`/`History`/`Balances` API. Also provides a `Render` router for gnoweb pages.
- **Banker**: handler for a single asset type. Built-ins are `CoinsBanker` (native chain coins) and `GRC20Banker` (any number of GRC20 tokens, resolved through a user-supplied `TokenListerFunc`).
- **Payment**: opaque value produced by a banker-specific helper (`NewCoinsPayment`, `NewGRC20Payment`). Each `Payment` is bound to a `BankerID()`, which is how the treasury routes it.

# 2. Usage

```go
import (
    "chain"
    "chain/banker"
    "chain/runtime"

    "gno.land/p/demo/tokens/grc20"
    "gno.land/p/nt/treasury/v0"
)

var (
    tokens = map[string]*grc20.Token{}
    tr     *treasury.Treasury
)

func init() {
    owner := runtime.CurrentRealm().Address() // this realm holds and sends the funds

    // Coins banker owned by this realm.
    coinsBanker, err := treasury.NewCoinsBankerWithOwner(
        owner,
        banker.NewBanker(banker.BankerTypeRealmSend),
    )
    if err != nil {
        panic(err)
    }

    // GRC20 banker that resolves tokens through a lister.
    grc20Banker, err := treasury.NewGRC20BankerWithOwner(owner, func() map[string]*grc20.Token {
        return tokens
    })
    if err != nil {
        panic(err)
    }

    tr, err = treasury.New(
        []treasury.Banker{coinsBanker, grc20Banker},
        runtime.CurrentRealm().PkgPath(),
    )
    if err != nil {
        panic(err)
    }
}

// SendUgnot transfers ugnot from the realm to `to`.
func SendUgnot(cur realm, to address, amount int64) {
    p := treasury.NewCoinsPayment(chain.Coins{{Denom: "ugnot", Amount: amount}}, to)
    if err := tr.Send(0, cur, p); err != nil {
        panic(err)
    }
}

// Render exposes the treasury under the realm's render path.
func Render(path string) string {
    return tr.Render(path)
}
```

# 3. API

## 3.1 Treasury

```go
// Builds a treasury with the provided bankers (at least one required, IDs must be unique).
// pkgPath is the realm's package path, used as the base for the Render router.
func New(bankers []Banker, pkgPath string) (*Treasury, error)

func (t *Treasury) Send(_ int, rlm realm, p Payment) error
func (t *Treasury) History(bankerID string, pageNumber, pageSize int) ([]Payment, error)
func (t *Treasury) Balances(bankerID string) ([]Balance, error)
func (t *Treasury) Address(bankerID string) (string, error)
func (t *Treasury) HasBanker(bankerID string) bool
func (t *Treasury) ListBankerIDs() []string

// Render entry points (a mux router is initialized by `New`).
func (t *Treasury) Render(path string) string
func (t *Treasury) RenderLanding(path string) string
func (t *Treasury) RenderBanker(bankerID, path string) string
func (t *Treasury) RenderBankerHistory(bankerID, path string) string
```

Render routes:
- `""` — landing page, lists each banker.
- `{banker}` — banker details (address, balances, last N payments).
- `{banker}/history` — paginated payment history.

The `history_size` query parameter on `{banker}` controls the preview size (default `5`, `0` hides the preview).

## 3.2 Banker and Payment interfaces

```go
type Banker interface {
    ID() string                     // unique banker ID used for routing
    Send(int, realm, Payment) error // thread the caller's cur; pass 0 as the first arg
    Balances() []Balance
    Address() string                // address used to receive payments
}

type Payment interface {
    BankerID() string    // routes the payment to a banker
    String() string
}

type Balance struct {
    Denom  string
    Amount int64
}

// Capability guard: any entry point that accepts a Banker from an external
// caller MUST verify it before invoking its methods. Validates dynamic type
// only (embedding-based wrappers are rejected), not captured state.
func IsCanonicalBanker(b Banker) bool
```

## 3.3 CoinsBanker

`Banker` for native chain coins. Owns an address and an inner `chain/banker.Banker` (must be the canonical one returned by `banker.NewBanker` — fake implementations are rejected).

```go
func NewCoinsBankerWithOwner(owner address, banker_ banker.Banker) (*CoinsBanker, error)

func NewCoinsPayment(coins chain.Coins, toAddress address) Payment
```

`CoinsBanker.ID()` returns `"Coins"`.

## 3.4 GRC20Banker

`Banker` for GRC20 tokens. Tokens are resolved at send time through a `TokenListerFunc`, so the set of supported tokens can change without rebuilding the banker.

```go
type TokenListerFunc func() map[string]*grc20.Token

func NewGRC20BankerWithOwner(owner address, lister TokenListerFunc) (*GRC20Banker, error)

func NewGRC20Payment(tokenKey string, amount int64, toAddress address) Payment
```

`GRC20Banker.ID()` returns `"GRC20"`. `tokenKey` must be a key in the map returned by the lister.

## 3.5 Errors

```go
ErrNoBankerProvided       // New called with empty bankers slice
ErrDuplicateBanker        // two bankers share the same ID
ErrBankerNotFound         // Send/History/... called with an unknown banker ID
ErrSendPaymentFailed      // wraps the underlying banker error
ErrCurrentRealmIsNotOwner // banker called from a realm other than its owner
ErrNoOwnerProvided
ErrInvalidPaymentType     // payment routed to the wrong banker type
ErrNonCanonicalBanker     // CoinsBanker built from a non-canonical std banker
ErrNonCanonicalBankerImpl // New given a Banker of a non-canonical type
ErrSpoofedRealm           // Send called with a non-current rlm
ErrNoListerProvided
ErrGRC20TokenNotFound
```

# 4. Security

The `Banker` capability model rests on three rules:

- **Construct your own bankers.** Never accept a pre-built `Banker` (including a `*WithOwner` value) from an external realm. A hostile `Balances`/`Address` can report data tied to an attacker address. `New` calls `IsCanonicalBanker` on each banker and rejects foreign types with `ErrNonCanonicalBankerImpl`.
- **`IsCanonicalBanker` checks dynamic TYPE only, not captured state.** Embedding-based wrappers (`type Evil struct { *CoinsBanker }`) are rejected because type assertions are nominal. Any public entry point that takes a `Banker` from a caller must call it before invoking the banker's methods.
- **Owner must match the acting realm.** `Send` asserts `rlm.IsCurrent()` (else `ErrSpoofedRealm`) and the banker rejects a caller that is not its owner (`ErrCurrentRealmIsNotOwner`). Set the owner to the realm that will actually send.
