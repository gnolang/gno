package absacc

import (
	"errors"
	"std"

	"gno.land/p/demo/auth"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/authreg"
	"gno.land/r/demo/sessions"
)

var (
	source   std.Realm
	id       seqid.ID
	accounts = make(map[seqid.ID]*account)
)

func init() {
	source = std.CurrentRealm()
	authreg.Register(Authenticate)
}

func CreateCallerAccount() seqid.ID {
	authorities := []Authority{
		{
			Provider: std.DerivePkgAddr("gno.land/r/demo/sessions"),
			EntityID: sessions.EntityID(std.PrevRealm().Addr()),
		},
	}
	return CreateAccount(authorities...)
}

func CreateAccount(authorities ...Authority) seqid.ID {
	if len(authorities) == 0 {
		panic(errors.New("must provide at least one authority, otherwise the account is locked"))
	}

	accountID := id.Next()
	acc := account{
		authorities: make(map[std.Address]map[string]struct{}),
	}
	acc.addAuthorities(authorities...)

	accounts[accountID] = &acc
	return accountID
}

func AddAuthorities(autok auth.Token, authorities ...Authority) {
	account := authenticateAccount(autok)
	account.addAuthorities(authorities...)
}

func RemoveAuthorities(autok auth.Token, authorities ...Authority) {
	account := authenticateAccount(autok)
	account.removeAuthorities(authorities...)
}

func AuthToken(accountID seqid.ID, subToken auth.Token) auth.Token {
	if _, ok := accounts[accountID]; !ok {
		panic(errors.New("unknown account"))
	}
	return &token{
		accountID: accountID,
		subToken:  subToken,
	}
}

func Authenticate(autok auth.Token) string {
	return ufmt.Sprintf("%d", uint64(authenticateID(autok)))
}

func EntityID(accountID seqid.ID) string {
	return ufmt.Sprintf("/%s/%d", source.Addr().String(), uint64(accountID))
}

func authenticateID(autok auth.Token) seqid.ID {
	val, ok := autok.(*token)
	if !ok || val == nil {
		panic(auth.ErrInvalidToken)
	}

	entityID := authreg.Authenticate(val.subToken)
	provider := val.subToken.Source().Addr()
	if _, ok := accounts[val.accountID].authorities[provider][entityID]; !ok {
		panic(auth.ErrInvalidToken)
	}

	return val.accountID
}

func authenticateAccount(autok auth.Token) *account {
	return accounts[authenticateID(autok)]
}

type Authority struct {
	Provider std.Address
	EntityID string
}

type token struct {
	accountID seqid.ID
	subToken  auth.Token
}

func (t *token) Source() std.Realm {
	return source
}

type account struct {
	authorities map[std.Address]map[string]struct{}
}

func (a *account) addAuthorities(authorities ...Authority) {
	for _, authority := range authorities {
		if _, ok := a.authorities[authority.Provider]; !ok {
			a.authorities[authority.Provider] = make(map[string]struct{})
		}
		a.authorities[authority.Provider][authority.EntityID] = struct{}{}
	}
}

func (a *account) removeAuthorities(authorities ...Authority) {
	for _, authority := range authorities {
		if _, ok := a.authorities[authority.Provider]; !ok {
			continue
		}
		if authority.EntityID == "" {
			delete(a.authorities, authority.Provider)
			continue
		}
		if _, ok := a.authorities[authority.Provider][authority.EntityID]; !ok {
			continue
		}
		if len(a.authorities[authority.Provider]) == 1 {
			delete(a.authorities, authority.Provider)
		} else {
			delete(a.authorities[authority.Provider], authority.EntityID)
		}
	}
	if len(a.authorities) == 1 {
		panic(errors.New("must keep at least one authority, otherwise the account will be locked"))
	}
}
