package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	p2pconn "github.com/gnolang/gno/tm2/pkg/p2p/conn"
)

var (
	remote string
	listen string
)

func init() {
	flag.StringVar(&listen, "listen", "", "set to :port if server, eg :8080")
	flag.StringVar(&remote, "remote", "", "remote ip:port")
	flag.Parse()
}

func main() {
	if listen != "" {
		fmt.Println("listening at", listen)
		ln, err := net.Listen("tcp", listen)
		if err != nil {
			// handle error
		}
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		handleConnection(conn)
	} else {
		// connect to remote.
		if remote == "" {
			panic("must specify remote ip:port unless server")
		}
		fmt.Println("connecting to", remote)
		conn, err := net.Dial("tcp", remote)
		if err != nil {
			panic(err)
		}
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	priv := ed25519.GenPrivKey()
	pub := priv.PubKey()
	fmt.Println("local pubkey:", pub)
	fmt.Println("local pubkey addr:", pub.Address())

	sconn, err := p2pconn.MakeSecretConnection(conn, priv)
	if err != nil {
		panic(err)
	}
	// Read line from sconn and print.
	go func() {
		sc := bufio.NewScanner(sconn)
		for sc.Scan() {
			line := sc.Text() // GET the line string
			fmt.Println(">>", line)
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
	}()
	// Read line from stdin and write.
	for {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := sc.Text() + "\n"
			_, err := sconn.Write([]byte(line))
			if err != nil {
				panic(err)
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
	}
}
