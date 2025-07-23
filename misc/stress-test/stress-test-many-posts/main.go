package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/connect"
	api_gen "github.com/gnolang/gnonative/v4/api/gen/go"
	"github.com/gnolang/gnonative/v4/api/gen/go/_goconnect"
	"github.com/gnolang/gnonative/v4/service"
)

func main() {
	// Start the Gno Native Kit gRPC service where the remote is gnoland.
	options := []service.GnoNativeOption{
		service.WithTcpAddr("localhost:0"),
		service.WithUseTcpListener(),
	}
	service, err := service.NewGnoNativeService(options...)
	if err != nil {
		log.Fatalf("failed to run GnoNativeService: %v", err)
		return
	}
	defer service.Close()

	// Create a Gno Native Kit gRPC client.
	client := _goconnect.NewGnoNativeServiceClient(
		http.DefaultClient,
		fmt.Sprintf("http://localhost:%d", service.GetTcpPort()),
	)

	// Get test1Address.
	res, err := client.AddressFromBech32(
		context.Background(),
		connect.NewRequest(&api_gen.AddressFromBech32Request{
			Bech32Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		}),
	)
	if err != nil {
		log.Fatal(err)
		return
	}
	test1Address := res.Msg.Address

	if err := setup(client, test1Address); err != nil {
		log.Fatal(err)
		return
	}

	if err := doAction(client, test1Address); err != nil {
		log.Fatal(err)
		return
	}
}

func setup(client _goconnect.GnoNativeServiceClient, test1Address []byte) error {
	// gnoland already has coins for test_1. Recover the test_1 key in our temporary on-disk keybase.
	_, err := client.CreateAccount(
		context.Background(),
		connect.NewRequest(&api_gen.CreateAccountRequest{
			NameOrBech32: "test_1",
			Mnemonic:     "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast",
			Password:     "password",
		}),
	)
	if err != nil {
		return err
	}
	_, err = client.ActivateAccount(
		context.Background(),
		connect.NewRequest(&api_gen.ActivateAccountRequest{
			NameOrBech32: "test_1",
		}),
	)
	if err != nil {
		return err
	}
	_, err = client.SetPassword(
		context.Background(),
		connect.NewRequest(&api_gen.SetPasswordRequest{
			Address:  test1Address,
			Password: "password",
		}),
	)
	if err != nil {
		return err
	}

	// test_1 is already registered. Create a board and a thread. List these fail if already created.
	res, err := client.Call(
		context.Background(),
		connect.NewRequest(&api_gen.CallRequest{
			GasFee:        "1000000ugnot",
			GasWanted:     50_000_000,
			CallerAddress: test1Address,
			Msgs: []*api_gen.MsgCall{{
				PackagePath: "gno.land/r/demo/boards",
				Fnc:         "CreateBoard",
				Args:        []string{"testboard"},
			}},
		}),
	)
	if err != nil {
		return err
	}
	for res.Receive() {
	}

	res, err = client.Call(
		context.Background(),
		connect.NewRequest(&api_gen.CallRequest{
			GasFee:        "1000000ugnot",
			GasWanted:     50_000_000,
			CallerAddress: test1Address,
			Msgs: []*api_gen.MsgCall{{
				PackagePath: "gno.land/r/demo/boards",
				Fnc:         "CreateThread",
				Args:        []string{"1", "Test thread", "Test post"},
			}},
		}),
	)
	if err != nil {
		return err
	}
	for res.Receive() {
	}

	return nil
}

func doAction(client _goconnect.GnoNativeServiceClient, test1Address []byte) error {
	postsPerCall := 50
	totalPostsWanted := 1_000_000

	// A script to call CreateReply postsPerCall times in one transaction.
	// Reply to board #1 post #1.
	code := `package main

import (
	"gno.land/r/demo/boards"
)

func main() {
	for i := 0; i < ` + strconv.Itoa(postsPerCall) + `; i++ {
		boards.CreateReply(cross, boards.BoardID(1), boards.PostID(1), boards.PostID(1), "reply")
	}
}`

	nCalls := 0
	totalElapsed := 0.0
	minElapsed := math.MaxFloat64
	maxElapsed := 0.0
	totalPosts := 0
	fmt.Printf("nPosts, avg. for %d posts [s], min for %d posts [s], max for %d posts [s]\n", postsPerCall, postsPerCall, postsPerCall)
	for totalPosts < totalPostsWanted {
		start := time.Now()

		res, err := client.Run(
			context.Background(),
			connect.NewRequest(&api_gen.RunRequest{
				GasFee:        "10000000ugnot",
				GasWanted:     1_000_000_000,
				CallerAddress: test1Address,
				Msgs: []*api_gen.MsgRun{{
					Package: code,
				}},
			}),
		)
		if err != nil {
			return err
		}
		for res.Receive() {
		}
		if res.Err() != nil {
			return res.Err()
		}
		elapsed := time.Now().Sub(start)
		elapsedSecs := float64(elapsed.Milliseconds()) / 1000.0
		nCalls++
		totalElapsed += elapsedSecs
		minElapsed = math.Min(minElapsed, elapsedSecs)
		maxElapsed = math.Max(maxElapsed, elapsedSecs)
		totalPosts += postsPerCall

		if totalPosts%1000 == 0 {
			fmt.Printf("%d, %f, %f, %f\n", totalPosts, totalElapsed/float64(nCalls), minElapsed, maxElapsed)
			nCalls = 0
			totalElapsed = 0.0
			minElapsed = 1000000.0
			maxElapsed = 0.0
		}
	}

	return nil
}
