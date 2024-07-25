package gnoclient

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"

	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Call tests
func TestCallSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	keybase := keys.NewInMemory()

	signer := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	// Make Msg config
	msg := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{"test argument"},
		Send:     "",
	}

	// Execute call
	res, err := client.Call(baseCfg, msg)
	require.NoError(t, err)

	expected := "(\"hi test argument\" string)\n\n"
	got := string(res.DeliverTx.Data)

	assert.Equal(t, expected, got)
}

func TestCallSingle_Sponsor_Integration(t *testing.T) {
	// Set up an in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Initialize in-memory key storage
	keybase := keys.NewInMemory()

	// Create signer accounts for sponsor and sponsoree
	sponsor := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	sponsoree := newInMemorySigner(t, keybase, generateMnemonic(t), "test2")

	sponsorInfo, err := sponsor.Info()
	require.NoError(t, err)

	sponsoreeInfo, err := sponsoree.Info()
	require.NoError(t, err)

	// Set up an RPC client to interact with the in-memory node
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Initialize sponsor and sponsoree clients with their respective signers and RPC client
	sponsorClient := Client{
		Signer:    sponsor,
		RPCClient: rpcClient,
	}

	sponsoreeClient := Client{
		Signer:    sponsoree,
		RPCClient: rpcClient,
	}

	// Fetch sponsoree account information before the transaction
	var sponsoreeAccountNumber uint64 = 0
	var sponsoreeSequence uint64 = 0

	sponsoreeBefore, _, _ := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	if sponsoreeBefore != nil {
		sponsoreeAccountNumber = sponsoreeBefore.AccountNumber
		sponsoreeSequence = sponsoreeBefore.Sequence
	}

	// Configure the transaction to be sponsored
	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted: 100000,
			GasFee:    "10000ugnot",
			Memo:      "Test memo",
		},
		SponsorAddress: sponsorInfo.GetAddress(),
	}

	// Create the message for the transaction
	msg := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{"test argument"},
	}

	// Sponsoree creates a new sponsor transaction
	tx, err := sponsoreeClient.NewSponsorTransaction(cfg, msg)
	require.NoError(t, err)

	// Sponsoree signs the transaction
	sponsorTx, err := sponsoreeClient.SignTransaction(*tx, sponsoreeAccountNumber, sponsoreeSequence)
	require.NoError(t, err)

	// Fetch sponsor account information before the transaction
	sponsorBefore, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)

	// Sponsor executes the transaction which received from sponsoree
	res, err := sponsorClient.ExecuteSponsorTransaction(*sponsorTx, sponsorBefore.AccountNumber, sponsorBefore.Sequence)
	require.NoError(t, err)

	// Check the result of the transaction execution
	expected := "(\"hi test argument\" string)\n\n"
	got := string(res.DeliverTx.Data)

	assert.Nil(t, err)
	assert.Equal(t, expected, got)

	// Query sponsoree's balance after the transaction
	sponsoreeAfter, _, err := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	require.NoError(t, err)
	assert.Equal(t, std.Coins(nil), sponsoreeAfter.GetCoins())

	// Query sponsor's balance after the transaction
	sponsorAfter, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)
	expectedSponsorAfter := sponsorBefore.GetCoins().Sub(std.MustParseCoins(cfg.BaseTxCfg.GasFee))
	assert.Equal(t, expectedSponsorAfter, sponsorAfter.GetCoins())
}

func TestCallMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	keybase := keys.NewInMemory()
	signer := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	// Make Msg configs
	msg1 := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{""},
		Send:     "",
	}

	// Same call, different argument
	msg2 := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{"test argument"},
		Send:     "",
	}

	expected := "(\"it works!\" string)\n\n(\"hi test argument\" string)\n\n"

	// Execute call
	res, err := client.Call(baseCfg, msg1, msg2)
	require.NoError(t, err)

	got := string(res.DeliverTx.Data)
	assert.Equal(t, expected, got)
}

