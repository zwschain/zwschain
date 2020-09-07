// Copyright 2015 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// Package utils contains internal helper functions for go-ethereum commands.
package utils

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"

	"os"
	"path/filepath"
	"runtime"
	"strings"

	"encoding/hex"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/fdlimit"
	"github.com/palletone/go-palletone/common/files"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/common/p2p/nat"
	"github.com/palletone/go-palletone/common/p2p/netutil"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/state"
	"github.com/palletone/go-palletone/light"
	"github.com/palletone/go-palletone/light/cors"
	"github.com/palletone/go-palletone/ptn"
	"github.com/palletone/go-palletone/ptn/downloader"
	"github.com/palletone/go-palletone/statistics/dashboard"
	"github.com/palletone/go-palletone/statistics/metrics"
	"github.com/palletone/go-palletone/statistics/metrics/prometheus"
	"github.com/palletone/go-palletone/txspool"
	"gopkg.in/urfave/cli.v1"
)

var (
	CommandHelpTemplate = `{{.cmd.Name}}{{if .cmd.Subcommands}} command{{end}}{{if .cmd.Flags}} 
[command options]{{end}} [arguments...] {{if .cmd.Description}}{{.cmd.Description}} {{end}}{{if .cmd.Subcommands}}
SUBCOMMANDS:
	{{range .cmd.Subcommands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
	{{end}}{{end}}{{if .categorizedFlags}}
{{range $idx, $categorized := .categorizedFlags}}{{$categorized.Name}} OPTIONS:
{{range $categorized.Flags}}{{"\t"}}{{.}}
{{end}}
{{end}}{{end}}`
)

