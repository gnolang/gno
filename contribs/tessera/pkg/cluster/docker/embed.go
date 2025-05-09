package docker

import _ "embed"

var (
	// dockerFile is the embedded Dockerfile tessera uses to build
	//go:embed Dockerfile
	dockerFile []byte

	// dockerIgnore is the embedded .dockerignore tessera uses for the image
	//go:embed .dockerignore
	dockerIgnore []byte
)