func TestCallMultiple_Sponsor_Integration(t *testing.T) {
	// Set up an in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Initialize in-memory key storage
	keybase := keys.NewInMemory()

	// Create signer accounts for sponsor and sponsoree
	sponsor := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	sponsoree := newInMemorySigner(t, keybase, generateMnemonic(t), "test2")

	sponsorInfo, err := sponsor.Info()
	require.NoError(t, err)

	sponsoreeInfo, err := sponsoree.Info()
	require.NoError(t, err)

	// Set up an RPC client to interact with the in-memory node
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Initialize sponsor and sponsoree clients with their respective signers and RPC client
	sponsorClient := Client{
		Signer:    sponsor,
		RPCClient: rpcClient,
	}

	sponsoreeClient := Client{
		Signer:    sponsoree,
		RPCClient: rpcClient,
	}

	// Fetch sponsoree account information before the transaction
	var sponsoreeAccountNumber uint64 = 0
	var sponsoreeSequence uint64 = 0

	sponsoreeBefore, _, _ := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	if sponsoreeBefore != nil {
		sponsoreeAccountNumber = sponsoreeBefore.AccountNumber
		sponsoreeSequence = sponsoreeBefore.Sequence
	}

	// Configure the transaction to be sponsored
	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted: 100000,
			GasFee:    "10000ugnot",
			Memo:      "Test memo",
		},
		SponsorAddress: sponsorInfo.GetAddress(),
	}

	// Create multiple messages for the transaction
	msg1 := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{"test1"},
	}

	msg2 := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{"test2"},
	}

	// Sponsoree creates a new sponsor transaction
	tx, err := sponsoreeClient.NewSponsorTransaction(cfg, msg1, msg2)
	require.NoError(t, err)

	// Sponsoree signs the transaction
	sponsorTx, err := sponsoreeClient.SignTransaction(*tx, sponsoreeAccountNumber, sponsoreeSequence)
	require.NoError(t, err)

	// Fetch sponsor account information before the transaction
	sponsorBefore, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)

	// Sponsor executes the transaction which received from sponsoree
	res, err := sponsorClient.ExecuteSponsorTransaction(*sponsorTx, sponsorBefore.AccountNumber, sponsorBefore.Sequence)
	require.NoError(t, err)

	// Check the result of the transaction execution
	expected := "(\"hi test1\" string)\n\n(\"hi test2\" string)\n\n"
	got := string(res.DeliverTx.Data)

	assert.Nil(t, err)
	assert.Equal(t, expected, got)

	// Query sponsoree's balance after the transaction
	sponsoreeAfter, _, err := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	require.NoError(t, err)
	assert.Equal(t, std.Coins(nil), sponsoreeAfter.GetCoins())

	// Query sponsor's balance after the transaction
	sponsorAfter, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)
	expectedSponsorAfter := sponsorBefore.GetCoins().Sub(std.MustParseCoins(cfg.BaseTxCfg.GasFee))
	assert.Equal(t, expectedSponsorAfter, sponsorAfter.GetCoins())
}

// Send tests
func TestSendSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	keybase := keys.NewInMemory()
	signer := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	// Make Send config for a new address on the blockchain
	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")
	amount := 10
	msg := MsgSend{
		ToAddress: toAddress,
		Send:      std.Coin{"ugnot", int64(amount)}.String(),
	}

	// Execute send
	res, err := client.Send(baseCfg, msg)
	require.NoError(t, err)
	assert.Equal(t, "", string(res.DeliverTx.Data))

	// Get the new account balance
	account, _, err := client.QueryAccount(toAddress)
	require.NoError(t, err)

	expected := std.Coins{{"ugnot", int64(amount)}}
	got := account.GetCoins()

	assert.Equal(t, expected, got)
}

func TestSendSingle_Sponsor_Integration(t *testing.T) {
	// Set up an in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Initialize in-memory key storage
	keybase := keys.NewInMemory()

	// Create signer accounts for sponsor and sponsoree
	sponsor := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	sender := newInMemorySigner(t, keybase, generateMnemonic(t), "test2")

	sponsorInfo, err := sponsor.Info()
	require.NoError(t, err)

	senderInfo, err := sender.Info()
	require.NoError(t, err)

	// Set up an RPC client to interact with the in-memory node
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Initialize sponsor and sponsoree clients with their respective signers and RPC client
	sponsorClient := Client{
		Signer:    sponsor,
		RPCClient: rpcClient,
	}

	senderClient := Client{
		Signer:    sender,
		RPCClient: rpcClient,
	}

	// Ensure sender has enough money to make msg send
	_, err = sponsorClient.Send(BaseTxCfg{
		GasWanted: 1000000,
		GasFee:    "100000ugnot",
		Memo:      "Test memo",
	}, MsgSend{
		ToAddress: senderInfo.GetAddress(),
		Send:      "100000ugnot",
	})
	require.NoError(t, err)

	// Fetch sender account information before the transaction
	var senderAccountNumber uint64 = 0
	var senderSequence uint64 = 0

	senderBefore, _, _ := senderClient.QueryAccount(senderInfo.GetAddress())
	if senderBefore != nil {
		senderAccountNumber = senderBefore.AccountNumber
		senderSequence = senderBefore.Sequence
	}

	// Configure the transaction to be sponsored
	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted: 1000000,
			GasFee:    "100000ugnot",
			Memo:      "Test memo",
		},
		SponsorAddress: sponsorInfo.GetAddress(),
	}

	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")

	// Create the message for the transaction
	msg := MsgSend{
		ToAddress: toAddress,
		Send:      "10000ugnot",
	}

	// sender creates a new sponsor transaction
	tx, err := senderClient.NewSponsorTransaction(cfg, msg)
	require.NoError(t, err)

	// sender signs the transaction
	sponsorTx, err := senderClient.SignTransaction(*tx, senderAccountNumber, senderSequence)
	require.NoError(t, err)

	// Fetch sponsor account information before the transaction
	sponsorBefore, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)

	// Sponsor executes the transaction which received from sender
	res, err := sponsorClient.ExecuteSponsorTransaction(*sponsorTx, sponsorBefore.AccountNumber, sponsorBefore.Sequence)
	require.NoError(t, err)

	// Check the result of the transaction execution
	expected := ""
	got := string(res.DeliverTx.Data)

	assert.Nil(t, err)
	assert.Equal(t, expected, got)

	// Query sponsor's balance after the transaction
	sponsorAfter, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)
	expectedSponsorAfter := sponsorBefore.GetCoins().Sub(std.MustParseCoins(cfg.BaseTxCfg.GasFee))
	assert.Equal(t, expectedSponsorAfter, sponsorAfter.GetCoins())

	// Query sender's balance after the transaction
	senderAfter, _, err := senderClient.QueryAccount(senderInfo.GetAddress())
	require.NoError(t, err)
	expectedSenderAfter := senderBefore.GetCoins().Sub(std.MustParseCoins(msg.Send))
	assert.Equal(t, expectedSenderAfter, senderAfter.GetCoins())

	// Query to's balance after the transaction
	toAfter, _, err := sponsorClient.QueryAccount(toAddress)
	require.NoError(t, err)
	expectedToAfter := std.NewCoins(std.MustParseCoin(msg.Send))
	assert.Equal(t, expectedToAfter, toAfter.GetCoins())
}

func TestSendMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	keybase := keys.NewInMemory()
	signer := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	// Make Msg configs
	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")
	amount1 := 10
	msg1 := MsgSend{
		ToAddress: toAddress,
		Send:      std.Coin{"ugnot", int64(amount1)}.String(),
	}

	// Same send, different argument
	amount2 := 20
	msg2 := MsgSend{
		ToAddress: toAddress,
		Send:      std.Coin{"ugnot", int64(amount2)}.String(),
	}

	// Execute send
	res, err := client.Send(baseCfg, msg1, msg2)
	assert.NoError(t, err)
	assert.Equal(t, "", string(res.DeliverTx.Data))

	// Get the new account balance
	account, _, err := client.QueryAccount(toAddress)
	assert.NoError(t, err)

	expected := std.Coins{{"ugnot", int64(amount1 + amount2)}}
	got := account.GetCoins()

	assert.Equal(t, expected, got)
}

func TestSendMultiple_Sponsor_Integration(t *testing.T) {
	// Set up an in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Initialize in-memory key storage
	keybase := keys.NewInMemory()

	// Create signer accounts for sponsor and sponsoree
	sponsor := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	sender := newInMemorySigner(t, keybase, generateMnemonic(t), "test2")

	sponsorInfo, err := sponsor.Info()
	require.NoError(t, err)

	senderInfo, err := sender.Info()
	require.NoError(t, err)

	// Set up an RPC client to interact with the in-memory node
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Initialize sponsor and sponsoree clients with their respective signers and RPC client
	sponsorClient := Client{
		Signer:    sponsor,
		RPCClient: rpcClient,
	}

	senderClient := Client{
		Signer:    sender,
		RPCClient: rpcClient,
	}

	// Ensure sender has enough money to make msg send
	_, err = sponsorClient.Send(BaseTxCfg{
		GasWanted: 1000000,
		GasFee:    "100000ugnot",
		Memo:      "Test memo",
	}, MsgSend{
		ToAddress: senderInfo.GetAddress(),
		Send:      "100000ugnot",
	})
	require.NoError(t, err)

	// Fetch sender account information before the transaction
	var senderAccountNumber uint64 = 0
	var senderSequence uint64 = 0

	senderBefore, _, _ := senderClient.QueryAccount(senderInfo.GetAddress())
	if senderBefore != nil {
		senderAccountNumber = senderBefore.AccountNumber
		senderSequence = senderBefore.Sequence
	}

	// Configure the transaction to be sponsored
	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted: 1000000,
			GasFee:    "100000ugnot",
			Memo:      "Test memo",
		},
		SponsorAddress: sponsorInfo.GetAddress(),
	}

	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")

	// Create the messages for the transaction
	var amount1 int64 = 20000
	msg1 := MsgSend{
		ToAddress: toAddress,
		Send:      std.NewCoin("ugnot", amount1).String(),
	}

	var amount2 int64 = 20000
	msg2 := MsgSend{
		ToAddress: toAddress,
		Send:      std.NewCoin("ugnot", amount2).String(),
	}

	// sender creates a new sponsor transaction
	tx, err := senderClient.NewSponsorTransaction(cfg, msg1, msg2)
	require.NoError(t, err)

	// sender signs the transaction
	sponsorTx, err := senderClient.SignTransaction(*tx, senderAccountNumber, senderSequence)
	require.NoError(t, err)

	// Fetch sponsor account information before the transaction
	sponsorBefore, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)

	// Sponsor executes the transaction which received from sender
	res, err := sponsorClient.ExecuteSponsorTransaction(*sponsorTx, sponsorBefore.AccountNumber, sponsorBefore.Sequence)
	require.NoError(t, err)

	// Check the result of the transaction execution
	expected := ""
	got := string(res.DeliverTx.Data)

	assert.Nil(t, err)
	assert.Equal(t, expected, got)

	// Query sender's balance after the transaction
	senderAfter, _, err := senderClient.QueryAccount(senderInfo.GetAddress())
	require.NoError(t, err)
	expectSenderAfter := senderBefore.GetCoins().Sub(std.NewCoins(std.NewCoin("ugnot", amount1+amount2)))
	assert.Equal(t, expectSenderAfter, senderAfter.GetCoins())

	// Query sponsor's balance after the transaction
	sponsorAfter, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)
	expectedSponsorAfter := sponsorBefore.GetCoins().Sub(std.MustParseCoins(cfg.BaseTxCfg.GasFee))
	assert.Equal(t, expectedSponsorAfter, sponsorAfter.GetCoins())

	// Query to's balance after the transaction
	toAfter, _, err := sponsorClient.QueryAccount(toAddress)
	require.NoError(t, err)
	expectToAfter := std.NewCoins(std.NewCoin("ugnot", amount1+amount2))
	assert.Equal(t, expectToAfter, toAfter.GetCoins())
}

