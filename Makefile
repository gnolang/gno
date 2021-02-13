# gnoland is the main executable of gnolang
# bring this back when ready
# gnoland:
#	echo "Building Gnoland for a wide range of systems"
#	mkdir build || true
#	go build -o build/gno-darwin-amd64 cmd/gnoland/main.go
#	GOOS=darwin GOARCH=arm64 go build -o build/gno-darwin-arm64 cmd/gnoland/main.go
#	GOOS=linux GOARCH=amd64 go build -o build/gno-linux-amd64 cmd/gnoland/main.go
#	GOOS=linux GOARCH=arm64 go build -o build/gno-linux-arm64 cmd/gnoland/main.go
#	GOOS=windows GOARCH=amd64 go build -o build/gno-win-amd64 cmd/gnoland/main.go
#	GOOS=windows GOARCH=amd64 go build -o build/gno-win-arm64 cmd/gnoland/main.go

# goscan scans go code
goscan:
	echo "Building goscan for a wide range of systems"
	mkdir build || true
	go build -o build/goscan cmd/goscan/goscan.go
#	GOOS=darwin GOARCH=arm64 go build -o build/goscan-darwin-arm64 cmd/goscan/goscan.go
#	GOOS=linux GOARCH=amd64 go build -o build/goscan-linux-amd64 cmd/goscan/goscan.go
#	GOOS=linux GOARCH=arm64 go build -o build/goscan-linux-arm64 cmd/goscan/goscan.go
#	GOOS=windows GOARCH=amd64 go build -o build/goscan-win-amd64 cmd/goscan/goscan.go
#	GOOS=windows GOARCH=amd64 go build -o build/goscan-win-arm64 cmd/goscan/goscan.go


# Is logos test code or will it be a part of the system?
logos:
	echo "Building logos for a wide range of systems"
	mkdir build || true
	go build -o build/logos logos/cmd/logos.go
#	GOOS=darwin GOARCH=arm64 go build -o build/logos-darwin-arm64 logos/cmd/logos.go
#	GOOS=linux GOARCH=amd64 go build -o build/logos-linux-amd64 logos/cmd/logos.go
#	GOOS=linux GOARCH=arm64 go build -o build/logos-linux-arm64 logos/cmd/logos.go
#	GOOS=windows GOARCH=amd64 go build -o build/logos-win-amd64 logos/cmd/logos.go
#	GOOS=windows GOARCH=amd64 go build -o build/logos-win-arm64 logos/cmd/logos.go


deps:
	go mod download


all:
	deps goscan logos





# TODO stringer -type=Op
# Unsure what the above refers to. Created a basic Makefile.

# TODO can probobaly make this more specific, but that may not be desirable.  Sometimes it's nice to check if it builds
# everywhere.

