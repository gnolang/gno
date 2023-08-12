package gnoverse

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/db"
)

func NewTestingSandbox() Sandbox {
	memDB := db.NewMemDB()

	s := Sandbox{
		DB: memDB,
	}

	err := s.Init()
	if err != nil {
		panic(fmt.Errorf("init testing sandbox: %w", err))
	}

	return s
}