// Run tests
func TestRunSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	keybase := keys.NewInMemory()
	signer := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	fileBody := `package main
import (
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/tests"
)
func main() {
	println(ufmt.Sprintf("- before: %d", tests.Counter()))
	for i := 0; i < 10; i++ {
		tests.IncCounter()
	}
	println(ufmt.Sprintf("- after: %d", tests.Counter()))
}`

	// Make Msg configs
	msg := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody,
				},
			},
		},
		Send: "",
	}

	res, err := client.Run(baseCfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "- before: 0\n- after: 10\n")
}

func TestRunSingle_Sponsor_Integration(t *testing.T) {
	// Set up an in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Initialize in-memory key storage
	keybase := keys.NewInMemory()

	// Create signer accounts for sponsor and sponsoree
	sponsor := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	sponsoree := newInMemorySigner(t, keybase, generateMnemonic(t), "test2")

	sponsorInfo, err := sponsor.Info()
	require.NoError(t, err)

	sponsoreeInfo, err := sponsoree.Info()
	require.NoError(t, err)

	// Set up an RPC client to interact with the in-memory node
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Initialize sponsor and sponsoree clients with their respective signers and RPC client
	sponsorClient := Client{
		Signer:    sponsor,
		RPCClient: rpcClient,
	}

	sponsoreeClient := Client{
		Signer:    sponsoree,
		RPCClient: rpcClient,
	}

	// Fetch sponsoree account information before the transaction
	var sponsoreeAccountNumber uint64 = 0
	var sponsoreeSequence uint64 = 0

	sponsoreeBefore, _, _ := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	if sponsoreeBefore != nil {
		sponsoreeAccountNumber = sponsoreeBefore.AccountNumber
		sponsoreeSequence = sponsoreeBefore.Sequence
	}

	// Configure the transaction to be sponsored
	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasFee:    "10000ugnot",
			GasWanted: 8000000,
			Memo:      "",
		},
		SponsorAddress: sponsorInfo.GetAddress(),
	}

	fileBody := `package main
	import (
		"gno.land/p/demo/ufmt"
		"gno.land/r/demo/tests"
	)
	func main() {
		println(ufmt.Sprintf("- before: %d", tests.Counter()))
		for i := 0; i < 10; i++ {
			tests.IncCounter()
		}
		println(ufmt.Sprintf("- after: %d", tests.Counter()))
	}`

	// Create the message for the transaction
	msg := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody,
				},
			},
		},
		Send: "",
	}

	// Sponsoree creates a new sponsor transaction
	tx, err := sponsoreeClient.NewSponsorTransaction(cfg, msg)
	require.NoError(t, err)

	// Sponsoree signs the transaction
	sponsorTx, err := sponsoreeClient.SignTransaction(*tx, sponsoreeAccountNumber, sponsoreeSequence)
	require.NoError(t, err)

	// Fetch sponsor account information before the transaction
	sponsorBefore, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)

	// Sponsor executes the transaction which received from sponsoree
	res, err := sponsorClient.ExecuteSponsorTransaction(*sponsorTx, sponsorBefore.AccountNumber, sponsorBefore.Sequence)
	require.NoError(t, err)

	// Check the result of the transaction execution
	expected := "- before: 0\n- after: 10\n"
	got := string(res.DeliverTx.Data)

	assert.Nil(t, err)
	assert.Equal(t, expected, got)

	// Query sponsoree's balance after the transaction
	sponsoreeAfter, _, err := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	require.NoError(t, err)
	assert.Equal(t, std.Coins(nil), sponsoreeAfter.GetCoins())

	// Query sponsor's balance after the transaction
	sponsorAfter, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)
	expectedSponsorAfter := sponsorBefore.GetCoins().Sub(std.MustParseCoins(cfg.BaseTxCfg.GasFee))
	assert.Equal(t, expectedSponsorAfter, sponsorAfter.GetCoins())
}

func TestRunMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	keybase := keys.NewInMemory()
	signer := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	fileBody1 := `package main
import (
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/tests"
)
func main() {
	println(ufmt.Sprintf("- before: %d", tests.Counter()))
	for i := 0; i < 10; i++ {
		tests.IncCounter()
	}
	println(ufmt.Sprintf("- after: %d", tests.Counter()))
}`

	fileBody2 := `package main
import (
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/deep/very/deep"
)
func main() {
	println(ufmt.Sprintf("%s", deep.Render("gnoclient!")))
}`

	// Make Msg configs
	msg1 := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody1,
				},
			},
		},
		Send: "",
	}
	msg2 := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody2,
				},
			},
		},
		Send: "",
	}

	expected := "- before: 0\n- after: 10\nhi gnoclient!\n"

	res, err := client.Run(baseCfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestRunMultiple_Sponsor_Integration(t *testing.T) {
	// Set up an in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Initialize in-memory key storage
	keybase := keys.NewInMemory()

	// Create signer accounts for sponsor and sponsoree
	sponsor := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	sponsoree := newInMemorySigner(t, keybase, generateMnemonic(t), "test2")

	sponsorInfo, err := sponsor.Info()
	require.NoError(t, err)

	sponsoreeInfo, err := sponsoree.Info()
	require.NoError(t, err)

	// Set up an RPC client to interact with the in-memory node
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Initialize sponsor and sponsoree clients with their respective signers and RPC client
	sponsorClient := Client{
		Signer:    sponsor,
		RPCClient: rpcClient,
	}

	sponsoreeClient := Client{
		Signer:    sponsoree,
		RPCClient: rpcClient,
	}

	// Fetch sponsoree account information before the transaction
	var sponsoreeAccountNumber uint64 = 0
	var sponsoreeSequence uint64 = 0

	sponsoreeBefore, _, _ := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	if sponsoreeBefore != nil {
		sponsoreeAccountNumber = sponsoreeBefore.AccountNumber
		sponsoreeSequence = sponsoreeBefore.Sequence
	}

	// Configure the transaction to be sponsored
	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasFee:    "10000ugnot",
			GasWanted: 8000000,
			Memo:      "",
		},
		SponsorAddress: sponsorInfo.GetAddress(),
	}

	fileBody1 := `package main
	import (
		"gno.land/p/demo/ufmt"
		"gno.land/r/demo/tests"
	)
	func main() {
		println(ufmt.Sprintf("- before: %d", tests.Counter()))
		for i := 0; i < 10; i++ {
			tests.IncCounter()
		}
		println(ufmt.Sprintf("- after: %d", tests.Counter()))
	}`

	fileBody2 := `package main
	import (
		"gno.land/p/demo/ufmt"
		"gno.land/r/demo/deep/very/deep"
	)
	func main() {
		println(ufmt.Sprintf("%s", deep.Render("gnoclient!")))
	}`

	// Make Msg configs
	msg1 := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody1,
				},
			},
		},
		Send: "",
	}
	msg2 := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody2,
				},
			},
		},
		Send: "",
	}

	// Sponsoree creates a new sponsor transaction
	tx, err := sponsoreeClient.NewSponsorTransaction(cfg, msg1, msg2)
	require.NoError(t, err)

	// Sponsoree signs the transaction
	sponsorTx, err := sponsoreeClient.SignTransaction(*tx, sponsoreeAccountNumber, sponsoreeSequence)
	require.NoError(t, err)

	// Fetch sponsor account information before the transaction
	sponsorBefore, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)

	// Sponsor executes the transaction which received from sponsoree
	res, err := sponsorClient.ExecuteSponsorTransaction(*sponsorTx, sponsorBefore.AccountNumber, sponsorBefore.Sequence)
	require.NoError(t, err)

	// Check the result of the transaction execution
	expected := "- before: 0\n- after: 10\nhi gnoclient!\n"
	got := string(res.DeliverTx.Data)

	assert.Nil(t, err)
	assert.Equal(t, expected, got)

	// Query sponsoree's balance after the transaction
	sponsoreeAfter, _, err := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	require.NoError(t, err)
	assert.Equal(t, std.Coins(nil), sponsoreeAfter.GetCoins())

	// Query sponsor's balance after the transaction
	sponsorAfter, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)
	expectedSponsorAfter := sponsorBefore.GetCoins().Sub(std.MustParseCoins(cfg.BaseTxCfg.GasFee))
	assert.Equal(t, expectedSponsorAfter, sponsorAfter.GetCoins())
}

