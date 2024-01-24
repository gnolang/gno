package main

import (
	"context"
	"fmt"

	ff "github.com/peterbourgon/ff/v4"
)

func (s *service) NewDeployCmd() *ff.Command {
	return &ff.Command{
		Name: "deploy",
		Exec: s.execDeploy,
	}
}

func (s *service) execDeploy(ctx context.Context, args []string) error {
	fmt.Println("exec deploy")
	return nil
}
