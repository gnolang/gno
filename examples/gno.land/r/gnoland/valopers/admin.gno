package valopers

import (
  "gno.land/p/moul/authz"
)

var auth *authz.Authorizer

func Auth() *authz.Authorizer {
  return auth
}

func UpdateInstructions(newInstructions string) {
  err := auth.Do("update-instructions", func() error {
    instructions = newInstructions
    return nil
  })

  if err != nil {
    panic(err)
  }
}
