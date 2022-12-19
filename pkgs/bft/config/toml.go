package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	osm "github.com/gnolang/gno/pkgs/os"
	"github.com/pelletier/go-toml"
)

// DefaultDirPerm is the default permissions used when creating directories.
const DefaultDirPerm = 0o700

var configTemplate *template.Template

func init() {
	var err error
	if configTemplate, err = template.New("configFileTemplate").Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

func LoadConfigFile(configFilePath string) *Config {
	bz, err := os.ReadFile(configFilePath)
	if err != nil {
		panic(err)
	}
	var config Config
	err = toml.Unmarshal(bz, &config)
	if err != nil {
		panic(err)
	}
	return &config
}

/****** these are for production settings ***********/

// WriteConfigFile renders config using the template and writes it to configFilePath.
func WriteConfigFile(configFilePath string, config *Config) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	osm.MustWriteFile(configFilePath, buffer.Bytes(), 0o644)
}

// Note: any changes to the comments/variables/field-names
// must be reflected in the appropriate struct in config/config.go
const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# TCP or UNIX socket address of the ABCI application,
# or the name of an ABCI application compiled in with the Tendermint binary
proxy_app = "{{ .BaseConfig.ProxyApp }}"

# A custom human readable name for this node
moniker = "{{ .BaseConfig.Moniker }}"

# If this node is many blocks behind the tip of the chain, FastSync
# allows them to catchup quickly by downloading blocks in parallel
# and verifying their commits
fast_sync = {{ .BaseConfig.FastSyncMode }}

# Database backend: goleveldb | cleveldb | boltdb
# * goleveldb (github.com/syndtr/goleveldb - most popular implementation)
#   - pure go
#   - stable
# * cleveldb (uses levigo wrapper)
#   - fast
#   - requires gcc
#   - use cleveldb build tag (go build -tags cleveldb)
# * boltdb (uses etcd's fork of bolt - go.etcd.io/bbolt)
#   - EXPERIMENTAL
#   - may be faster is some use-cases (random reads - indexer)
#   - use boltdb build tag (go build -tags boltdb)
db_backend = "{{ .BaseConfig.DBBackend }}"

# Database directory
db_dir = "{{ js .BaseConfig.DBPath }}"

# Output level for logging, including package level options
log_level = "{{ .BaseConfig.LogLevel }}"

# Output format: 'plain' (colored text) or 'json'
log_format = "{{ .BaseConfig.LogFormat }}"

##### additional base config options #####

# Path to the JSON file containing the initial validator set and other meta data
genesis_file = "{{ js .BaseConfig.Genesis }}"

# Path to the JSON file containing the private key to use as a validator in the consensus protocol
priv_validator_key_file = "{{ js .BaseConfig.PrivValidatorKey }}"

# Path to the JSON file containing the last sign state of a validator
priv_validator_state_file = "{{ js .BaseConfig.PrivValidatorState }}"

# TCP or UNIX socket address for Tendermint to listen on for
# connections from an external PrivValidator process
priv_validator_laddr = "{{ .BaseConfig.PrivValidatorListenAddr }}"

# Path to the JSON file containing the private key to use for node authentication in the p2p protocol
node_key_file = "{{ js .BaseConfig.NodeKey }}"

# Mechanism to connect to the ABCI application: socket | grpc
abci = "{{ .BaseConfig.ABCI }}"

# TCP or UNIX socket address for the profiling server to listen on
prof_laddr = "{{ .BaseConfig.ProfListenAddress }}"

# If true, query the ABCI app on connecting to a new peer
# so the app can decide if we should keep the connection or not
filter_peers = {{ .BaseConfig.FilterPeers }}

##### advanced configuration options #####

##### rpc server configuration options #####
[rpc]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "{{ .RPC.ListenAddress }}"

# A list of origins a cross-domain request can be executed from
# Default value '[]' disables cors support
# Use '["*"]' to allow any origin
cors_allowed_origins = [{{ range .RPC.CORSAllowedOrigins }}{{ printf "%q, " . }}{{end}}]

# A list of methods the client is allowed to use with cross-domain requests
cors_allowed_methods = [{{ range .RPC.CORSAllowedMethods }}{{ printf "%q, " . }}{{end}}]

# A list of non simple headers the client is allowed to use with cross-domain requests
cors_allowed_headers = [{{ range .RPC.CORSAllowedHeaders }}{{ printf "%q, " . }}{{end}}]

# TCP or UNIX socket address for the gRPC server to listen on
# NOTE: This server only supports /broadcast_tx_commit
grpc_laddr = "{{ .RPC.GRPCListenAddress }}"