func init() {
	cli.AppHelpTemplate = `{{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} 
	[command options]{{end}} [arguments...]

VERSION:
   {{.Version}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

	cli.CommandHelpTemplate = CommandHelpTemplate
}

// NewApp creates an app with sane defaults.
func NewApp(gitCommit, usage string) *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = ""
	//app.Authors = nil
	app.Email = ""
	app.Version = configure.Version
	if len(gitCommit) >= 8 {
		app.Version += "-" + gitCommit[:8]
	}
	app.Usage = usage
	return app
}

// These are all the command line flags we support.
// If you add to this list, please remember to include the
// flag in the appropriate command definition.
//
// The flags are defined here so their names and help texts
// are the same for all commands.

var (
	// General settings
	DataDirFlag = DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: DirectoryString{""}, //DirectoryString{node.DefaultDataDir()},
	}
	KeyStoreDirFlag = DirectoryFlag{
		Name:  "keystore",
		Usage: "Directory for the keystore (default = inside the datadir)",
	}
	NoUSBFlag = cli.BoolFlag{
		Name:  "nousb",
		Usage: "Disables monitoring for and managing USB hardware wallets",
	}
	NetworkIdFlag = cli.Uint64Flag{
		Name:  "networkid",
		Usage: "Network identifier (integer, 1=Frontier, 2=Morden (disused), 3=Ropsten, 4=Rinkeby)",
		Value: ptn.DefaultConfig.NetworkId,
	}
	TestnetFlag = cli.BoolFlag{
		Name:  "testnet",
		Usage: "Ropsten network: pre-configured proof-of-work test network",
	}
	DeveloperFlag = cli.BoolFlag{
		Name:  "dev",
		Usage: "Ephemeral proof-of-authority network with a pre-funded developer account, mining enabled",
	}
	DeveloperPeriodFlag = cli.IntFlag{
		Name:  "dev.period",
		Usage: "Block period to use in developer mode (0 = mine only if transaction pending)",
	}
	IdentityFlag = cli.StringFlag{
		Name:  "identity",
		Usage: "Custom node name",
	}
	DocRootFlag = DirectoryFlag{
		Name:  "docroot",
		Usage: "Document Root for HTTPClient file scheme",
		Value: DirectoryString{homeDir()},
	}
	FastSyncFlag = cli.BoolFlag{
		Name:  "fast",
		Usage: "Enable fast syncing through state downloads (replaced by --syncmode)",
	}
	LightModeFlag = cli.BoolFlag{
		Name:  "light",
		Usage: "Enable light client mode (replaced by --syncmode)",
	}
	defaultSyncMode = ptn.DefaultConfig.SyncMode
	SyncModeFlag    = TextMarshalerFlag{
		Name:  "syncmode",
		Usage: `Blockchain sync mode ("fast", "full", or "light")`,
		Value: &defaultSyncMode,
	}
	GCModeFlag = cli.StringFlag{
		Name:  "gcmode",
		Usage: `Blockchain garbage collection mode ("full", "archive")`,
		Value: "full",
	}
	LightServFlag = cli.IntFlag{
		Name:  "lightserv",
		Usage: "Maximum percentage of time allowed for serving LES requests (0-90)",
		Value: 0,
	}
	LightPeersFlag = cli.IntFlag{
		Name:  "lightpeers",
		Usage: "Maximum number of LES client peers",
		Value: ptn.DefaultConfig.LightPeers,
	}
	LightKDFFlag = cli.BoolFlag{
		Name:  "lightkdf",
		Usage: "Reduce key-derivation RAM & CPU usage at some expense of KDF strength",
	}
	// Dashboard settings
	DashboardEnabledFlag = cli.BoolFlag{
		Name:  "dashboard",
		Usage: "Enable the dashboard",
	}
	DashboardAddrFlag = cli.StringFlag{
		Name:  "dashboard.addr",
		Usage: "Dashboard listening interface",
		Value: dashboard.DefaultConfig.Host,
	}
	DashboardPortFlag = cli.IntFlag{
		Name:  "dashboard.host",
		Usage: "Dashboard listening port",
		Value: dashboard.DefaultConfig.Port,
	}
	DashboardRefreshFlag = cli.DurationFlag{
		Name:  "dashboard.refresh",
		Usage: "Dashboard metrics collection refresh rate",
		Value: dashboard.DefaultConfig.Refresh,
	}

	// Transaction pool settings
	TxPoolNoLocalsFlag = cli.BoolFlag{
		Name:  "txpool.nolocals",
		Usage: "Disables price exemptions for locally submitted transactions",
	}
	TxPoolJournalFlag = cli.StringFlag{
		Name:  "txpool.journal",
		Usage: "Disk journal for local transaction to survive node restarts",
		Value: txspool.DefaultTxPoolConfig.Journal,
	}
	TxPoolRejournalFlag = cli.DurationFlag{
		Name:  "txpool.rejournal",
		Usage: "Time interval to regenerate the local transaction journal",
		Value: txspool.DefaultTxPoolConfig.Rejournal,
	}
	TxPoolPriceLimitFlag = cli.Uint64Flag{
		Name:  "txpool.pricelimit",
		Usage: "Minimum gas price limit to enforce for acceptance into the pool",
		//Value: ptn.DefaultConfig.TxPool.PriceLimit,
	}
	TxPoolPriceBumpFlag = cli.Uint64Flag{
		Name:  "txpool.pricebump",
		Usage: "Price bump percentage to replace an already existing transaction",
		Value: ptn.DefaultConfig.TxPool.PriceBump,
	}
	TxPoolGlobalSlotsFlag = cli.Uint64Flag{
		Name:  "txpool.globalslots",
		Usage: "Maximum number of executable transaction slots for all accounts",
		Value: ptn.DefaultConfig.TxPool.GlobalSlots,
	}
	//TxPoolAccountQueueFlag = cli.Uint64Flag{
	//	Name:  "txpool.accountqueue",
	//	Usage: "Maximum number of non-executable transaction slots permitted per account",
	//	Value: ptn.DefaultConfig.TxPool.AccountQueue,
	//}
	TxPoolGlobalQueueFlag = cli.Uint64Flag{
		Name:  "txpool.globalqueue",
		Usage: "Maximum number of non-executable transaction slots for all accounts",
		Value: ptn.DefaultConfig.TxPool.GlobalQueue,
	}
	TxPoolLifetimeFlag = cli.DurationFlag{
		Name:  "txpool.lifetime",
		Usage: "Maximum amount of time non-executable transaction are queued",
		Value: ptn.DefaultConfig.TxPool.Lifetime,
	}
	TxPoolRemovetimeFlag = cli.DurationFlag{
		Name:  "txpool.removetime",
		Usage: "Maximum amount of time txpool transaction are removed",
		Value: ptn.DefaultConfig.TxPool.Removetime,
	}

	// Performance tuning settings
	CacheFlag = cli.IntFlag{
		Name:  "cache",
		Usage: "Megabytes of memory allocated to internal caching",
		Value: 1024,
	}
	CacheDatabaseFlag = cli.IntFlag{
		Name:  "cache.database",
		Usage: "Percentage of cache memory allowance to use for database io",
		Value: 75,
	}
	CacheGCFlag = cli.IntFlag{
		Name:  "cache.gc",
		Usage: "Percentage of cache memory allowance to use for trie pruning",
		Value: 25,
	}
	TrieCacheGenFlag = cli.IntFlag{
		Name:  "trie-cache-gens",
		Usage: "Number of trie node generations to keep in memory",
		Value: int(state.MaxTrieCacheGen),
	}
	// Miner settings
	MiningEnabledFlag = cli.BoolFlag{
		Name:  "mine",
		Usage: "Enable mining",
	}
	MinerThreadsFlag = cli.IntFlag{
		Name:  "minerthreads",
		Usage: "Number of CPU threads to use for mining",
		Value: runtime.NumCPU(),
	}
	//TargetGasLimitFlag = cli.Uint64Flag{
	//	Name:  "targetgaslimit",
	//	Usage: "Target gas limit sets the artificial target gas floor for the blocks to mine",
	//Value: configure.GenesisGasLimit,
	//}
	EtherbaseFlag = cli.StringFlag{
		Name:  "etherbase",
		Usage: "Public address for block mining rewards (default = first account created)",
		Value: "0",
	}
	CryptoLibFlag = cli.StringFlag{
		Name:  "cryptolib",
		Usage: "set crypto lib,1st byte sign algorithm: 0,ECDSA-S256;1,GM-SM2 2ed byte hash algorithm: 0,SHA3;1,GM-SM3",
		Value: hex.EncodeToString(ptn.DefaultConfig.CryptoLib),
	}
	ExtraDataFlag = cli.StringFlag{
		Name:  "extradata",
		Usage: "Block extra data set by the miner (default = client version)",
	}
	// Account settings
	UnlockedAccountFlag = cli.StringFlag{
		Name:  "unlock",
		Usage: "Comma separated list of accounts to unlock",
		Value: "",
	}
	PasswordFileFlag = cli.StringFlag{
		Name:  "password",
		Usage: "Password file to use for non-interactive password input",
		Value: "",
	}

	VMEnableDebugFlag = cli.BoolFlag{
		Name:  "vmdebug",
		Usage: "Record information useful for VM and contract debugging",
	}
	// Logging and debug settings
	EthStatsURLFlag = cli.StringFlag{
		Name:  "ethstats",
		Usage: "Reporting URL of a ethstats service (nodename:secret@host:port)",
	}
	MetricsEnabledFlag = cli.BoolFlag{
		Name:  metrics.MetricsEnabledFlag,
		Usage: "Enable metrics collection and reporting",
	}
	FakePoWFlag = cli.BoolFlag{
		Name:  "fakepow",
		Usage: "Disables proof-of-work verification",
	}
	NoCompactionFlag = cli.BoolFlag{
		Name:  "nocompaction",
		Usage: "Disables db compaction after import",
	}
	// RPC settings
	RPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "Enable the HTTP-RPC server",
	}
	RPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: node.DefaultHTTPHost,
	}
	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: node.DefaultHTTPPort,
	}
	RPCCORSDomainFlag = cli.StringFlag{
		Name:  "rpccorsdomain",
		Usage: "Comma separated list of domains from which to accept cross origin requests (browser enforced)",
		Value: "",
	}
	RPCVirtualHostsFlag = cli.StringFlag{
		Name: "rpcvhosts",
		Usage: "Comma separated list of virtual hostnames from which to accept requests (server enforced). " +
			"Accepts '*' wildcard.",
		Value: strings.Join(node.DefaultConfig.HTTPVirtualHosts, ","),
	}
	RPCApiFlag = cli.StringFlag{
		Name:  "rpcapi",
		Usage: "API's offered over the HTTP-RPC interface",
		Value: "",
	}
	IPCDisabledFlag = cli.BoolFlag{
		Name:  "ipcdisable",
		Usage: "Disable the IPC-RPC server",
	}
	IPCPathFlag = DirectoryFlag{
		Name:  "ipcpath",
		Usage: "Filename for IPC socket/pipe within the datadir (explicit paths escape it)",
	}
	WSEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "Enable the WS-RPC server",
	}
	WSListenAddrFlag = cli.StringFlag{
		Name:  "wsaddr",
		Usage: "WS-RPC server listening interface",
		Value: node.DefaultWSHost,
	}
	WSPortFlag = cli.IntFlag{
		Name:  "wsport",
		Usage: "WS-RPC server listening port",
		Value: node.DefaultWSPort,
	}
	WSApiFlag = cli.StringFlag{
		Name:  "wsapi",
		Usage: "API's offered over the WS-RPC interface",
		Value: "",
	}
	WSAllowedOriginsFlag = cli.StringFlag{
		Name:  "wsorigins",
		Usage: "Origins from which to accept websockets requests",
		Value: "",
	}
	ExecFlag = cli.StringFlag{
		Name:  "exec",
		Usage: "Execute JavaScript statement",
	}
	PreloadJSFlag = cli.StringFlag{
		Name:  "preload",
		Usage: "Comma separated list of JavaScript files to preload into the console",
	}

	// Network Settings
	MaxPeersFlag = cli.IntFlag{
		Name:  "maxpeers",
		Usage: "Maximum number of network peers (network disabled if set to 0)",
		Value: 25,
	}
	MaxPendingPeersFlag = cli.IntFlag{
		Name:  "maxpendpeers",
		Usage: "Maximum number of pending connection attempts (defaults used if set to 0)",
		Value: 0,
	}
	ListenPortFlag = cli.IntFlag{
		Name:  "port",
		Usage: "Network listening port",
		Value: 30303,
	}
	BootnodesFlag = cli.StringFlag{
		Name:  "bootnodes",
		Usage: "Comma separated pnode URLs for P2P discovery bootstrap (set v4+v5 instead for light servers)",
		Value: "",
	}
	BootnodesV4Flag = cli.StringFlag{
		Name:  "bootnodesv4",
		Usage: "Comma separated pnode URLs for P2P v4 discovery bootstrap (light server, full nodes)",
		Value: "",
	}
	BootnodesV5Flag = cli.StringFlag{
		Name:  "bootnodesv5",
		Usage: "Comma separated pnode URLs for P2P v5 discovery bootstrap (light server, light nodes)",
		Value: "",
	}
	NodeKeyFileFlag = cli.StringFlag{
		Name:  "nodekey",
		Usage: "P2P node key file",
	}
	NodeKeyHexFlag = cli.StringFlag{
		Name:  "nodekeyhex",
		Usage: "P2P node key as hex (for testing)",
	}
	NATFlag = cli.StringFlag{
		Name:  "nat",
		Usage: "NAT port mapping mechanism (any|none|upnp|pmp|extip:<IP>)",
		Value: "any",
	}
	NoDiscoverFlag = cli.BoolFlag{
		Name:  "nodiscover",
		Usage: "Disables the peer discovery mechanism (manual peer addition)",
	}
	DiscoveryV5Flag = cli.BoolFlag{
		Name:  "v5disc",
		Usage: "Enables the experimental RLPx V5 (Topic Discovery) mechanism",
	}
	NetrestrictFlag = cli.StringFlag{
		Name:  "netrestrict",
		Usage: "Restricts network communication to the given IP networks (CIDR masks)",
	}

	// ATM the url is left to the user and deployment to
	JSpathFlag = cli.StringFlag{
		Name:  "jspath",
		Usage: "JavaScript root path for `loadScript`",
		Value: ".",
	}

	DagValue3Flag = cli.IntFlag{
		Name:  "dag.dbcache",
		Usage: "Dag dbcache",
		Value: ptn.DefaultConfig.Dag.DbCache,
	}

	LogOutputPathFlag = cli.StringFlag{
		Name:  "log.path",
		Usage: "Log path",
		Value: "", //strings.Join(log.DefaultConfig.OutputPaths, ","),
	}

	LogLevelFlag = cli.StringFlag{
		Name:  "log.lvl",
		Usage: "Log lvl",
		Value: log.DefaultConfig.LoggerLvl,
	}
	LogIsDebugFlag = cli.BoolFlag{
		Name:  "log.debug",
		Usage: "Log debug",
		//Value: ptn.DefaultConfig.Log.IsDebug,
	}
	LogErrPathFlag = cli.StringFlag{
		Name:  "log.errpath",
		Usage: "Log errpath",
		Value: "", //strings.Join(log.DefaultConfig.ErrorOutputPaths, ","),
	}
	LogEncodingFlag = cli.StringFlag{
		Name:  "log.encoding",
		Usage: "Log encoding",
		Value: log.DefaultConfig.Encoding,
	}
	LogOpenModuleFlag = cli.StringFlag{
		Name:  "log.openmodule",
		Usage: "Log openmodule",
		Value: "all", //strings.Join(log.DefaultConfig.OpenModule, ","),
	}
)

// MakeDataDir retrieves the currently requested data directory, terminating
// if none (or the empty string) is specified. If the node is starting a testnet,
// the a subdirectory of the specified datadir will be used.
//func MakeDataDir(ctx *cli.Context) string {
//	if path := ctx.GlobalString(DataDirFlag.Name); path != "" {
//		if ctx.GlobalBool(TestnetFlag.Name) {
//			return filepath.Join(path, "testnet")
//		}
//		return path
//	}
//	Fatalf("Cannot determine default data directory, please set manually (--datadir)")
//	return ""
//}

// setNodeKey creates a node key from set command line flags, either loading it
// from a file or as a specified hex value. If neither flags were provided, this
// method returns nil and an emphemeral key is to be generated.
func setNodeKey(ctx *cli.Context, cfg *p2p.Config) {
	var (
		hex  = ctx.GlobalString(NodeKeyHexFlag.Name)
		file = ctx.GlobalString(NodeKeyFileFlag.Name)
		key  *ecdsa.PrivateKey
		err  error
	)
	switch {
	case file != "" && hex != "":
		Fatalf("Options %q and %q are mutually exclusive", NodeKeyFileFlag.Name, NodeKeyHexFlag.Name)
	case file != "":
		if key, err = crypto.LoadECDSA(file); err != nil {
			Fatalf("Option %q: %v", NodeKeyFileFlag.Name, err)
		}
		cfg.PrivateKey = key
	case hex != "":
		if key, err = crypto.HexToECDSA(hex); err != nil {
			Fatalf("Option %q: %v", NodeKeyHexFlag.Name, err)
		}
		cfg.PrivateKey = key
	}
}

// setNodeUserIdent creates the user identifier from CLI flags.
/*func setNodeUserIdent(ctx *cli.Context, cfg *node.Config) {
	if identity := ctx.GlobalString(IdentityFlag.Name); len(identity) > 0 {
		cfg.UserIdent = identity
	}
}*/

// setBootstrapNodes creates a list of bootstrap nodes from the command line
// flags, reverting to pre-configured ones if none have been specified.
func setBootstrapNodes(ctx *cli.Context, cfg *p2p.Config) {
	urls := configure.MainnetBootnodes
	switch {
	case ctx.GlobalIsSet(BootnodesFlag.Name) || ctx.GlobalIsSet(BootnodesV4Flag.Name):
		if ctx.GlobalIsSet(BootnodesV4Flag.Name) {
			urls = strings.Split(ctx.GlobalString(BootnodesV4Flag.Name), ",")
		} else {
			urls = strings.Split(ctx.GlobalString(BootnodesFlag.Name), ",")
		}
	case ctx.GlobalBool(TestnetFlag.Name):
		urls = configure.TestnetBootnodes
	case len(cfg.BootstrapNodes) > 0:
		return // already set, don't apply defaults.
	}

	cfg.BootstrapNodes = make([]*discover.Node, 0, len(urls))
	for _, url := range urls {
		node, err := discover.ParseNode(url)
		if err != nil {
			log.Error("Bootstrap URL invalid", "pnode", url, "err", err)
			continue
		}
		cfg.BootstrapNodes = append(cfg.BootstrapNodes, node)
	}
}

// setListenAddress creates a TCP listening address string from set command
// line flags.
func setListenAddress(ctx *cli.Context, cfg *p2p.Config) {
	if ctx.GlobalIsSet(ListenPortFlag.Name) {
		cfg.ListenAddr = fmt.Sprintf(":%d", ctx.GlobalInt(ListenPortFlag.Name))
	}
}

// setNAT creates a port mapper from command line flags.
func setNAT(ctx *cli.Context, cfg *p2p.Config) {
	if ctx.GlobalIsSet(NATFlag.Name) {
		natif, err := nat.Parse(ctx.GlobalString(NATFlag.Name))
		if err != nil {
			Fatalf("Option %s: %v", NATFlag.Name, err)
		}
		cfg.NAT = natif
	}
}

// splitAndTrim splits input separated by a comma
// and trims excessive white space from the substrings.
/*func splitAndTrim(input string) []string {
	result := strings.Split(input, ",")
	for i, r := range result {
		result[i] = strings.TrimSpace(r)
	}
	return result
}*/

// setHTTP creates the HTTP RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
/*func setHTTP(ctx *cli.Context, cfg *node.Config) {
	if ctx.GlobalBool(RPCEnabledFlag.Name) && cfg.HTTPHost == "" {
		cfg.HTTPHost = "127.0.0.1"
		if ctx.GlobalIsSet(RPCListenAddrFlag.Name) {
			cfg.HTTPHost = ctx.GlobalString(RPCListenAddrFlag.Name)
		}
	}

	if ctx.GlobalIsSet(RPCPortFlag.Name) {
		cfg.HTTPPort = ctx.GlobalInt(RPCPortFlag.Name)
	}
	if ctx.GlobalIsSet(RPCCORSDomainFlag.Name) {
		cfg.HTTPCors = splitAndTrim(ctx.GlobalString(RPCCORSDomainFlag.Name))
	}
	if ctx.GlobalIsSet(RPCApiFlag.Name) {
		cfg.HTTPModules = splitAndTrim(ctx.GlobalString(RPCApiFlag.Name))
	}
	if ctx.GlobalIsSet(RPCVirtualHostsFlag.Name) {
		cfg.HTTPVirtualHosts = splitAndTrim(ctx.GlobalString(RPCVirtualHostsFlag.Name))
	}
}*/

// setWS creates the WebSocket RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
/*func setWS(ctx *cli.Context, cfg *node.Config) {
	if ctx.GlobalBool(WSEnabledFlag.Name) && cfg.WSHost == "" {
		cfg.WSHost = "127.0.0.1"
		if ctx.GlobalIsSet(WSListenAddrFlag.Name) {
			cfg.WSHost = ctx.GlobalString(WSListenAddrFlag.Name)
		}
	}

	if ctx.GlobalIsSet(WSPortFlag.Name) {
		cfg.WSPort = ctx.GlobalInt(WSPortFlag.Name)
	}
	if ctx.GlobalIsSet(WSAllowedOriginsFlag.Name) {
		cfg.WSOrigins = splitAndTrim(ctx.GlobalString(WSAllowedOriginsFlag.Name))
	}
	if ctx.GlobalIsSet(WSApiFlag.Name) {
		cfg.WSModules = splitAndTrim(ctx.GlobalString(WSApiFlag.Name))
	}
}*/

// setIPC creates an IPC path configuration from the set command line flags,
// returning an empty string if IPC was explicitly disabled, or the set path.
/*func setIPC(ctx *cli.Context, cfg *node.Config) {
	checkExclusive(ctx, IPCDisabledFlag, IPCPathFlag)
	switch {
	case ctx.GlobalBool(IPCDisabledFlag.Name):
		cfg.IPCPath = ""
	case ctx.GlobalIsSet(IPCPathFlag.Name):
		cfg.IPCPath = ctx.GlobalString(IPCPathFlag.Name)
	}
}*/

// makeDatabaseHandles raises out the number of allowed file handles per process
// for Geth and returns half of the allowance to assign to the database.
func makeDatabaseHandles() int {
	limit, err := fdlimit.Current()
	if err != nil {
		Fatalf("Failed to retrieve file descriptor allowance: %v", err)
	}
	if limit < 2048 {
		if err := fdlimit.Raise(2048); err != nil {
			Fatalf("Failed to raise file descriptor allowance: %v", err)
		}
	}
	if limit > 2048 { // cap database file descriptors even if more is available
		limit = 2048
	}
	return limit / 2 // Leave half for networking and other stuff
}

// MakeAddress converts an account specified directly as a hex encoded string or
// a key index in the key store to an internal account representation.
func MakeAddress(ks *keystore.KeyStore, account string) (accounts.Account, error) {
	// If the specified account is a valid address, return it
	addr, err := common.StringToAddress(account)
	if err == nil {
		return accounts.Account{Address: addr}, nil
	} else {
		return accounts.Account{}, fmt.Errorf("invalid account address: %s", account)
	}

}

// setEtherbase retrieves the etherbase either from the directly specified
// command line flags or from the keystore if CLI indexed.
//func setEtherbase(ctx *cli.Context, ks *keystore.KeyStore, cfg *ptn.Config) {
//	if ctx.GlobalIsSet(EtherbaseFlag.Name) {
//		account, err := MakeAddress(ks, ctx.GlobalString(EtherbaseFlag.Name))
//		if err != nil {
//			Fatalf("Option %q: %v", EtherbaseFlag.Name, err)
//		}
//		cfg.Etherbase = account.Address
//	}
//}

// MakePasswordList reads password lines from the file specified by the global --password flag.
func MakePasswordList(ctx *cli.Context) []string {
	path := ctx.GlobalString(PasswordFileFlag.Name)
	if path == "" {
		return nil
	}
	text, err := ioutil.ReadFile(path)
	if err != nil {
		Fatalf("Failed to read password file: %v", err)
	}
	lines := strings.Split(string(text), "\n")
	// Sanitize DOS line endings.
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\n")
	}
	return lines
}

func SetP2PConfig(ctx *cli.Context, cfg *p2p.Config) {
	setNodeKey(ctx, cfg)
	setNAT(ctx, cfg)
	setListenAddress(ctx, cfg)
	setBootstrapNodes(ctx, cfg)
	//setBootstrapNodesV5(ctx, cfg)

	lightClient := ctx.GlobalBool(LightModeFlag.Name) || ctx.GlobalString(SyncModeFlag.Name) == "light"
	lightServer := ctx.GlobalInt(LightServFlag.Name) != 0
	lightPeers := ctx.GlobalInt(LightPeersFlag.Name)

	if ctx.GlobalIsSet(MaxPeersFlag.Name) {
		cfg.MaxPeers = ctx.GlobalInt(MaxPeersFlag.Name)
		if lightServer && !ctx.GlobalIsSet(LightPeersFlag.Name) {
			cfg.MaxPeers += lightPeers
		}
	} else {
		if lightServer {
			cfg.MaxPeers += lightPeers
		}
		if lightClient && ctx.GlobalIsSet(LightPeersFlag.Name) && cfg.MaxPeers < lightPeers {
			cfg.MaxPeers = lightPeers
		}
	}
	if !(lightClient || lightServer) {
		lightPeers = 0
	}
	ptnPeers := cfg.MaxPeers - lightPeers
	if lightClient {
		ptnPeers = 0
	}
	log.Debug("Maximum peer count", "PTN", ptnPeers, "LES", lightPeers, "total", cfg.MaxPeers)

	if ctx.GlobalIsSet(MaxPendingPeersFlag.Name) {
		cfg.MaxPendingPeers = ctx.GlobalInt(MaxPendingPeersFlag.Name)
	}
	if ctx.GlobalIsSet(NoDiscoverFlag.Name) { //|| lightClient {
		cfg.NoDiscovery = true
	}

	if netrestrict := ctx.GlobalString(NetrestrictFlag.Name); netrestrict != "" {
		list, err := netutil.ParseNetlist(netrestrict)
		if err != nil {
			Fatalf("Option %q: %v", NetrestrictFlag.Name, err)
		}
		cfg.NetRestrict = list
	}

	if ctx.GlobalBool(DeveloperFlag.Name) {
		// --dev mode can't use p2p networking.
		cfg.MaxPeers = 0
		cfg.ListenAddr = ":0"
		cfg.NoDiscovery = true
		//cfg.DiscoveryV5 = false
	}
}

// SetNodeConfig applies node-related command line flags to the config.
// 检查命令行中有没有 node 相关的配置，如果有的话覆盖掉cfg中的配置。
func SetNodeConfig(ctx *cli.Context, cfg *node.Config, configDir string) string {
	// setIPC(ctx, cfg)
	// setHTTP(ctx, cfg)
	// setWS(ctx, cfg)
	// setNodeUserIdent(ctx, cfg)

	switch {
	case ctx.GlobalIsSet(DataDirFlag.Name):
		cfg.DataDir, _ = filepath.Abs(ctx.GlobalString(DataDirFlag.Name))
	case ctx.GlobalBool(DeveloperFlag.Name):
		cfg.DataDir = "" // unless explicitly requested, use memory databases
	case ctx.GlobalBool(TestnetFlag.Name):
		cfg.DataDir = filepath.Join(node.DefaultDataDir(), "testnet")
	}

	// 重新计算为绝对路径
	if !filepath.IsAbs(cfg.DataDir) {
		path := filepath.Join(configDir, cfg.DataDir)
		cfg.DataDir = common.GetAbsPath(path)
	}
	dataDir := cfg.DataDir

	if ctx.GlobalIsSet(KeyStoreDirFlag.Name) {
		cfg.KeyStoreDir = ctx.GlobalString(KeyStoreDirFlag.Name)
	}

	if cfg.KeyStoreDir == "" {
		cfg.KeyStoreDir = filepath.Join(dataDir, "keystore")
	}

	// 重新计算为绝对路径
	if !filepath.IsAbs(cfg.KeyStoreDir) {
		path := filepath.Join(configDir, cfg.KeyStoreDir)
		cfg.KeyStoreDir = common.GetAbsPath(path)
	}

	// if ctx.GlobalIsSet(LightKDFFlag.Name) {
	// 	cfg.UseLightweightKDF = ctx.GlobalBool(LightKDFFlag.Name)
	// }
	// if ctx.GlobalIsSet(NoUSBFlag.Name) {
	// 	cfg.NoUSB = ctx.GlobalBool(NoUSBFlag.Name)
	// }

	return dataDir
}

/*
func setGPO(ctx *cli.Context, cfg *gasprice.Config) {
	if ctx.GlobalIsSet(GpoBlocksFlag.Name) {
		cfg.Blocks = ctx.GlobalInt(GpoBlocksFlag.Name)
	}
	if ctx.GlobalIsSet(GpoPercentileFlag.Name) {
		cfg.Percentile = ctx.GlobalInt(GpoPercentileFlag.Name)
	}
}
*/
func SetTxPoolConfig(ctx *cli.Context, cfg *txspool.TxPoolConfig) {
	if ctx.GlobalIsSet(TxPoolNoLocalsFlag.Name) {
		cfg.NoLocals = ctx.GlobalBool(TxPoolNoLocalsFlag.Name)
	}
	if ctx.GlobalIsSet(TxPoolJournalFlag.Name) {
		cfg.Journal = ctx.GlobalString(TxPoolJournalFlag.Name)
	}
	if ctx.GlobalIsSet(TxPoolRejournalFlag.Name) {
		cfg.Rejournal = ctx.GlobalDuration(TxPoolRejournalFlag.Name)
	}
	//	if ctx.GlobalIsSet(TxPoolPriceLimitFlag.Name) {
	//		cfg.PriceLimit = ctx.GlobalUint64(TxPoolPriceLimitFlag.Name)
	//	}
	if ctx.GlobalIsSet(TxPoolPriceBumpFlag.Name) {
		cfg.PriceBump = ctx.GlobalUint64(TxPoolPriceBumpFlag.Name)
	}
	if ctx.GlobalIsSet(TxPoolGlobalSlotsFlag.Name) {
		cfg.GlobalSlots = ctx.GlobalUint64(TxPoolGlobalSlotsFlag.Name)
	}
	//if ctx.GlobalIsSet(TxPoolAccountQueueFlag.Name) {
	//	cfg.AccountQueue = ctx.GlobalUint64(TxPoolAccountQueueFlag.Name)
	//}
	if ctx.GlobalIsSet(TxPoolGlobalQueueFlag.Name) {
		cfg.GlobalQueue = ctx.GlobalUint64(TxPoolGlobalQueueFlag.Name)
	}
	if ctx.GlobalIsSet(TxPoolLifetimeFlag.Name) {
		cfg.Lifetime = ctx.GlobalDuration(TxPoolLifetimeFlag.Name)
	}
	if ctx.GlobalIsSet(TxPoolRemovetimeFlag.Name) {
		cfg.Removetime = ctx.GlobalDuration(TxPoolRemovetimeFlag.Name)
	}
}

// checkExclusive verifies that only a single isntance of the provided flags was
// set by the user. Each flag might optionally be followed by a string type to
// specialize it further.
func checkExclusive(ctx *cli.Context, args ...interface{}) {
	set := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		// Make sure the next argument is a flag and skip if not set
		flag, ok := args[i].(cli.Flag)
		if !ok {
			panic(fmt.Sprintf("invalid argument, not cli.Flag type: %T", args[i]))
		}
		// Check if next arg extends current and expand its name if so
		name := flag.GetName()

		if i+1 < len(args) {
			switch option := args[i+1].(type) {
			case string:
				// Extended flag, expand the name and shift the arguments
				if ctx.GlobalString(flag.GetName()) == option {
					name += "=" + option
				}
				i++

			case cli.Flag:
			default:
				panic(fmt.Sprintf("invalid argument, not cli.Flag or string extension: %T", args[i+1]))
			}
		}
		// Mark the flag if it's set
		if ctx.GlobalIsSet(flag.GetName()) {
			set = append(set, "--"+name)
		}
	}
	if len(set) > 1 {
		Fatalf("Flags %v can't be used at the same time", strings.Join(set, ", "))
	}
}

// func setGPO(ctx *cli.Context, cfg *gasprice.Config) {
// 	if ctx.GlobalIsSet(GpoBlocksFlag.Name) {
// 		cfg.Blocks = ctx.GlobalInt(GpoBlocksFlag.Name)
// 	}
// 	if ctx.GlobalIsSet(GpoPercentileFlag.Name) {
// 		cfg.Percentile = ctx.GlobalInt(GpoPercentileFlag.Name)
// 	}
// }

// SetDagConfig applies dag related command line flags to the config.
func SetDagConfig(ctx *cli.Context, cfg *dagconfig.Config, dataDir string) {
	//	if ctx.GlobalIsSet(DagValue1Flag.Name) {
	//		cfg.DbPath = ctx.GlobalString(DagValue1Flag.Name)
	//	}
	//	if ctx.GlobalIsSet(DagValue2Flag.Name) {
	//		cfg.DbName = ctx.GlobalString(DagValue2Flag.Name)
	//	}
	if ctx.GlobalIsSet(DagValue3Flag.Name) {
		cfg.DbCache = ctx.GlobalInt(DagValue3Flag.Name)
	}
	// 重新计算为绝对路径
	if !filepath.IsAbs(cfg.DbPath) {
		path := filepath.Join(dataDir, cfg.DbPath)
		cfg.DbPath = common.GetAbsPath(path)
	}

	dagconfig.DagConfig = *cfg
}

//func SetContractConfig(ctx *cli.Context, cfg *contractcfg.Config, dataDir string) {
//	// 重新计算为绝对路径
//	if !filepath.IsAbs(cfg.ContractFileSystemPath) {
//		path := filepath.Join(dataDir, cfg.ContractFileSystemPath)
//		cfg.ContractFileSystemPath = common.GetAbsPath(path)
//	}
//}

func SetLogConfig(ctx *cli.Context, cfg *log.Config, configDir string, isInConsole bool) {
	// 1. 重新计算log.output的路径
	if temp := ctx.GlobalString(LogOutputPathFlag.Name); temp != "" {
		outputPaths := strings.Split(temp, ",")

		newOutputPaths := make([]string, 0)
		for _, outputPath := range outputPaths {
			if outputPath == "" {
				continue
			}

			if outputPath != log.LogStdout {
				if !filepath.IsAbs(outputPath) {
					outputPath = filepath.Join(common.GetWorkPath(), outputPath)
				}
				if files.IsDir(outputPath) {
					outputPath = filepath.Join(outputPath, filepath.Base(log.DefaultConfig.OutputPaths[1]))
				}
			}

			newOutputPaths = append(newOutputPaths, outputPath)
		}

		cfg.OutputPaths = newOutputPaths
	} else {
		for i, outputPath := range cfg.OutputPaths {
			if outputPath == log.LogStdout {
				continue
			}

			if !filepath.IsAbs(outputPath) {
				outputPath = filepath.Join(configDir, outputPath)
			}

			cfg.OutputPaths[i] = common.GetAbsPath(outputPath)
		}
	}

	// 2. 处理其他 log 配置
	if ctx.GlobalIsSet(LogLevelFlag.Name) {
		cfg.LoggerLvl = ctx.GlobalString(LogLevelFlag.Name)
	}
	if ctx.GlobalIsSet(LogIsDebugFlag.Name) {
		cfg.Development = ctx.GlobalBool(LogIsDebugFlag.Name)
	}
	if ctx.GlobalIsSet(LogEncodingFlag.Name) {
		cfg.Encoding = ctx.GlobalString(LogEncodingFlag.Name)
	}
	//if temp := ctx.GlobalString(LogOpenModuleFlag.Name); temp != "" {
	//	cfg.OpenModule = strings.Split(temp, ",")
	//}

	// 3. 重新计算log.ErrPath的路径
	if temp := ctx.GlobalString(LogErrPathFlag.Name); temp != "" {
		errPaths := strings.Split(temp, ",")

		newErrPaths := make([]string, 0)
		for _, errPath := range errPaths {
			if errPath == "" {
				continue
			}

			if errPath != log.LogStderr {
				if !filepath.IsAbs(errPath) {
					errPath = filepath.Join(common.GetWorkPath(), errPath)
				}
				if files.IsDir(errPath) {
					errPath = filepath.Join(errPath, filepath.Base(log.DefaultConfig.ErrorOutputPaths[1]))
				}
			}

			newErrPaths = append(newErrPaths, errPath)
		}

		cfg.ErrorOutputPaths = newErrPaths
	} else {
		for i, errPath := range cfg.ErrorOutputPaths {
			if errPath == log.LogStderr {
				continue
			}

			if !filepath.IsAbs(errPath) {
				errPath = filepath.Join(configDir, errPath)
			}

			cfg.ErrorOutputPaths[i] = common.GetAbsPath(errPath)
		}
	}

	// 3. 应用 log 配置
	log.LogConfig = *cfg

	// 4. 处理console的特殊情况
	if isInConsole {
		log.ConsoleInitLogger()
	}
}

// SetPtnConfig applies ptn-related command line flags to the config.
func SetPtnConfig(ctx *cli.Context, stack *node.Node, cfg *ptn.Config) {
	// Avoid conflicting network flags
	checkExclusive(ctx, TestnetFlag)
	checkExclusive(ctx, FastSyncFlag, LightModeFlag, SyncModeFlag)
	checkExclusive(ctx, LightServFlag, LightModeFlag)
	checkExclusive(ctx, LightServFlag, SyncModeFlag, "light")

	ks := stack.GetKeyStore()

	switch {
	case ctx.GlobalIsSet(SyncModeFlag.Name):
		cfg.SyncMode = *GlobalTextMarshaler(ctx, SyncModeFlag.Name).(*downloader.SyncMode)
	case ctx.GlobalBool(FastSyncFlag.Name):
		cfg.SyncMode = downloader.FastSync
	case ctx.GlobalBool(LightModeFlag.Name):
		cfg.SyncMode = downloader.LightSync
	}
	if ctx.GlobalIsSet(LightServFlag.Name) {
		cfg.LightServ = ctx.GlobalInt(LightServFlag.Name)
	}
	if ctx.GlobalIsSet(LightPeersFlag.Name) {
		cfg.LightPeers = ctx.GlobalInt(LightPeersFlag.Name)
	}
	if ctx.GlobalIsSet(NetworkIdFlag.Name) {
		cfg.NetworkId = ctx.GlobalUint64(NetworkIdFlag.Name)
	}

	if ctx.GlobalIsSet(CacheFlag.Name) || ctx.GlobalIsSet(CacheDatabaseFlag.Name) {
		cfg.DatabaseCache = ctx.GlobalInt(CacheFlag.Name) * ctx.GlobalInt(CacheDatabaseFlag.Name) / 100
	}
	cfg.DatabaseHandles = makeDatabaseHandles()

	if gcmode := ctx.GlobalString(GCModeFlag.Name); gcmode != "full" && gcmode != "archive" {
		Fatalf("--%s must be either 'full' or 'archive'", GCModeFlag.Name)
	}
	cfg.NoPruning = ctx.GlobalString(GCModeFlag.Name) == "archive"

	if ctx.GlobalIsSet(CacheFlag.Name) || ctx.GlobalIsSet(CacheGCFlag.Name) {
		cfg.TrieCache = ctx.GlobalInt(CacheFlag.Name) * ctx.GlobalInt(CacheGCFlag.Name) / 100
	}
	if ctx.GlobalIsSet(MinerThreadsFlag.Name) {
		cfg.MinerThreads = ctx.GlobalInt(MinerThreadsFlag.Name)
	}
	if ctx.GlobalIsSet(DocRootFlag.Name) {
		cfg.DocRoot = ctx.GlobalString(DocRootFlag.Name)
	}
	if ctx.GlobalIsSet(ExtraDataFlag.Name) {
		cfg.ExtraData = []byte(ctx.GlobalString(ExtraDataFlag.Name))
	}
	if ctx.GlobalIsSet(CryptoLibFlag.Name) {
		t := ctx.GlobalString(CryptoLibFlag.Name)
		cfg.CryptoLib, _ = hex.DecodeString(t)
	}
	if ctx.GlobalIsSet(VMEnableDebugFlag.Name) {
		// TODO(fjl): force-enable this in --dev mode
		cfg.EnablePreimageRecording = ctx.GlobalBool(VMEnableDebugFlag.Name)
	}

	// Override any default configs for hard coded networks.
	switch {
	case ctx.GlobalBool(TestnetFlag.Name):
		if !ctx.GlobalIsSet(NetworkIdFlag.Name) {
			cfg.NetworkId = 1
		}
		cfg.Genesis = gen.DefaultTestnetGenesisBlock()
	case ctx.GlobalBool(DeveloperFlag.Name):
		// Create new developer account or reuse existing one
		var (
			developer accounts.Account
			err       error
		)
		if accs := ks.Accounts(); len(accs) > 0 {
			developer = ks.Accounts()[0]
		} else {
			developer, err = ks.NewAccount("")
			if err != nil {
				Fatalf("Failed to create developer account: %v", err)
			}
		}
		if err := ks.Unlock(developer, ""); err != nil {
			Fatalf("Failed to unlock developer account: %v", err)
		}
		log.Info("Using developer account", "address", developer.Address)

		//if !ctx.GlobalIsSet(GasPriceFlag.Name) {
		//	cfg.GasPrice = big.NewInt(1)
		//}
	}
	// TODO(fjl): move trie cache generations into config
	if gen := ctx.GlobalInt(TrieCacheGenFlag.Name); gen > 0 {
		state.MaxTrieCacheGen = uint16(gen)
	}
	//cfg.TokenSubProtocol = strings.ToLower(cfg.Dag.MainToken)
}

// SetDashboardConfig applies dashboard related command line flags to the config.
func SetPrometheusConfig(ctx *cli.Context, cfg *prometheus.Config) {
	//cfg.Host = ctx.GlobalString(DashboardAddrFlag.Name)
	//cfg.Port = ctx.GlobalInt(DashboardPortFlag.Name)
	//cfg.Refresh = ctx.GlobalDuration(DashboardRefreshFlag.Name)
}

// RegisterPtnService adds an PalletOne client to the stack.
func RegisterPtnService(stack *node.Node, cfg *ptn.Config) {
	var err error
	if cfg.SyncMode == downloader.LightSync {
		err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
			return light.New(ctx, cfg, configure.LPSProtocol, stack.CacheDb, stack.IsTestNet)
		})
	} else {
		err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
			fullNode, err := ptn.New(ctx, cfg, stack.CacheDb, stack.IsTestNet)
			if fullNode != nil && cfg.LightServ > 0 {
				ls, _ := light.NewLesServer(fullNode, cfg, configure.LPSProtocol)
				fullNode.AddLesServer(ls)

				cs, _ := cors.NewCoresServer(fullNode, cfg)
				fullNode.AddCorsServer(cs)
			}
			return fullNode, err
		})
	}

	if err != nil {
		Fatalf("Failed to register the PalletOne service: %v", err)
	}
}

// RegisterDashboardService adds a dashboard to the stack.
func RegisterDashboardService(stack *node.Node, cfg *dashboard.Config, commit string) {
	stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return dashboard.New(cfg, commit)
	})
}

// RegisterPtnStatsService configures the PalletOne Stats daemon and adds it to
// th egiven node.
//func RegisterPtnStatsService(stack *node.Node, url string) {
//	err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
//		// Retrieve ptn service
//		var ptnServ *ptn.PalletOne
//		ctx.Service(&ptnServ)
//
//		return ptnstats.New(url, ptnServ)
//	})
//
//	if err != nil {
//		Fatalf("Failed to register the PalletOne Stats service: %v", err)
//	}
//}

func RegisterPrometheusService(stack *node.Node, config prometheus.Config) {
	stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return prometheus.New(config)
	})
}

// SetupNetwork configures the system for either the main net or some test network.
func SetupNetwork(ctx *cli.Context) {
	// TODO(fjl): move target gas limit into config
	// 配置gas limit值
	//configure.TargetGasLimit = ctx.GlobalUint64(TargetGasLimitFlag.Name)
}

// MakeChainDatabase open an LevelDB using the flags passed to the client and will hard crash if it fails.
//func MakeChainDatabase(ctx *cli.Context, stack *node.Node) ptndb.Database {
//	var (
//		cache   = ctx.GlobalInt(CacheFlag.Name) * ctx.GlobalInt(CacheDatabaseFlag.Name) / 100
//		handles = makeDatabaseHandles()
//	)
//	name := "chaindata"
//	if ctx.GlobalBool(LightModeFlag.Name) {
//		name = "lightchaindata"
//	}
//	chainDb, err := stack.OpenDatabase(name, cache, handles)
//	if err != nil {
//		Fatalf("Could not open database: %v", err)
//	}
//	return chainDb
//}

func MakeGenesis(ctx *cli.Context) *core.Genesis {
	var genesis *core.Genesis
	switch {
	case ctx.GlobalBool(TestnetFlag.Name):
		genesis = gen.DefaultTestnetGenesisBlock()
	case ctx.GlobalBool(DeveloperFlag.Name):
		Fatalf("Developer chains are ephemeral")
	}
	return genesis
}

// MakeConsolePreloads retrieves the absolute paths for the console JavaScript
// scripts to preload before starting.
func MakeConsolePreloads(ctx *cli.Context) []string {
	// Skip preloading if there's nothing to preload
	if ctx.GlobalString(PreloadJSFlag.Name) == "" {
		return nil
	}
	// Otherwise resolve absolute paths and return them
	preloads := []string{}

	assets := ctx.GlobalString(JSpathFlag.Name)
	for _, file := range strings.Split(ctx.GlobalString(PreloadJSFlag.Name), ",") {
		preloads = append(preloads, common.AbsolutePath(assets, strings.TrimSpace(file)))
	}
	return preloads
}

// MigrateFlags sets the global flag from a local flag when it's set.
// This is a temporary function used for migrating old command/flags to the
// new format.
//
// e.g. gptn account new --keystore /tmp/mykeystore --lightkdf
//
// is equivalent after calling this method with:
//
// gptn --keystore /tmp/mykeystore --lightkdf account new
//
// This allows the use of the existing configuration functionality.
// When all flags are migrated this function can be removed and the existing
// configuration functionality must be changed that is uses local flags
func MigrateFlags(action func(ctx *cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.GlobalSet(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}
