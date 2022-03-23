package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/bft/config"
	"github.com/gnolang/gno/pkgs/bft/node"
	"github.com/gnolang/gno/pkgs/bft/privval"
	bft "github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/log"
	osm "github.com/gnolang/gno/pkgs/os"
)

func main() {
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
	gnoApp, err := gnoland.NewApp(rootDir, logger)
	if err != nil {
		panic(fmt.Sprintf("error in creating new app: %v", err))
	}
	cfg.LocalApp = gnoApp
	gnoNode, err := node.DefaultNewNode(cfg, logger)
	if err != nil {
		panic(fmt.Sprintf("error in creating node: %v", err))
	}
	if err := gnoNode.Start(); err != nil {
		panic(fmt.Sprintf("error in start node: %v", err))
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
	gen.ChainID = "testchain"
	gen.Validators = []bft.GenesisValidator{
		bft.GenesisValidator{
			Address: pvPub.Address(),
			PubKey:  pvPub,
			Power:   10,
			Name:    "testvalidator",
		},
	}
	gen.AppState = gnoland.GnoGenesisState{
		Balances: []string{
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=10000gnot",
			"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj=1000000000gnot",
			// from Github:
			"g1589c8cekvmjfmy0qrd4f3z52r7fn7rgk02667s=10000gnot", // @mefodica #83
			"g13sm84nuqed3fuank8huh7x9mupgw22uft3lcl8=10000gnot", // @hipsterhead91 #81
			"g1m6732pkrngu9vrt0g7056lvr9kcqc4mv83xl5q=10000gnot", // @paironcorp #80
			"g1wg88rhzlwxjd2z4j5de5v5xq30dcf6rjq3dhsj=10000gnot", // @dmvrt #79
			"g18pmaskasz7mxj6rmgrl3al58xu45a7w0l5nmc0=10000gnot", // @asverty #78
			"g19wwhkmqlns70604ksp6rkuuu42qhtvyh05lffz=10000gnot", // @litvintech #77
			"g187982000zsc493znqt828s90cmp6hcp2erhu6m=10000gnot", // @mt2721 #76
			"g1ndpsnrspdnauckytvkfv8s823t3gmpqmtky8pl=10000gnot", // @rk-cosmostation #75
			"g16ja66d65emkr0zxd2tu7xjvm7utthyhpej0037=10000gnot", // @tymoxa #74
			"g1ds24jj9kqjcskd0gzu24r9e4n62ggye230zuv5=10000gnot", // @anarcher #72
			"g1trkzq75ntamsnw9xnrav2v7gy2lt5g6p29yhdr=10000gnot", // @sontrinh16 #71
			"g1rrf8s5mrmu00sx04fzfsvc399fklpeg2x0a7mz=10000gnot", // @masterpi-2124 #70
			"g19p5ntfvpt4lwq4jqsmnxsnelhf3tff9scy3w8w=10000gnot", // @mgialong215 #69
			"g1tue8l73d6rq4vhqdsp2sr3zhuzpure3k2rnwpz=10000gnot", // @nguyenvuong1122000 #68
			"g14hhsss4ngx5kq77je5g0tl4vftg8qp45ceadk3=10000gnot", // @darksoulcrypto #67
			"g1768hvkh7anhd40ch4h7jdh6j3mpcs7hrat4gl0=10000gnot", // @faddat #66
			"g15fa8kyjhu88t9dr8zzua8fwdvkngv5n8yqsm0n=10000gnot", // @emink07 #65
			"g1xhccdjcscuhgmt3quww6qdy3j3czqt3urc2eac=10000gnot", // @muratcicek1 #64
			"g1z629z04f85k4t5gnkk5egpxw9tqxeec435esap=10000gnot", // @blackhox #63
			"g1pfldkplz9puq0v82lu9vqcve9nwrxuq9qe5ttv=10000gnot", // @mihailshumilov #62
			"g152pn0g5qfgxr7yx8zlwjq48hytkafd8x7egsfv=10000gnot", // @ivan123-ops #61
			"g1cf2ye686ke38vjyqakreprljum4xu6rwf5jskq=10000gnot", // @cryptobtcbuyer #60
			"g1c5shztyaj4gjrc5zlwmh9xhex5w7l4asffs2w6=10000gnot", // @mikhailradusha #59
			"g1lhpx2ktk0ha3qw42raxq4m24a4c4xqxyrgv54q=10000gnot", // @danil00524 #58
			"g1026p54q0j902059sm2zsv37krf0ghcl7gmhyv7=10000gnot", // @sashamaxymchuk #57
			"g1n4yvwnv77frq2ccuw27dmtjkd7u4p4jg0pgm7k=10000gnot", // @nook0ne #56
			"g13m7f2e6r3lh3ykxupacdt9sem2tlvmaamwjhll=10000gnot", // @rickau123 #53
			"g19uxluuecjlsqvwmwu8sp6pxaaqfhk972q975xd=10000gnot", // @habibrr #52
			"g1j80fpcsumfkxypvydvtwtz3j4sdwr8c2u0lr64=10000gnot", // @danny-pham #51
			"g1tjdpptuk9eysq6z38nscqyycr998xjyx3w8jvw=10000gnot", // @nhhtrung #50
			"g19t3n89slfemgd3mwuat4lajwcp0yxrkadgeg7a=10000gnot", // @itisnullable #49
			"g1yqndt8xx92l9h494jfruz2w79swzjes3n4wqjc=10000gnot", // @caojs #48
			"g13278z0a5ufeg80ffqxpda9dlp599t7ekregcy6=10000gnot", // @alstn3726 #47
			"g1ht236wjd83x96uqwh9rh3fq6pylyn78mtwq9v6=10000gnot", // @soaryong-c #46
			"g1fj9jccm3zjnqspq7lp2g7lj4czyfq0s35600g9=10000gnot", // @piux2 #45
			"g1wwppuzdns5u6c6jqpkzua24zh6ppsus6399cea=10000gnot", // @spacepotahto #44
			"g1k8pjnguyu36pkc8hy0ufzgpzfmj2jl78la7ek3=10000gnot", // @rhinostake #43
			"g1e8umkzumtxgs8399lw0us4rclea3xl5gxy9spp=10000gnot", // @imperator-co #42
			"g14qekdkj2nmmwea4ufg9n002a3pud23y8k7ugs5=10000gnot", // @dylanschultzie #41
			"g19w2488ntfgpduzqq3sk4j5x387zynwknqdvjqf=10000gnot", // @nullnames #40
			"g1495y3z7zrej4rendysnw5kaeu4g3d7x7w0734g=10000gnot", // @dimasik #85
			"g1hygx8ga9qakhkczyrzs9drm8j8tu4qds9y5e3r=10000gnot", // @zoynitskiy #94
			"g1f977l6wxdh3qu60kzl75vx2wmzswu68l03r8su=10000gnot", // @catShaark #92
			"g1644qje5rx6jsdqfkzmgnfcegx4dxkjh6rwqd69=10000gnot", // @MadafakAvril14th #91
			"g1mzjajymvmtksdwh3wkrndwj6zls2awl9q83dh6=10000gnot", // @Vanlee #89
			// NOTE: Thanks guys, no more keys through genesis this way though:
			// going forward we will have an ATOM-fee based faucet so anyone can get
			// faucet tokens by making a spam-prevention transaction.
			// TODO: make it a bounty item.
		},
	}
	return gen
}

func writeGenesisFile(gen *bft.GenesisDoc, filePath string) {
	err := gen.SaveAs(filePath)
	if err != nil {
		panic(err)
	}
}
