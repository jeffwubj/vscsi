package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"acln.ro/zerocopy"
	"github.com/mdlayher/vsock"
	flag "github.com/spf13/pflag"
)

const (
	VSockPort = 2222
)

var isServer = flag.BoolP("server", "s", false, "Server")

func main() {

	flag.Parse()

	if *isServer {
		fmt.Println("start server...")
		err := vSockToTCP()
		if err != nil {
			fmt.Errorf(err.Error())
			os.Exit(255)
		}

	} else {
		fmt.Println("start client...")
		err := tcpTovSock()
		if err != nil {
			fmt.Errorf(err.Error())
			os.Exit(255)
		}
	}

	defer func() {
		fmt.Println("exit...")
	}()
}

func tcpTovSock() error {

	listener, err := net.Listen("tcp", "127.0.0.1:3260")
	if err != nil {
		fmt.Printf("listen fail, err: %v\n", err)
		return err
	}

	connTCP, err := listener.Accept()
	if err != nil {
		fmt.Printf("accept fail, err: %v\n", err)
		return err
	}

	connVSOCK, err := send(2, VSockPort)
	defer connVSOCK.Close()

	if err != nil {
		return err
	}

	proxy(connVSOCK, connTCP)

	select {}
	return nil
}

func vSockToTCP() error {
	connISCSI, err := net.Dial("tcp", "127.0.0.1:3260")
	defer connISCSI.Close()
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}

	connvSock, err := receive(VSockPort)
	defer connvSock.Close()
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}

	proxy(connvSock, connISCSI)

	select {}
}

func proxy(upstream, downstream net.Conn) error {
	go zerocopy.Transfer(upstream, downstream)
	go zerocopy.Transfer(downstream, upstream)
	return nil
}

func send(cid, port uint32) (net.Conn, error) {
	fatalf := func(format string, a ...interface{}) {
		log.Fatalf("vscp: send: "+format, a...)
	}

	c, err := vsock.Dial(cid, port, nil)
	if err != nil {
		fatalf("failed to dial: %v", err)
		return nil, err
	}

	return c, nil
}

func receive(port uint32) (net.Conn, error) {
	// Log helper functions.
	logf := func(format string, a ...interface{}) {
		log.Printf("receive: "+format, a...)
	}

	fatalf := func(format string, a ...interface{}) {
		log.Fatalf("vscp: receive: "+format, a...)
	}

	logf("opening listener: %d", port)

	// TODO(mdlayher): support vsock.Local binds for testing.

	l, err := vsock.Listen(port, nil)
	if err != nil {
		fatalf("failed to listen: %v", err)
		return nil, err
	}
	// defer l.Close()

	// Show server's address for setting up client flags.
	log.Printf("receive: listening: %s", l.Addr())

	// Accept a single connection, and receive stream from that connection.
	c, err := l.Accept()
	if err != nil {
		fatalf("failed to accept: %v", err)
		return nil, err
	}
	return c, nil
}
