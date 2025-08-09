package gh

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl"
)

var (
	accounts       avl.Tree    // uri -> Account
	repos          avl.Tree    // uri -> Repo
	issueOrPRs     avl.Tree    // uri -> IssueOrPR
	lastUpdateTime time.Time   // used by the bot to only upload the diff
	oracleAddr     std.Address = "g1eunnckcl6r8ncwj0lrpxu9g5062xcvwxqlrf29"
	adminAddr      std.Address = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @manfred
)

func OracleLastUpdated() time.Time { return lastUpdateTime }

func OracleUpsertAccount(id, name, kind string) {
	assertIsOracle()
	lastUpdateTime = time.Now()

	// get or create
	account := &Account{}
	res, ok := accounts.Get(id)
	if ok {
		account = res.(*Account)
	} else {
		account.id = id
	}

	// update fields
	account.name = name
	account.kind = kind

	if err := account.Validate(); err != nil {
		panic(err)
	}

	// save
	accounts.Set(id, account)
}

func OracleUpsertRepo(id string, isPrivate, isFork bool) {
	assertIsOracle()
	lastUpdateTime = time.Now()

	// get or create
	repo := &Repo{}
	res, ok := repos.Get(id)
	if ok {
		repo = res.(*Repo)
	} else {
		repo.id = id
	}

	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		panic("invalid id")
	}
	ownerID := parts[0]
	name := parts[1]

	// update fields
	repo.name = name
	repo.isPrivate = isPrivate
	repo.isFork = isFork
	repo.owner = AccountByID(ownerID)

	if err := repo.Validate(); err != nil {
		panic(err)
	}

	// save
	repos.Set(id, repo)
}

func AdminSetOracleAddr(new std.Address) {
	assertIsAdmin()
	oracleAddr = new
}

// XXX: remove once it will be easy to query private variables' state.
func AdminGetOracleAddr() std.Address { return oracleAddr }

func assertIsAdmin() {
	if std.GetOrigCaller() != adminAddr {
		panic("restricted area")
	}
}

func assertIsOracle() {
	if std.GetOrigCaller() != oracleAddr {
		panic("restricted area")
	}
}

// TODO: could be a great fit for a vector-based/state machine approach, mostly for optimizations
// func OracleApplyVectors(vectors ...)
