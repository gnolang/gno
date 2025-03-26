package main

import (
	"context"
	"dagger/gno-fee/internal/dagger"
	"strings"
)

type GnoFee struct{}

func (m *GnoFee) BuildAndRun(ctx context.Context, source *dagger.Directory) (int, error) {
	execOpts := dagger.ContainerWithExecOpts{
		UseEntrypoint: true,
	}

	gnoService := source.
		DockerBuild(dagger.DirectoryDockerBuildOpts{Target: "gnoland"}).
		WithExposedPort(26657).
		WithExec([]string{"config", "init"}, execOpts).
		WithExec([]string{"config", "set", "rpc.laddr", "tcp://0.0.0.0:26657"}, execOpts).
		AsService(dagger.ContainerAsServiceOpts{
			Args:          []string{"start", "--lazy", "--log-level", "info"},
			UseEntrypoint: execOpts.UseEntrypoint,
		})

	return dag.Container().
		From("alpine").
		WithServiceBinding("gno", gnoService).
		WithExec(strings.Split("apk add jq curl", " ")).
		WithExec([]string{"sh", "-c",
			"[ $(curl -s gno:26657/status | jq -r '.result.sync_info.latest_block_height') -ge 1 ]"}).
		ExitCode(ctx)
}
