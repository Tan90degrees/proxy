package main

import (
	"flag"
	"fmt"
	"os"
	"proxy/httpp"
	"proxy/socketp"
)

const DEFAULTPORT string = "10086"

func main() {
	ps := flag.Bool("ps", false, "Start proxy server.")
	ts := flag.Bool("ts", false, "Start tunnel server.")
	socks := flag.Bool("d", false, "")
	port := flag.String("p", DEFAULTPORT, "Set port.")
	flag.Parse()
	if *socks {
		socketp.Server()
	}
	if *ps == *ts {
		fmt.Println("Can only open one server.")
		os.Exit(0)
	}
	if *ps {
		httpp.ProxyServer(*port)
	} else {
		httpp.TunnelServer(*port)
	}
	os.Exit(0)
}
