## Go ZWS Chain

Official golang implementation of the zwschain protocol.

## Building the source

Building gptn requires both a Go (version 1.12 or later) and a C compiler.
You can install them using your favourite package manager.
set GO111MODULE:

```bash
export GO111MODULE=on
```

Once the dependencies are installed, run

```bash
make gptn
```

or, to build the full suite of utilities:

```bash
make all
```

but, to build the full suite of utilities in window,you should:

```bash
go get ./...
go get -u ./...
go build
```

## Executables

The go-zwschain project comes with several wrappers/executables found in the `cmd` directory.

| Command    | Description |
|:----------:|-------------|
| **`gptn`** | Our main zwschain CLI client. It is the entry point into the zwschain network (main-, test- or private net), capable of running as a full node (default) archive node (retaining all historical state) or a light node (retrieving data live). It can be used by other processes as a gateway into the zwschain network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `gptn --help` |

## Running gptn

Going through all the possible command line flags is out of scope here, but we've
enumerated a few common parameter combos to get you up to speed quickly on how you can run your
own Gptn instance.

### Full node on the main zwschain network

By far the most common scenario is people wanting to simply interact with the zwschain network:
create accounts; transfer funds; deploy and interact with contracts. For this particular use-case
the user doesn't care about years-old historical data, so we can fast-sync quickly to the current
state of the network. To do so:

```bash
$ mkdir your_dir
$ gptn --datadir="your_dir" newgenesis path/to/your-genesis.json
$ gptn --datadir="your_dir" init path/to/your-genesis.json
$ gptn --datadir="your_dir" --configfile /path/to/your_config.toml
```

This command will:

 * Start gptn in fast sync mode (default, can be changed with the `--syncmode` flag), causing it to
   download more data in exchange for avoiding processing the entire history of the zwschain network,
   which is very CPU intensive.

### Configuration

As an alternative to passing the numerous flags to the `gptn` binary, you can also pass a configuration file via:

```bash
$ gptn --configfile /path/to/your_config.toml
```

To get a template configuration file you can use the `dumpconfig` subcommand to export current default configurations:

```bash
$ gptn dumpconfig /path/to/your_config.toml
```

Open the ` /path/to/your_config.toml ` file in your favorite text editor, and set the field values what you want to change, uncommenting them if necessary.

### Operating a private network

Maintaining your own private network is more involved as a lot of configurations taken for granted in
the official networks need to be manually set up.

#### Defining the private genesis state

First, you'll need to create the genesis state of your networks, which all nodes need to be aware of and agree upon. This consists of a JSON file (e.g. call it `genesis.json`):

You can create a JSON file for the genesis state of a new chain with an existing account or a newly created account named `your-genesis.json` by running this command:

```bash
$ gptn newgenesis path/to/your-genesis.json
```

#### Defining the private mediator parameters

First, you'll need to create the mediator parameters of your networks, which all nodes need to be aware of and agree upon. This consists of a TOML file (e.g. call it `zws.toml`):

```bash
[MediatorPlugin]
EnableProducing = true
EnableStaleProduction = true
EnableConsecutiveProduction = false

[[MediatorPlugin.Mediators]]
Address = ""
Password = ""
InitPrivKey = ""
InitPubKey = ""
```

Get InitPrivKey and InitPubKey with the following command

```bash
$ gptn mediator initdks
```

InitPrivKey = private key, InitPubKey = public key

When running command `gptn --datadir="your_dir" newgenesis` will create Address and input your password.

##### Customization of the genesis file

If you want to customize the network’s genesis state, edit the newly created your-genesis.json file. This allows you to control things such as:

* The initial values of chain parameters
* Assets and their initial distribution

With the genesis state defined in the above JSON file, you'll need to initialize **every** Gptn node with it prior to starting it up to ensure all blockchain parameters are correctly set:

```bash
$ gptn init path/to/your-genesis.json
```

## Contribution

Thank you for considering to help out with the source code! We welcome contributions from
anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to go-zwschain, please fork, fix, commit and send a pull request
for the maintainers to review and merge into the main code base. If you wish to submit more
complex changes though, please check up with the core devs first
to ensure those changes are in line with the general philosophy of the project and/or get some
early feedback which can make both your efforts much lighter as well as our review and merge
procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting) guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
 * Pull requests need to be based on and opened against the `master` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "ptn, rpc: make trace configs optional"

## License

The go-zwschain binaries (i.e. all code inside of the `cmd` directory) is licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also included
in our repository in the `COPYING` file.
