<div style="text-align:center">
  <img
  src="https://raw.githubusercontent.com/Dharitri-org/sme-dharitri/master/dharitri_logo_01.svg"
  alt="DHaritri">
</div>
<br>

[![](https://img.shields.io/badge/made%20by-Dharitri%20Network-blue.svg)](http://dharitri.org/)
[![](https://img.shields.io/badge/project-Dharitri%20Network%20Testnet-blue.svg)](https://testnet.dharitri.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dharitri-org/sme-dharitri)](https://goreportcard.com/report/github.com/Dharitri-org/sme-dharitri)
[![LoC](https://tokei.rs/b1/Dharitri-org/sme-dharitri?category=code)](https://github.com/Dharitri-org/sme-dharitri)
[![API Reference](https://godoc.org/github.com/Dharitri-org/sme-dharitri?status.svg)](https://godoc.org/github.com/Dharitri-org/sme-dharitri)
[![riot.im](https://img.shields.io/badge/riot.im-JOIN%20CHAT-green.svg)](https://riot.im/app/#/room/#dharitri:matrix.org)

# sme-dharitri

The go implementation for the Dharitri Network testnet

## Installation and running

In order to join the network as an observer or as a validator, the required steps to **build from source and setup explicitly** are explained below. 
Alternatively, in order to use the Docker Image, jump to [Using the Docker Image](#using-the-docker-image).

### Step 1: install & configure go:
The installation of go should proceed as shown in official golang installation guide https://golang.org/doc/install . In order to run the node, minimum golang version should be 1.12.4.

### Step 2: clone the repository and build the binaries:
The main branch that will be used is the master branch. Alternatively, an older release tag can be used.

```
# set $GOPATH if not set and export to ~/.profile along with Go binary path
$ if [[ $GOPATH=="" ]]; then GOPATH="$HOME/go" fi
$ mkdir -p $GOPATH/src/github.com/Dharitri-org
$ cd $GOPATH/src/github.com/Dharitri-org
$ git clone https://github.com/Dharitri-org/sme-dharitri
$ cd sme-dharitri && git checkout master
$ GO111MODULE=on go mod vendor
$ cd cmd/node && go build
```
The Node depends on the Core Virtual Machine, which is a separate binary. Depending on the preferred setup, there are two slightly different options to build Core.

<b>Option A</b>: for development, which also implies running tests:

First, create a persistent environment variable named `$CORE_PATH`. For example, place it in `~/.profile`, then restart the user session:
```
export CORE_PATH="$HOME/Core/core"
``` 

Note that the path includes the name of the binary, `core`.
Secondly, run `make core`
```
$ cd $GOPATH/src/github.com/Dharitri-org/sme-dharitri
$ make core
````
The binary will be generated at `$CORE_PATH`. Whether you run the Node itself or the tests, this path wil be used to start Core.

<b>Option B</b>: no development needed, just running the node
```
$ cd $GOPATH/src/github.com/Dharitri-org/sme-dharitri
$ CORE_PATH=./cmd/node make core
```
The Core binary will be built and placed near the node

### Step 3: creating the node’s identity:
In order to be registered in the Dharitri Network, a node must possess 2 types of (secret key, public key) pairs. One is used to identify the node’s credential used to generate transactions (having the sender field its account address) and the other is used in the process of the block signing. Please note that this is a preliminary mechanism, in the next releases the first (private, public key) pair will be dropped when the staking mechanism will be fully implemented. To build and run the keygenerator, the following commands will need to be run:

```
$ cd $GOPATH/src/github.com/Dharitri-org/sme-dharitri/cmd/keygenerator
$ go build
$ ./keygenerator
```

### Start the node 
#### Step 4a: Join Dharitri testnet:
Follow the steps outlined [here](https://docs.dharitri.org/validators/install). This is because in order to join the testnet you need a specific node configuration.
______
OR
______
#### Step 4b: copying credentials and starting a node in a separate network:
The previous generated .pem file needs to be copied in the same directory where the node binary resides in order to start the node.

```
$  cp validatorKey.pem ./../node/config/
$  cd ../node && ./node
```

The node binary has some flags defined (for a brief description, the user can use --help flag). Those flags can be used to directly alter the configuration values defined in .toml/.json files and can be used when launching more than one instance of the binary. 

### Running the tests	
```	
$ go test ./...	
```

## Compiling new fields in .proto files (should be updated when required PR will be merged in gogo protobuf master branch):
1. Download protoc compiler: https://github.com/protocolbuffers/protobuf/releases 
 (if you are running under linux on a x64 you might want to download protoc-3.11.4-linux-x86_64.zip)
2. Expand archive, copy the /include/google folder in /usr/include using <br>
`sudo cp -r google /usr/include`
3. Copy bin/protoc using <br>
`sudo cp protoc  /usr/bin` 
4. Fetch the repo github.com/Dharitri-org/protobuf
5. Compile gogo slick & copy binary using
```
cd protoc-gen-gogoslick
go build
sudo cp protoc-gen-gogoslick /usr/bin/
```

Done

## Using the Docker Image

The following command runs a Node using the **latest** [Docker image](https://hub.docker.com/r/Dharitri-org/sme-dharitri-node) and maps a container folder to a local one that holds the necessary configuration:

```
docker run -d -v /absolute/path/to/config/:/data/ Dharitri-org/sme-dharitri-node:latest \
 --nodes-setup-file="/data/nodesSetup.json" \
 --p2p-config="/data/config/p2p.toml" \
 --validator-key-pem-file="/data/keys/validatorKey.pem"
 ```

 In the snippet above, make sure you adjust the path to a valid configuration folder and also provide the appropriate command line arguments to the Node. For more details go to [Node CLI](https://docs.dharitri.org/validators/node-cli).

 In order to run a container using the latest **development** version instead, use the image **`/Dharitri-org/sme-dharitri-node:devlatest`**. Furthermore, in order to use a specific release or tagged branch, use the image **`Dharitri-org/sme-dharitri-node:<tag>`**.


## Progress

### Done
- [x] Cryptography
  - [x] Schnorr Signature
  - [x] Belare-Neven Signature
  - [x] BLS Signature
  - [x] Modified BLS Multi-signature
- [x] Datastructures
  - [x] Transaction
  - [x] Block
  - [x] Account
  - [x] Trie
- [x] Execution
  - [x] Transaction
  - [x] Block
  - [x] State update
  - [x] Synchronization
  - [x] Shard Fork choice
- [x] Peer2Peer - libp2p
- [x] Consensus - SPoS
- [x] Sharding - fixed number
  - [x] Transaction dispatcher 
  - [x] Transaction
  - [x] State
  - [x] Network - Message dispatching
- [x] MetaChain
  - [x] Data Structures
  - [x] Block Processor
  - [x] Interceptors/Resolvers
  - [x] Consensus
- [x] Block K finality scheme
- [x] VM - K-Framework
  - [x] K Framework go backend
  - [x] IELE Core
  - [x] IELE Core tests
  - [x] IELE Adapter
- [x] Smart Contracts on a Sharded Architecture
  - [x] Concept reviewed
  - [x] VM integration
  - [x] SC Deployment
- [x] Governance
  - [x] Concept reviewed
- [x] Economics
  - [x] Concept reviewed  
- [x] Optimizations
  - [x] Randomness
  - [x] Consensus
- [x] Bootstrap from storage
- [x] Testing 
  - [x] Unit tests
  - [x] Integration tests
  - [x] TeamCity continuous integration
  - [x] Manual testing
- [x] Epochs
  - [x] Nodes dispatcher (shuffling)
- [x] Network sharding
  - [x] Optimized wiring protocol
- [x] VM
  - [x] EVM Core
  - [x] EVM Core tests
  - [x] EVM Adapter
- [x] Fee structure
- [x] Smart Contracts on a Sharded Architecture
  - [x] Async callbacks
- [x] Testing
  - [x] Automate tests with AWS
  - [x] Nodes Monitoring

### In progress

- [ ] Smart Contracts on a Sharded Architecture
  - [ ] Dependency checker + SC migration
  - [ ] Storage rent + SC backup & restore
- [ ] Adaptive State Sharding
  - [ ] Splitting
  - [ ] Merging 
  - [ ] Redundancy
- [ ] Privacy
- [ ] DEX integration
- [ ] Interoperability
- [ ] Optimizations
  - [ ] Smart Contract 
- [ ] Governance
  - [ ] SC for MOA IP
  - [ ] Enforced Upgrade mechanism for voted MOA IP
- [ ] Bugfixing


## Contribution
Thank you for considering to help out with the source code! We welcome contributions from anyone on the internet, and are grateful for even the smallest of fixes to Dharitri!

If you'd like to contribute to Dharitri, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core developers first on our [riot channel](https://riot.im/app/#/room/#dharitri:matrix.org) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

 - Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting) guidelines.
 - Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
 - Pull requests need to be based on and opened against the master branch.
 - Commit messages should be prefixed with the package(s) they modify.
    - E.g. "core/indexer: fixed a typo"

Please see the [documentation](https://docs.dharitri.org/) for more details on the Dharitri project.

