Until now, this program makes a chat connection. 

To run its required to install GoLang on machine and execute the following commands on terminal:

```
go run main.go -n "terminal1" -p 8000
```

This command starts a peer connection named "terminal1" listen on port 8000 via HTTP.

and after, run on a second terminal the following command: 

```
go run main.go -n "terminal2" -p 8002 -j localhost:8000
```

This commad initiates a peer with name "terminal2" that listens on HTTP pot 8002 and knows a peer running on localhost:8000, it connects to known peer and join the network. 

