# sample pss protocol

This example illustrates how to implement a protocol of some complexity using `pss`.

The `sim.go` driver implements the protocol on a normal `devp2p` connection using the simulations framework. 

The `sim_pss.go` driver implements the same protocol on over `pss` using the simulations framework.

The `main.go` and `main_pss.go` files are respective standalone binaries .

Files in `service/` and `protocol/` implement the protocol itself, and are shared between both drivers. The pss and swarm specific code is isolated to `bzz/`. This way, the extra implmentation needed for `pss` is hopefully clear.

Its currently based on the branch on https://github.com/nolash/go-ethereum/tree/sos18-demo-2
