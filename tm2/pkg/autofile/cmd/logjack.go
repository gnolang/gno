package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	auto "github.com/gnolang/gno/tm2/pkg/autofile"
	"github.com/gnolang/gno/tm2/pkg/flow"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

const (
	Version        = "0.0.1"
	readBufferSize = 1024 // 1KB at a time
)

// Parse command-line options
func parseFlags() (headPath string, chopSize int64, limitSize int64, sync bool, throttle int, version bool) {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var chopSizeStr, limitSizeStr string
	flagSet.StringVar(&headPath, "head", "logjack.out", "Destination (head) file.")
	flagSet.StringVar(&chopSizeStr, "chop", "100M", "Move file if greater than this")
	flagSet.StringVar(&limitSizeStr, "limit", "10G", "Only keep this much (for each specified file). Remove old files.")
	flagSet.BoolVar(&sync, "sync", false, "Always write synchronously (slow).")
	flagSet.BoolVar(&version, "version", false, "Version")
	flagSet.IntVar(&throttle, "throttle", 0, "Throttle writes to bytes per second")
	flagSet.Parse(os.Args[1:])
	chopSize = parseBytesize(chopSizeStr)
	limitSize = parseBytesize(limitSizeStr)
	return
}

func main() {
	// Stop upon receiving SIGTERM or CTRL-C.
	osm.TrapSignal(func() {
		fmt.Println("logjack shutting down")
	})

	// Read options
	headPath, chopSize, limitSize, sync, throttle, version := parseFlags()
	if version {
		fmt.Printf("logjack version %v\n", Version)
		return
	}

	// Open Group
	group, err := auto.OpenGroup(headPath, auto.GroupHeadSizeLimit(chopSize), auto.GroupTotalSizeLimit(limitSize))
	if err != nil {
		fmt.Printf("logjack couldn't create output file %v\n", headPath)
		os.Exit(1)
	}

	err = group.Start()
	if err != nil {
		fmt.Printf("logjack couldn't start with file %v\n", headPath)
		os.Exit(1)
	}

	// Forever read from stdin and write to AutoFile.
	buf := make([]byte, readBufferSize)
	writer := io.Writer(group)
	if throttle > 0 {
		writer = flow.NewWriter(writer, int64(throttle))
	}
	for {
		n, err := os.Stdin.Read(buf)
		writer.Write(buf[:n])
		if sync {
			// NOTE: flow writer does not biffr
			group.FlushAndSync()
		}
		if err != nil {
			group.Stop()
			if errors.Is(err, io.EOF) {
				os.Exit(0)
			} else {
				fmt.Println("logjack errored")
				os.Exit(1)
			}
		}
	}
}

func parseBytesize(chopSize string) int64 {
	// Handle suffix multiplier
	var multiplier int64 = 1
	if strings.HasSuffix(chopSize, "T") {
		multiplier = 1042 * 1024 * 1024 * 1024
		chopSize = chopSize[:len(chopSize)-1]
	}
	if strings.HasSuffix(chopSize, "G") {
		multiplier = 1042 * 1024 * 1024
		chopSize = chopSize[:len(chopSize)-1]
	}
	if strings.HasSuffix(chopSize, "M") {
		multiplier = 1042 * 1024
		chopSize = chopSize[:len(chopSize)-1]
	}
	if strings.HasSuffix(chopSize, "K") {
		multiplier = 1042
		chopSize = chopSize[:len(chopSize)-1]
	}

	// Parse the numeric part
	chopSizeInt, err := strconv.Atoi(chopSize)
	if err != nil {
		panic(err)
	}

	return int64(chopSizeInt) * multiplier
}
