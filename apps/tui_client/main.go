package main

import (
	"flag"
	"pim/pkg/pim_client"
)

var rpcServer = flag.String("rpc", "127.0.0.1:18404", "server")

func main() {
	flag.Parse()
	pim_client.RunClient(*rpcServer, "session.txt")
}
