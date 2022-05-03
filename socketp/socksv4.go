package socketp

import (
	"errors"
	"io"
	"log"
	"net"
)

const (
	REQ_OK = 90 + iota
	REQ_REJ
	REQ_REJ_AUTH_FAIL
	REQ_REJ_ID_DIFF
)

const NETWORK_ERR = "uknown network error"
const MASK_IP_4A uint8 = 15 // 00001111

func rspV4(conn net.Conn, msg *[]byte) error {
	num, err := conn.Write((*msg)[:8])
	if err != nil {
		log.Println(err)
		return err
	}
	if num != 8 {
		log.Println("")
		return errors.New(NETWORK_ERR)
	}
	return nil
}

func handleV4(conn net.Conn, msg *[]byte) {
	// log.Printf("Socks V4 to: %s\n", conn.RemoteAddr())
	var dstAddr *net.TCPAddr
	if (*msg)[4] == 0 && (*msg)[5] == 0 && (*msg)[6] == 0 && (*msg)[7] != 0 {
		begin := 9
		end := 9
		for i := 8; i < MSG_MAX_lEN; i++ {
			if (*msg)[i] == 0 {
				begin = i + 1
				for j := i; j < MSG_MAX_lEN; j++ {
					if (*msg)[j] == 0 {
						end = j
						break
					}
				}
				break
			}
		}
		rIP, err := net.LookupIP(string((*msg)[begin:end]))
		if err != nil {
			log.Println(err)
			(*msg)[1] = REQ_REJ
			err = rspV4(conn, msg)
			if err != nil {
				log.Println(err)
			}
			return
		}
		dstAddr = &net.TCPAddr{IP: rIP[0], Port: (int((*msg)[2]) << 8) + int(((*msg)[3]))}
	} else {
		dstAddr = &net.TCPAddr{IP: (*msg)[4:8], Port: (int((*msg)[2]) << 8) + int(((*msg)[3]))}
	}

	if (*msg)[1] == 1 { // Connect
		(*msg)[0] = 0
		connServ, err := net.DialTCP("tcp", nil, dstAddr)
		if err != nil {
			log.Println(err)
			(*msg)[1] = REQ_REJ
			err = rspV4(conn, msg)
			if err != nil {
				log.Println(err)
			}
			return
		}
		(*msg)[1] = REQ_OK
		err = rspV4(conn, msg)
		if err != nil {
			log.Println(err)
			return
		}
		go func() {
			io.Copy(connServ, conn)
			connServ.Close()
		}()
		io.Copy(conn, connServ)
		conn.Close()
	} else if (*msg)[1] == 2 { // Bind
		// (*msg)[0] = 0
		// ln, err := net.ListenTCP("tcp", &net.TCPAddr{IP: (*msg)[4:8], Port: (int((*msg)[2]) << 8) + int(((*msg)[3]))})
		// if err != nil {
		// 	log.Fatal(err)
		// 	(*msg)[1] = REQ_REJ
		// 	rspV4(conn, msg)
		// 	return
		// }
		return
	}
}
