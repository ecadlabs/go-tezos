
# Work in progress

go-tezos is a Go client library for Tezos RPC’s

--

This client RPC library is in development and should be considered alpha.
Contributors are welcome. We will start to tag releases of this library in
November 2018.

The library will be useful to anyone wanting to build tools, products or
services on top of the Tezos RPC API. 

The library will be:

* Well tested
* Nightly Integration tests against official Tezos docker images
* Written in Idiomatic Go
* Aim to have complete coverage of the Tezos API and stay up to date with new RPCs or changes to existing RPCs

# Tezos RPC Api documentation

The best known RPC API docs are available here: http://tezos.gitlab.io/mainnet/ 

# Users of `go-tezos`

* A prometheus metrics exporter for a Tezos node https://github.com/ecadlabs/tezos_exporter

## Development

### Running a tezos RPC node using docker-compose

To run a local tezos RPC node using docker, run the following command:

`docker-compose up`

The node will generate an identity and then, it will the chain from other nodes
on the network. The process of synchronizing or downloading the chain can take
some time, but most of the RPC will work while this process completes.

The `alphanet` image tag means you are not interacting with the live `mainnet`.
You can connect to `mainnet` with the `tezos/tezos:mainnet` image, but it takes
longer to sync.

The `docker-compose.yml` file uses volumes, so when you restart the node, it
won't have to regenerate an identity, or sync the entire chain.

### Running a tezos RPC node using docker

If you want to run a tezos node quickly, without using `docker-compose` try:

`docker run -it --rm --name tezos_node -p 8732:8732 tezos/tezos:alphanet tezos-node`

### Interacting with tezos RPC

With the tezos-node docker image, you can test that the RPC interface is
working:

`curl localhost:8732/network/stat`

The tezos-client cli is available in the docker image, and can be run as
follows:

`docker exec -it tezos_node tezos-client -A 0.0.0.0 man`

`docker exec -it tezos_node tezos-client -A 0.0.0.0 rpc list`

Create a shell alias that you can run from your docker host for convenience;

`alias tezos-client='sudo docker exec -it -e TEZOS_CLIENT_UNSAFE_DISABLE_DISCLAIMER=Y tezos_node tezos-client -A 0.0.0.0'`



