package socketp

import (
	"log"
	"net"
)

const (
	MIN_CONNECT_MSG_LEN = 8
	MSG_MAX_lEN         = 256
)

func solve(conn net.Conn) {
	msg := make([]byte, MSG_MAX_lEN)
	num, err := conn.Read(msg)
	if err != nil {
		log.Println(err)
		return
	}
	if num <= MIN_CONNECT_MSG_LEN {
		log.Println("Wrong message from", conn.RemoteAddr())
		return
	}
	switch msg[0] {
	case 4:
		handleV4(conn, &msg)
		return
	case 5:
		handleV5(conn, &msg)
		return
	default:
		return
	}
}

func Server() {
	ln, err := net.Listen("tcp", "0.0.0.0:10086")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go solve(conn)
	}
}
