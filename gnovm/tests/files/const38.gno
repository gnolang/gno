package main

import "net"

func main() {
	v := net.TCPAddr{}
	const c = v.IP
}

// Error:
// main/files/const38.gno:7:8: v<VPBlock(1,0)>.IP (variable of type gonative{net.TCPAddr}) is not constant