// AddPackage tests
func TestAddPackageSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	keybase := keys.NewInMemory()
	signer := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	body := `package echo

func Echo(str string) string {
	return str
}`

	fileName := "echo.gno"
	deploymentPath := "gno.land/p/demo/integration/test/echo"
	deposit := "100ugnot"

	// Make Msg config
	msg := MsgAddPackage{
		Package: &std.MemPackage{
			Name: "echo",
			Path: deploymentPath,
			Files: []*std.MemFile{
				{
					Name: fileName,
					Body: body,
				},
			},
		},
		Deposit: deposit,
	}

	// Execute AddPackage
	_, err = client.AddPackage(baseCfg, msg)
	assert.NoError(t, err)

	// Check for deployed file on the node
	query, err := client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath),
	})
	require.NoError(t, err)
	assert.Equal(t, string(query.Response.Data), fileName)

	// Query balance to validate deposit
	baseAcc, _, err := client.QueryAccount(gnolang.DerivePkgAddr(deploymentPath))
	require.NoError(t, err)
	assert.Equal(t, baseAcc.GetCoins().String(), deposit)
}

func TestAddPackageSingle_Sponsor_Integration(t *testing.T) {
	// Set up an in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Initialize in-memory key storage
	keybase := keys.NewInMemory()

	// Create signer accounts for sponsor and sponsoree
	sponsor := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	sponsoree := newInMemorySigner(t, keybase, generateMnemonic(t), "test2")

	sponsorInfo, err := sponsor.Info()
	require.NoError(t, err)

	sponsoreeInfo, err := sponsoree.Info()
	require.NoError(t, err)

	// Set up an RPC client to interact with the in-memory node
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Initialize sponsor and sponsoree clients with their respective signers and RPC client
	sponsorClient := Client{
		Signer:    sponsor,
		RPCClient: rpcClient,
	}

	sponsoreeClient := Client{
		Signer:    sponsoree,
		RPCClient: rpcClient,
	}

	// Ensure sponsoree has enough money to make msg addpackage
	_, err = sponsorClient.Send(BaseTxCfg{
		GasWanted: 1000000,
		GasFee:    "100000ugnot",
		Memo:      "Test memo",
	}, MsgSend{
		ToAddress: sponsoreeInfo.GetAddress(),
		Send:      "100000ugnot",
	})
	require.NoError(t, err)

	// Fetch sponsoree account information before the transaction
	var sponsoreeAccountNumber uint64 = 0
	var sponsoreeSequence uint64 = 0

	sponsoreeBefore, _, _ := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	if sponsoreeBefore != nil {
		sponsoreeAccountNumber = sponsoreeBefore.AccountNumber
		sponsoreeSequence = sponsoreeBefore.Sequence
	}

	// Configure the transaction to be sponsored
	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasFee:    "10000ugnot",
			GasWanted: 8000000,
			Memo:      "",
		},
		SponsorAddress: sponsorInfo.GetAddress(),
	}

	body := `package echo

func Echo(str string) string {
	return str
}`

	fileName := "echo.gno"
	deploymentPath := "gno.land/p/demo/integration/test/echo"
	deposit := "100ugnot"

	// Make Msg config
	msg := MsgAddPackage{
		Package: &std.MemPackage{
			Name: "echo",
			Path: deploymentPath,
			Files: []*std.MemFile{
				{
					Name: fileName,
					Body: body,
				},
			},
		},
		Deposit: deposit,
	}

	// Sponsoree creates a new sponsor transaction
	tx, err := sponsoreeClient.NewSponsorTransaction(cfg, msg)
	require.NoError(t, err)

	// Sponsoree signs the transaction
	sponsorTx, err := sponsoreeClient.SignTransaction(*tx, sponsoreeAccountNumber, sponsoreeSequence)
	require.NoError(t, err)

	// Fetch sponsor account information before the transaction
	sponsorBefore, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)

	// Sponsor executes the transaction which received from sponsoree
	res, err := sponsorClient.ExecuteSponsorTransaction(*sponsorTx, sponsorBefore.AccountNumber, sponsorBefore.Sequence)
	require.NoError(t, err)

	// Check the result of the transaction execution
	expected := ""
	got := string(res.DeliverTx.Data)

	assert.Nil(t, err)
	assert.Equal(t, expected, got)

	// Check for deployed file on the node
	query, err := sponsorClient.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath),
	})
	require.NoError(t, err)
	assert.Equal(t, string(query.Response.Data), fileName)

	// Query package's balance to validate the deposit amount
	baseAcc, _, err := sponsorClient.QueryAccount(gnolang.DerivePkgAddr(deploymentPath))
	require.NoError(t, err)
	assert.Equal(t, baseAcc.GetCoins().String(), deposit)

	// Query sponsoree's balance after the transaction
	sponsoreeAfter, _, err := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	require.NoError(t, err)
	expectedSponsoreeAfter := sponsoreeBefore.GetCoins().Sub(std.MustParseCoins(deposit))
	assert.Equal(t, expectedSponsoreeAfter, sponsoreeAfter.GetCoins())

	// Query sponsor's balance after the transaction
	sponsorAfter, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)
	expectedSponsorAfter := sponsorBefore.GetCoins().Sub(std.MustParseCoins(cfg.BaseTxCfg.GasFee))
	assert.Equal(t, expectedSponsorAfter, sponsorAfter.GetCoins())
}

func TestAddPackageMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	keybase := keys.NewInMemory()
	signer := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	deposit := "100ugnot"
	deploymentPath1 := "gno.land/p/demo/integration/test/echo"

	body1 := `package echo

func Echo(str string) string {
	return str
}`

	deploymentPath2 := "gno.land/p/demo/integration/test/hello"
	body2 := `package hello

func Hello(str string) string {
	return "Hello " + str + "!"
}`

	msg1 := MsgAddPackage{
		Package: &std.MemPackage{
			Name: "echo",
			Path: deploymentPath1,
			Files: []*std.MemFile{
				{
					Name: "echo.gno",
					Body: body1,
				},
			},
		},
		Deposit: "",
	}

	msg2 := MsgAddPackage{
		Package: &std.MemPackage{
			Name: "hello",
			Path: deploymentPath2,
			Files: []*std.MemFile{
				{
					Name: "gno.mod",
					Body: "module gno.land/p/demo/integration/test/hello",
				},
				{
					Name: "hello.gno",
					Body: body2,
				},
			},
		},
		Deposit: deposit,
	}

	// Execute AddPackage
	_, err = client.AddPackage(baseCfg, msg1, msg2)
	assert.NoError(t, err)

	// Check Package #1
	query, err := client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath1),
	})
	require.NoError(t, err)
	assert.Equal(t, string(query.Response.Data), "echo.gno")

	// Query balance to validate deposit
	baseAcc, _, err := client.QueryAccount(gnolang.DerivePkgAddr(deploymentPath1))
	require.NoError(t, err)
	assert.Equal(t, baseAcc.GetCoins().String(), "")

	// Check Package #2
	query, err = client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath2),
	})
	require.NoError(t, err)
	assert.Contains(t, string(query.Response.Data), "hello.gno")
	assert.Contains(t, string(query.Response.Data), "gno.mod")

	// Query balance to validate deposit
	baseAcc, _, err = client.QueryAccount(gnolang.DerivePkgAddr(deploymentPath2))
	require.NoError(t, err)
	assert.Equal(t, baseAcc.GetCoins().String(), deposit)
}

