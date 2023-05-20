package socksp

import (
	"log"
	"net"
	"time"
)

const (
	MIN_CONNECT_MSG_LEN = 8
	MSG_MAX_lEN         = 2048
	READ_TIME_OUT       = time.Second * 5
	WRITE_TIME_OUT      = time.Second * 5
	DIAL_TIME_OUT       = time.Second * 5
)

func solve(conn net.Conn) {
	msg := make([]byte, MSG_MAX_lEN)
	// err := conn.SetReadDeadline(time.Now().Add(READ_TIME_OUT))
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	_, err := conn.Read(msg)
	if err != nil {
		log.Println(err)
		return
	}
	// if num <= MIN_CONNECT_MSG_LEN {
	// 	log.Println("Wrong message from", conn.RemoteAddr())
	// 	return
	// }
	switch msg[0] {
	case SOCKS_4_VERSION_NUMBER:
		handleV4(conn, msg)
		return
	case SOCKS_5_VERSION_NUMBER:
		err = handleV5(conn, msg)
		if err != nil {
			log.Println(err)
		}
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
		// log.Println(conn.RemoteAddr().String())
		go solve(conn)
	}
}
