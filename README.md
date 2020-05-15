## Peer to peer chat

That code simulates a peer to peer (P2P) broadcast chat that recognizes new network nodes and propagates their connection to already known nodes.

### How do I get set up? ###

To run this project locally, first of all you need to clone this repository, the project had the following requirements: 

#### Requirements ####

* docker
* docker-compose

To run the project you just need to execute `execute.sh` file passing required arguments. The required arguments are listed below:

* n: specifies the node name, should be unique for each node;
* p: specifies in which port current node will listen;
* j: specifies IP and port (using format \<ip\>:\<port\>) of a node that is already on network, if not specified will be created a new network with a single node;

### Running Locally ###

To run an example locally, you can run the following command:

```sh
./execute.sh -n wesley -p 8000
```

This command will create a new chat network with a single node named `wesley` listening on port `8000`

So, you can add a new node to that network specifying `wesley`\`s IP and port:

```sh
./execute.sh -n john -p 8001 -j localhost:8000 //open a new terminal to run this command
```

At this point, a new node named `john` are inserted on chat where `wesley` knows about `john` and `john` knows about `wesley`

If a new node enters on this network this node will know about `wesley` and `john` and they will know about the new node too. 

When a node sends a message, all nodes on network receives that. If a node sends a message and the destination does not receive that, the current node starts a disconnection election sending a disconnection test message to all other nodes asking to them if they have connection to possible diconnected node yet. If 2/3 of network confirms node disconection, node are disconected.