func TestAddPackageMultiple_Sponsor_Integration(t *testing.T) {
	// Set up an in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Initialize in-memory key storage
	keybase := keys.NewInMemory()

	// Create signer accounts for sponsor and sponsoree
	sponsor := newInMemorySigner(t, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
	sponsoree := newInMemorySigner(t, keybase, generateMnemonic(t), "test2")

	sponsorInfo, err := sponsor.Info()
	require.NoError(t, err)

	sponsoreeInfo, err := sponsoree.Info()
	require.NoError(t, err)

	// Set up an RPC client to interact with the in-memory node
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Initialize sponsor and sponsoree clients with their respective signers and RPC client
	sponsorClient := Client{
		Signer:    sponsor,
		RPCClient: rpcClient,
	}

	sponsoreeClient := Client{
		Signer:    sponsoree,
		RPCClient: rpcClient,
	}

	// Ensure sponsoree has enough money to make msg addpackage
	_, err = sponsorClient.Send(BaseTxCfg{
		GasWanted: 1000000,
		GasFee:    "100000ugnot",
		Memo:      "Test memo",
	}, MsgSend{
		ToAddress: sponsoreeInfo.GetAddress(),
		Send:      "100000ugnot",
	})
	require.NoError(t, err)

	// Fetch sponsoree account information before the transaction
	var sponsoreeAccountNumber uint64 = 0
	var sponsoreeSequence uint64 = 0

	sponsoreeBefore, _, _ := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	if sponsoreeBefore != nil {
		sponsoreeAccountNumber = sponsoreeBefore.AccountNumber
		sponsoreeSequence = sponsoreeBefore.Sequence
	}

	// Configure the transaction to be sponsored
	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasFee:    "10000ugnot",
			GasWanted: 8000000,
			Memo:      "",
		},
		SponsorAddress: sponsorInfo.GetAddress(),
	}

	deposit := "100ugnot"
	deploymentPath1 := "gno.land/p/demo/integration/test/echo"

	body1 := `package echo

func Echo(str string) string {
	return str
}`

	deploymentPath2 := "gno.land/p/demo/integration/test/hello"
	body2 := `package hello

func Hello(str string) string {
	return "Hello " + str + "!"
}`

	msg1 := MsgAddPackage{
		Package: &std.MemPackage{
			Name: "echo",
			Path: deploymentPath1,
			Files: []*std.MemFile{
				{
					Name: "echo.gno",
					Body: body1,
				},
			},
		},
		Deposit: "",
	}

	msg2 := MsgAddPackage{
		Package: &std.MemPackage{
			Name: "hello",
			Path: deploymentPath2,
			Files: []*std.MemFile{
				{
					Name: "gno.mod",
					Body: "module gno.land/p/demo/integration/test/hello",
				},
				{
					Name: "hello.gno",
					Body: body2,
				},
			},
		},
		Deposit: deposit,
	}

	// Sponsoree creates a new sponsor transaction
	tx, err := sponsoreeClient.NewSponsorTransaction(cfg, msg1, msg2)
	require.NoError(t, err)

	// Sponsoree signs the transaction
	sponsorTx, err := sponsoreeClient.SignTransaction(*tx, sponsoreeAccountNumber, sponsoreeSequence)
	require.NoError(t, err)

	// Fetch sponsor account information before the transaction
	sponsorBefore, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)

	// Sponsor executes the transaction which received from sponsoree
	res, err := sponsorClient.ExecuteSponsorTransaction(*sponsorTx, sponsorBefore.AccountNumber, sponsorBefore.Sequence)
	require.NoError(t, err)

	// Check the result of the transaction execution
	expected := ""
	got := string(res.DeliverTx.Data)

	assert.Nil(t, err)
	assert.Equal(t, expected, got)

	// Check Package #1
	query, err := sponsorClient.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath1),
	})
	require.NoError(t, err)
	assert.Equal(t, string(query.Response.Data), "echo.gno")

	// Query package's balance to validate the deposit amount
	baseAcc, _, err := sponsorClient.QueryAccount(gnolang.DerivePkgAddr(deploymentPath1))
	require.NoError(t, err)
	assert.Equal(t, baseAcc.GetCoins().String(), "")

	// Check Package #2
	query, err = sponsorClient.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath2),
	})
	require.NoError(t, err)
	assert.Contains(t, string(query.Response.Data), "hello.gno")
	assert.Contains(t, string(query.Response.Data), "gno.mod")

	// Query package's balance to validate the deposit amount
	baseAcc, _, err = sponsorClient.QueryAccount(gnolang.DerivePkgAddr(deploymentPath2))
	require.NoError(t, err)
	assert.Equal(t, baseAcc.GetCoins().String(), deposit)

	// Query sponsoree's balance after the transaction
	sponsoreeAfter, _, err := sponsoreeClient.QueryAccount(sponsoreeInfo.GetAddress())
	require.NoError(t, err)
	expectedSponsoreeAfter := sponsoreeBefore.GetCoins().Sub(std.MustParseCoins(deposit))
	assert.Equal(t, expectedSponsoreeAfter, sponsoreeAfter.GetCoins())

	// Query sponsor's balance after the transaction
	sponsorAfter, _, err := sponsorClient.QueryAccount(sponsorInfo.GetAddress())
	require.NoError(t, err)
	expectedSponsorAfter := sponsorBefore.GetCoins().Sub(std.MustParseCoins(cfg.BaseTxCfg.GasFee))
	assert.Equal(t, expectedSponsorAfter, sponsorAfter.GetCoins())
}

// todo add more integration tests:
// MsgCall with Send field populated (single/multiple)
// MsgRun with Send field populated (single/multiple)

func newInMemorySigner(t *testing.T, kb keys.Keybase, mnemonic, accName string) *SignerFromKeybase {
	t.Helper()

	_, err := kb.CreateAccount(accName, mnemonic, "", "", uint32(0), uint32(0))
	require.NoError(t, err)

	return &SignerFromKeybase{
		Keybase:  kb,                // Stores keys in memory or on disk
		Account:  accName,           // Account name or bech32 format
		Password: "",                // Password for encryption
		ChainID:  "tendermint_test", // Chain ID for transaction signing
	}
}

func generateMnemonic(t *testing.T) string {
	t.Helper()

	entropy, err := bip39.NewEntropy(256)
	require.NoError(t, err)

	mnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)

	return mnemonic
}