# Maximum number of simultaneous connections.
# Does not include RPC (HTTP&WebSocket) connections. See max_open_connections
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
# Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}
# 1024 - 40 - 10 - 50 = 924 = ~900
grpc_max_open_connections = {{ .RPC.GRPCMaxOpenConnections }}

# Activate unsafe RPC commands like /dial_seeds and /unsafe_flush_mempool
unsafe = {{ .RPC.Unsafe }}

# Maximum number of simultaneous connections (including WebSocket).
# Does not include gRPC connections. See grpc_max_open_connections
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
# Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}
# 1024 - 40 - 10 - 50 = 924 = ~900
max_open_connections = {{ .RPC.MaxOpenConnections }}

# How long to wait for a tx to be committed during /broadcast_tx_commit.
# WARNING: Using a value larger than 10s will result in increasing the
# global HTTP write timeout, which applies to all connections and endpoints.
# See https://github.com/tendermint/classic/issues/3435
timeout_broadcast_tx_commit = "{{ .RPC.TimeoutBroadcastTxCommit }}"

# Maximum size of request body, in bytes
max_body_bytes = {{ .RPC.MaxBodyBytes }}

# Maximum size of request header, in bytes
max_header_bytes = {{ .RPC.MaxHeaderBytes }}

# The path to a file containing certificate that is used to create the HTTPS server.
# Migth be either absolute path or path related to tendermint's config directory.
# If the certificate is signed by a certificate authority,
# the certFile should be the concatenation of the server's certificate, any intermediates,
# and the CA's certificate.
# NOTE: both tls_cert_file and tls_key_file must be present for Tendermint to create HTTPS server. Otherwise, HTTP server is run.
tls_cert_file = "{{ .RPC.TLSCertFile }}"

# The path to a file containing matching private key that is used to create the HTTPS server.
# Migth be either absolute path or path related to tendermint's config directory.
# NOTE: both tls_cert_file and tls_key_file must be present for Tendermint to create HTTPS server. Otherwise, HTTP server is run.
tls_key_file = "{{ .RPC.TLSKeyFile }}"

##### peer to peer configuration options #####
[p2p]

# Address to listen for incoming connections
laddr = "{{ .P2P.ListenAddress }}"

# Address to advertise to peers for them to dial
# If empty, will use the same port as the laddr,
# and will introspect on the listener or use UPnP
# to figure out the address.
external_address = "{{ .P2P.ExternalAddress }}"

# Comma separated list of seed nodes to connect to
seeds = "{{ .P2P.Seeds }}"

# Comma separated list of nodes to keep persistent connections to
persistent_peers = "{{ .P2P.PersistentPeers }}"

# UPNP port forwarding
upnp = {{ .P2P.UPNP }}

# Maximum number of inbound peers
max_num_inbound_peers = {{ .P2P.MaxNumInboundPeers }}

# Maximum number of outbound peers to connect to, excluding persistent peers
max_num_outbound_peers = {{ .P2P.MaxNumOutboundPeers }}

# Time to wait before flushing messages out on the connection
flush_throttle_timeout = "{{ .P2P.FlushThrottleTimeout }}"

# Maximum size of a message packet payload, in bytes
max_packet_msg_payload_size = {{ .P2P.MaxPacketMsgPayloadSize }}

# Rate at which packets can be sent, in bytes/second
send_rate = {{ .P2P.SendRate }}

# Rate at which packets can be received, in bytes/second
recv_rate = {{ .P2P.RecvRate }}

# Set true to enable the peer-exchange reactor
pex = {{ .P2P.PexReactor }}

# Seed mode, in which node constantly crawls the network and looks for
# peers. If another node asks it for addresses, it responds and disconnects.
#
# Does not work if the peer-exchange reactor is disabled.
seed_mode = {{ .P2P.SeedMode }}

# Comma separated list of peer IDs to keep private (will not be gossiped to other peers)
private_peer_ids = "{{ .P2P.PrivatePeerIDs }}"

# Toggle to disable guard against peers connecting from the same ip.
allow_duplicate_ip = {{ .P2P.AllowDuplicateIP }}

# Peer connection configuration.
handshake_timeout = "{{ .P2P.HandshakeTimeout }}"
dial_timeout = "{{ .P2P.DialTimeout }}"

##### mempool configuration options #####
[mempool]

recheck = {{ .Mempool.Recheck }}
broadcast = {{ .Mempool.Broadcast }}
wal_dir = "{{ js .Mempool.WalPath }}"

# Maximum number of transactions in the mempool
size = {{ .Mempool.Size }}

# Limit the total size of all txs in the mempool.
# This only accounts for raw transactions (e.g. given 1MB transactions and
# max_txs_bytes=5MB, mempool will only accept 5 transactions).
max_pending_txs_bytes = {{ .Mempool.MaxPendingTxsBytes }}

