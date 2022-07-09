package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/bft/config"
	"github.com/gnolang/gno/pkgs/bft/node"
	"github.com/gnolang/gno/pkgs/bft/privval"
	bft "github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/log"
	osm "github.com/gnolang/gno/pkgs/os"
	vmm "github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
)

func main() {
	args := os.Args[1:]
	err := runMain(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

var flags struct {
	skipFailingGenesisTxs bool
	skipStart             bool
	airdropFile           string
	chainID               string
	genesisRemote         string
}

func runMain(args []string) error {
	fs := flag.NewFlagSet("gnoland", flag.ExitOnError)
	fs.BoolVar(&flags.skipFailingGenesisTxs, "skip-failing-genesis-txs", false, "don't panic when replaying invalid genesis txs")
	fs.BoolVar(&flags.skipStart, "skip-start", false, "quit after initialization, don't start the node")
	fs.StringVar(&flags.airdropFile, "airdrop-file", "", "optional airdrop file")
	fs.StringVar(&flags.chainID, "chainid", "dev", "chainid")
	fs.StringVar(&flags.genesisRemote, "genesis-remote", "localhost:26657", "replacement for '%%REMOTE%%' in genesis")
	fs.Parse(args)

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	rootDir := "testdir"
	cfg := config.LoadOrMakeConfigWithOptions(rootDir, func(cfg *config.Config) {
		cfg.Consensus.CreateEmptyBlocks = false
		cfg.Consensus.CreateEmptyBlocksInterval = 60 * time.Second
	})

	// create priv validator first.
	// need it to generate genesis.json
	newPrivValKey := cfg.PrivValidatorKeyFile()
	newPrivValState := cfg.PrivValidatorStateFile()
	priv := privval.LoadOrGenFilePV(newPrivValKey, newPrivValState)

	// write genesis file if missing.
	genesisFilePath := filepath.Join(rootDir, cfg.Genesis)
	if !osm.FileExists(genesisFilePath) {
		genDoc := makeGenesisDoc(priv.GetPubKey())
		writeGenesisFile(genDoc, genesisFilePath)
	}

	// create application and node.
	gnoApp, err := gnoland.NewApp(rootDir, flags.skipFailingGenesisTxs, logger)
	if err != nil {
		return fmt.Errorf("error in creating new app: %w", err)
	}
	cfg.LocalApp = gnoApp
	gnoNode, err := node.DefaultNewNode(cfg, logger)
	if err != nil {
		return fmt.Errorf("error in creating node: %w", err)
	}
	fmt.Fprintln(os.Stderr, "Node created.")

	if flags.skipStart {
		fmt.Fprintln(os.Stderr, "'--skip-start' is set. Exiting.")
		return nil
	}

	if err := gnoNode.Start(); err != nil {
		return fmt.Errorf("error in start node: %w", err)
	}

	// run forever
	osm.TrapSignal(func() {
		if gnoNode.IsRunning() {
			_ = gnoNode.Stop()
		}
	})
	select {} // run forever
}

// Makes a local test genesis doc with local privValidator.
func makeGenesisDoc(pvPub crypto.PubKey) *bft.GenesisDoc {
	gen := &bft.GenesisDoc{}
	gen.GenesisTime = time.Now()
	gen.ChainID = flags.chainID
	gen.ConsensusParams = abci.ConsensusParams{
		Block: &abci.BlockParams{
			// TODO: update limits.
			MaxTxBytes:   1000000,  // 1MB,
			MaxDataBytes: 2000000,  // 2MB,
			MaxGas:       10000000, // 10M gas
			TimeIotaMS:   100,      // 100ms
		},
	}
	gen.Validators = []bft.GenesisValidator{
		{
			Address: pvPub.Address(),
			PubKey:  pvPub,
			Power:   10,
			Name:    "testvalidator",
		},
	}
	// Define genesis balances.
	test1 := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	// Maximum: 1,000,000,000 GNOTs.
	balances := []string{
		// Proposed distribution:
		//  * ATOM_AIRDROP =           750000000000000ugnot 75%
		//  * NEW_TENDERMINT_OR_JAE =  100000000000000ugnot 10%
		//  * CORE_MISSION_DAO =       100000000000000ugnot 10%
		//  * EARLY_CONTRIBUTORS =      50000000000000ugnot  5%
		//
		// NOTES:
		//  - Prop69 YES voters & ICF slashed from ATOM_AIRDROP
		//  - Prop69 NO & NWV rewarded (see PEACE.md) from ATOM_AIRDROP
		//  - NEW_TENDERMINT_JAE reduced by amount of AIB premine going to NewTendermint, to make total equal 10% (Jae's allocation will be transparent)
		//
		// EARLY_CONTRIBUTORS: 50M GNOT, of which:
		//  - 10K test1 (temporary)
		//  - 100K test2 (temporary)
		//  - 2M  faucet0 and faucet1 (temporary)
		//  - 300K contributor (TODO, add more)
		//  - 45K request requested from github

		// Test accounts
		"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=10000000000ugnot",  // test1
		"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj=100000000000ugnot", // test2
		// Faucet accounts.
		"g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv=1000000000000ugnot", // faucet0 (jae)
		"g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa=1000000000000ugnot", // faucet1 (moul)
		// Contributors premine (TODO).
		"g15gdm49ktawvkrl88jadqpucng37yxutucuwaef=100000000000ugnot", // @chadwick
		"g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq=100000000000ugnot", // @moul
		"g14da4n9hcynyzz83q607uu8keuh9hwlv42ra6fa=100000000000ugnot", // @piux2
		// "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj=100000000000ugnot", // @difranco
		// Requests from Github.
		"g1589c8cekvmjfmy0qrd4f3z52r7fn7rgk02667s=1000000000ugnot", // @mefodica #83
		"g13sm84nuqed3fuank8huh7x9mupgw22uft3lcl8=1000000000ugnot", // @hipsterhead91 #81
		"g1m6732pkrngu9vrt0g7056lvr9kcqc4mv83xl5q=1000000000ugnot", // @paironcorp #80
		"g1wg88rhzlwxjd2z4j5de5v5xq30dcf6rjq3dhsj=1000000000ugnot", // @dmvrt #79
		"g18pmaskasz7mxj6rmgrl3al58xu45a7w0l5nmc0=1000000000ugnot", // @asverty #78
		"g19wwhkmqlns70604ksp6rkuuu42qhtvyh05lffz=1000000000ugnot", // @litvintech #77
		"g187982000zsc493znqt828s90cmp6hcp2erhu6m=1000000000ugnot", // @mt2721 #76
		"g1ndpsnrspdnauckytvkfv8s823t3gmpqmtky8pl=1000000000ugnot", // @rk-cosmostation #75
		"g16ja66d65emkr0zxd2tu7xjvm7utthyhpej0037=1000000000ugnot", // @tymoxa #74
		"g1ds24jj9kqjcskd0gzu24r9e4n62ggye230zuv5=1000000000ugnot", // @anarcher #72
		"g1trkzq75ntamsnw9xnrav2v7gy2lt5g6p29yhdr=1000000000ugnot", // @sontrinh16 #71
		"g1rrf8s5mrmu00sx04fzfsvc399fklpeg2x0a7mz=1000000000ugnot", // @masterpi-2124 #70
		"g19p5ntfvpt4lwq4jqsmnxsnelhf3tff9scy3w8w=1000000000ugnot", // @mgialong215 #69
		"g1tue8l73d6rq4vhqdsp2sr3zhuzpure3k2rnwpz=1000000000ugnot", // @nguyenvuong1122000 #68
		"g14hhsss4ngx5kq77je5g0tl4vftg8qp45ceadk3=1000000000ugnot", // @darksoulcrypto #67
		"g1768hvkh7anhd40ch4h7jdh6j3mpcs7hrat4gl0=1000000000ugnot", // @faddat #66
		"g15fa8kyjhu88t9dr8zzua8fwdvkngv5n8yqsm0n=1000000000ugnot", // @emink07 #65
		"g1xhccdjcscuhgmt3quww6qdy3j3czqt3urc2eac=1000000000ugnot", // @muratcicek1 #64
		"g1z629z04f85k4t5gnkk5egpxw9tqxeec435esap=1000000000ugnot", // @blackhox #63
		"g1pfldkplz9puq0v82lu9vqcve9nwrxuq9qe5ttv=1000000000ugnot", // @mihailshumilov #62
		"g152pn0g5qfgxr7yx8zlwjq48hytkafd8x7egsfv=1000000000ugnot", // @ivan123-ops #61
		"g1cf2ye686ke38vjyqakreprljum4xu6rwf5jskq=1000000000ugnot", // @cryptobtcbuyer #60
		"g1c5shztyaj4gjrc5zlwmh9xhex5w7l4asffs2w6=1000000000ugnot", // @mikhailradusha #59
		"g1lhpx2ktk0ha3qw42raxq4m24a4c4xqxyrgv54q=1000000000ugnot", // @danil00524 #58
		"g1026p54q0j902059sm2zsv37krf0ghcl7gmhyv7=1000000000ugnot", // @sashamaxymchuk #57
		"g1n4yvwnv77frq2ccuw27dmtjkd7u4p4jg0pgm7k=1000000000ugnot", // @nook0ne #56
		"g13m7f2e6r3lh3ykxupacdt9sem2tlvmaamwjhll=1000000000ugnot", // @rickau123 #53
		"g19uxluuecjlsqvwmwu8sp6pxaaqfhk972q975xd=1000000000ugnot", // @habibrr #52
		"g1j80fpcsumfkxypvydvtwtz3j4sdwr8c2u0lr64=1000000000ugnot", // @danny-pham #51
		"g1tjdpptuk9eysq6z38nscqyycr998xjyx3w8jvw=1000000000ugnot", // @nhhtrung #50
		"g19t3n89slfemgd3mwuat4lajwcp0yxrkadgeg7a=1000000000ugnot", // @itisnullable #49
		"g1yqndt8xx92l9h494jfruz2w79swzjes3n4wqjc=1000000000ugnot", // @caojs #48
		"g13278z0a5ufeg80ffqxpda9dlp599t7ekregcy6=1000000000ugnot", // @alstn3726 #47
		"g1ht236wjd83x96uqwh9rh3fq6pylyn78mtwq9v6=1000000000ugnot", // @soaryong-c #46
		"g1fj9jccm3zjnqspq7lp2g7lj4czyfq0s35600g9=1000000000ugnot", // @piux2 #45
		"g1wwppuzdns5u6c6jqpkzua24zh6ppsus6399cea=1000000000ugnot", // @spacepotahto #44
		"g1k8pjnguyu36pkc8hy0ufzgpzfmj2jl78la7ek3=1000000000ugnot", // @rhinostake #43
		"g1e8umkzumtxgs8399lw0us4rclea3xl5gxy9spp=1000000000ugnot", // @imperator-co #42
		"g14qekdkj2nmmwea4ufg9n002a3pud23y8k7ugs5=1000000000ugnot", // @dylanschultzie #41
		"g19w2488ntfgpduzqq3sk4j5x387zynwknqdvjqf=1000000000ugnot", // @nullnames #40
		"g1495y3z7zrej4rendysnw5kaeu4g3d7x7w0734g=1000000000ugnot", // @dimasik #85
		"g1hygx8ga9qakhkczyrzs9drm8j8tu4qds9y5e3r=1000000000ugnot", // @zoynitskiy #94
		"g1f977l6wxdh3qu60kzl75vx2wmzswu68l03r8su=1000000000ugnot", // @catShaark #92
		"g1644qje5rx6jsdqfkzmgnfcegx4dxkjh6rwqd69=1000000000ugnot", // @MadafakAvril14th #91
		"g1mzjajymvmtksdwh3wkrndwj6zls2awl9q83dh6=1000000000ugnot", // @Vanlee #89

		// NOTE: Thanks guys, no more keys through genesis this way though:
		// going forward we will have a faucet so anyone can get
		// tokens to pay the spam-prevention tx fee.
	}

	// Load distribution.
	if flags.airdropFile != "" {
		airdrop := loadAirdrop(flags.airdropFile)
		balances = append(balances, airdrop...)
	}

	// Load initial packages from examples.
	txs := []std.Tx{}
	for _, path := range []string{
		"p/ufmt",
		"p/avl",
		"p/grc/grc20",
		"p/grc/grc20/impl",
		"p/grc/grc721",
		"p/maths",
		"r/users",
		"r/foo20",
		"r/boards",
		"r/banktest",
	} {
		// open files in directory as MemPackage.
		memPkg := gno.ReadMemPackage(filepath.Join(".", "examples", "gno.land", path), "gno.land/"+path)
		var tx std.Tx
		tx.Msgs = []std.Msg{
			vmm.MsgAddPackage{
				Creator: test1,
				Package: memPkg,
				Deposit: nil,
			},
		}
		tx.Fee = std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
		txs = append(txs, tx)
	}
	// load genesis txs from file.
	txsBz := osm.MustReadFile("./gnoland/genesis/genesis_txs.txt")
	// txsBz := osm.MustReadFile("./txexport.log.16")
	txsLines := strings.Split(string(txsBz), "\n")
	for _, txLine := range txsLines {
		if txLine == "" {
			continue // skip empty line
		}

		// patch the TX
		txLine = strings.ReplaceAll(txLine, "%%CHAINID%%", flags.chainID)
		txLine = strings.ReplaceAll(txLine, "%%REMOTE%%", flags.genesisRemote)

		var tx std.Tx
		amino.MustUnmarshalJSON([]byte(txLine), &tx)
		txs = append(txs, tx)
	}
	// construct genesis AppState.
	gen.AppState = gnoland.GnoGenesisState{
		Balances: balances,
		Txs:      txs,
	}
	return gen
}

func writeGenesisFile(gen *bft.GenesisDoc, filePath string) {
	err := gen.SaveAs(filePath)
	if err != nil {
		panic(err)
	}
}

func loadAirdrop(airdropFile string) []string {
	bz := osm.MustReadFile(airdropFile)
	line := strings.TrimSuffix(string(bz), "\n")
	balances := strings.Split(line, "\n")

	for i, v := range balances {
		// cosmos10008uvk6fj3ja05u092ya5sx6fn355wavael4j:g10008uvk6fj3ja05u092ya5sx6fn355walp9u5k=3204884ugnot
		// split and drop cosmos address.
		a := strings.Split(v, ":")
		parts := strings.Split(a[1], "=")
		if len(parts) != 2 {
			fmt.Printf("error: %v\n", a)
		}
		balances[i] = a[1]

	}
	return balances
}
