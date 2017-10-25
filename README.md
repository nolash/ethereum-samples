# p2p programming in go-ethereum

## USING THESE EXAMPLES

Pss is part of Swarm POC 3, which is currently being merged into ethereum master. Before this merge is completed, the code examples cannot be run with the current official go-ethereum implementation. They will also not work with any of the active swarm development branches, because it relies on total functionality provided only in part by each of them.

This repository keeps a temporary merge branch providing all necessary functionality along with a pss API that will work as-is after the POC 3 merge. Please use this branch for these examples.

https://github.com/nolash/go-ethereum/tree/pss-tmp-binfix

If you only plan to use the `pss` library (you do not need a working `swarm` executable), please use this branch:

https://github.com/ethersphere/go-ethereum/tree/pss

## DESCRIPTION

These code examples are intended to demonstrate the main building blocks of peer-to-peer comunications in go-ethereum. They are organized in chapters, where every example in a chapter (ideally) builds on the next.

They form the basis for a future tutorial with a detailed simple-to-follow narrative.

## TODO

* Write general introduction to components in go-ethereum devp2p
* Write accompanying tutorials
* Add example of lowlevel PSS implementation

## Contents

### A - Lowlevel devp2p

p2p server is the core structure for communications between nodes. It creates and maintains ip connections between nodes, and embeds p2p encryption with RLP serialization (RLPx) on the data connection.

* A1_Server.go 

  Initializing and starting the p2p Server

* A2_Connect.go

  Connecting two p2p Servers

* A3_Events.go

  Receiving notification of completed connection

* A4_Message.go

  Sending a message from server A to server B

* A5_Reply.go

  A sample p2p ping protocol implementation

### B - Remote Procedure Calls

* B1_RPC.go

  Create IPC RPC server with one method and call it  

* B2_Method.go

  Retrieve information from p2p server through RPC

* B3_Message.go

  Receive notification of p2p messaging events through RPC

### C - Service node

This entity encapsulates the p2p server and the RPC call APIs, in packages called "services." A Node.Service is defined as an interface, and objects of this interface are registered with the node, and automatically started in alphabetical order when the service node is started.

* C1_Stack.go

  Create and start a service node

* C2_Nodeinfo.go

  Local accessors to RPC APIs

* C3_Service.go

  Defining and running a service

* C4_Full.go

  Servicenode ping protocol implementation controlled through RPC

### D - Complex nodes

`devp2p` provides a framework for designing autonomous protocol handling code. This chapter shows how to implement one, and how to combine several services providing their own APIs and protocols in the same service node.

* D1_Protocols.go

  p2p protocol abstraction layer enabling automatic recognition of messages and external message handlers

* D2_Multiservice.go

  Registering multiple services with the service node

### E - Pss

Pss enables encrypted messaging between nodes that aren't directly connected through p2p server, by relaying the message through nodes between them. Relaying is done with swarm's kademlia routing. The message is encrypted end-to-end using ephemeral public key cryptography. 

* E1_Pss.go

  Set up a pss-activated swarm node, and send a message using public key encryption. 

* E2_PssRouting.go

  Demonstrates how to perform dark routing in pss.

* E3_PssSym.go

  Create an arbitrary symmetric encryption key, and send message with it.

* E4_PssHandshake.go

  Using the builtin convenience method for Diffie-Hellmann key exchange.

* E5_PssProtocol.go

  Implementing devp2p style protocols over pss.

* E6_PssClient.go

  Mounting devp2p style protocols on an RPC connection.