# Size of the cache (used to filter transactions we saw earlier) in transactions
cache_size = {{ .Mempool.CacheSize }}

##### consensus configuration options #####
[consensus]

wal_file = "{{ js .Consensus.WalPath }}"

timeout_propose = "{{ .Consensus.TimeoutPropose }}"
timeout_propose_delta = "{{ .Consensus.TimeoutProposeDelta }}"
timeout_prevote = "{{ .Consensus.TimeoutPrevote }}"
timeout_prevote_delta = "{{ .Consensus.TimeoutPrevoteDelta }}"
timeout_precommit = "{{ .Consensus.TimeoutPrecommit }}"
timeout_precommit_delta = "{{ .Consensus.TimeoutPrecommitDelta }}"
timeout_commit = "{{ .Consensus.TimeoutCommit }}"

# Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)
skip_timeout_commit = {{ .Consensus.SkipTimeoutCommit }}

# EmptyBlocks mode and possible interval between empty blocks
create_empty_blocks = {{ .Consensus.CreateEmptyBlocks }}
create_empty_blocks_interval = "{{ .Consensus.CreateEmptyBlocksInterval }}"

# Reactor sleep duration parameters
peer_gossip_sleep_duration = "{{ .Consensus.PeerGossipSleepDuration }}"
peer_query_maj23_sleep_duration = "{{ .Consensus.PeerQueryMaj23SleepDuration }}"
`

/****** these are for test settings ***********/

func ResetTestRoot(testName string) *Config {
	return ResetTestRootWithChainID(testName, "")
}

func ResetTestRootWithChainID(testName string, chainID string) *Config {
	// create a unique, concurrency-safe test directory under os.TempDir()
	rootDir, err := ioutil.TempDir("", fmt.Sprintf("%s-%s_", chainID, testName))
	if err != nil {
		panic(err)
	}
	// ensure config and data subdirs are created
	if err := osm.EnsureDir(filepath.Join(rootDir, defaultConfigDir), DefaultDirPerm); err != nil {
		panic(err)
	}
	if err := osm.EnsureDir(filepath.Join(rootDir, defaultDataDir), DefaultDirPerm); err != nil {
		panic(err)
	}

	baseConfig := DefaultBaseConfig()
	configFilePath := filepath.Join(rootDir, defaultConfigFilePath)
	genesisFilePath := filepath.Join(rootDir, baseConfig.Genesis)
	privKeyFilePath := filepath.Join(rootDir, baseConfig.PrivValidatorKey)
	privStateFilePath := filepath.Join(rootDir, baseConfig.PrivValidatorState)

	// Write default config file if missing.
	if !osm.FileExists(configFilePath) {
		WriteConfigFile(configFilePath, DefaultConfig())
	}
	if !osm.FileExists(genesisFilePath) {
		if chainID == "" {
			chainID = "tendermint_test"
		}
		testGenesis := fmt.Sprintf(testGenesisFmt, chainID)
		osm.MustWriteFile(genesisFilePath, []byte(testGenesis), 0o644)
	}
	// we always overwrite the priv val
	osm.MustWriteFile(privKeyFilePath, []byte(testPrivValidatorKey), 0o644)
	osm.MustWriteFile(privStateFilePath, []byte(testPrivValidatorState), 0o644)

	config := TestConfig().SetRootDir(rootDir)
	return config
}

var testGenesisFmt = `{
  "genesis_time": "2018-10-10T08:20:13.695936996Z",
  "chain_id": "%s",
  "validators": [
    {
      "pub_key": {
        "@type": "/tm.PubKeyEd25519",
        "value": "cVt6w3C1DWYwwkAirnbsL49CoOe8T8ZR2BCB8MeOGRg="
      },
      "power": "10",
      "name": ""
    }
  ],
  "app_hash": ""
}`

var testPrivValidatorKey = `{
  "address": "g1uvwz22t0l2fv9az93wutmlusrjv5zdwx2n32d5",
  "pub_key": {
    "@type": "/tm.PubKeyEd25519",
    "value": "cVt6w3C1DWYwwkAirnbsL49CoOe8T8ZR2BCB8MeOGRg="
  },
  "priv_key": {
    "@type": "/tm.PrivKeyEd25519",
    "value": "Qq4Q9QH2flPSIJShbXPIocbrQtQ4S7Kdn31uI3sKZoJxW3rDcLUNZjDCQCKuduwvj0Kg57xPxlHYEIHwx44ZGA=="
  }
}`

var testPrivValidatorState = `{
  "height": "0",
  "round": "0",
  "step": 0
}`
