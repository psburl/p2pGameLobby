That code simulates a broadcast chat that recognizes new network nodes and propagates their connection to already known nodes.

To use this code you need to run the following command on terminal:

```
go run main.go -n <node_name> -p <node_port> -j <known_node_on_net>
```

The following commands creates, for example, 4 network nodes.

```
go run main.go -n node1 -p 8000 -j localhost:8001
go run main.go -n node2 -p 8001 -j localhost:8000
go run main.go -n node3 -p 8002 -j localhost:8001
go run main.go -n node4 -p 8003 -j localhost:8000
```

On this example all nodes knows all other nodes.

When a node sends a message and the destination does not receive that, the current node starts a disconnection
election sending a disconnection test message to all other nodes asking to them if they have connection to possible diconnected node yet. If 2/3 of network confirms node disconection, node are disconected.