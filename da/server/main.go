package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/chandiniv1/go-da/da"
)

func main() {
	// Create an instance of the RollkitService.
	dataAvailabilityLayerClient := new(da.DataAvailabilityLayerClient)

	// Register the service with the RPC package.
	rpc.Register(dataAvailabilityLayerClient)

	// Listen for incoming RPC requests on a specific network address and port.
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Error starting RPC server:", err)
	}
	defer listener.Close()

	log.Println("RPC server is listening on :1234...")

	// Accept and handle incoming RPC requests.
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error accepting connection:", err)
		}
		go rpc.ServeConn(conn)
	}
}
