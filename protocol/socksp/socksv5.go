package socksp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

const SOCKS_5_VERSION_NUMBER = uint8(5)

const (
	SOCKS_5_AUTH_NONE = uint8(iota)
	SOCKS_5_AUTH_GSSAPI
	SOCKS_5_AUTH_USERNAME_PASSWORD
	SOCKS_5_AUTH_IANA_ASSIGNED
	SOCKS_5_AUTH_PRIVATE_RSV = uint8(128)
	SOCKS_5_AUTH_NO_METHODS  = uint8(255)
)

const (
	ADDR_TYPE_IPV4   = uint8(1)
	ADDR_TYPE_DOMAIN = uint8(3)
	ADDR_TYPE_IPV6   = uint8(4)
)

const (
	SOCKS_5_RES_SUCCEED = uint8(iota)
	SOCKS_5_RES_FAIL
	SOCKS_5_RES_NOT_ALLOWED
	SOCKS_5_RES_NET_UNREACHABLE
	SOCKS_5_RES_HOST_UNREACHABLE
	SOCKS_5_RES_CONN_REFUSED
	SOCKS_5_RES_TTL_EXPIRED
	SOCKS_5_RES_CMD_NOT_SUPPORT
	SOCKS_5_RES_ADDR_NOT_SUPPORT

	SOCKS_5_RES_BOTTOM
)

const (
	SOCKS_5_REQ_CMD_CONNECT = uint8(iota + 1)
	SOCKS_5_REQ_CMD_BIND
	SOCKS_5_REQ_CMD_UDP
)

type socksV5AuthReq struct {
	version   uint8
	numMethod uint8
	methods   []byte
}

// type socksV5AuthRes struct {
// 	version uint8
// 	method  uint8
// }

type socksV5SubAuthReq struct{}

type socksV5SubAuthRes struct {
	version uint8
	status  uint8
}

type socksV5Req struct {
	version  uint8
	cmd      uint8
	rsv      uint8
	addrType uint8
	addr     []byte
	port     uint16 // 大端
}

type socksV5Res struct {
	version  uint8
	status   uint8
	rsv      uint8
	addrType uint8
	addr     []byte
	port     uint16
}

func encodeV5Res(res *socksV5Res) []byte {
	var data []byte
	data = append(data, res.version, res.status, res.rsv, res.addrType)
	data = append(data, res.addr...)
	data = append(data, []byte{byte(res.port >> 8), byte(res.port << 8)}...)
	return data
}

func handleV5ConnectIP(req *socksV5Req, conn net.Conn, ipv4 bool) error {
	var newConn *net.TCPConn
	var err error
	res := &socksV5Res{
		version: SOCKS_5_VERSION_NUMBER,
		addr:    req.addr,
		port:    req.port,
	}

	if ipv4 {
		res.addrType = ADDR_TYPE_IPV4
	} else {
		res.addrType = ADDR_TYPE_IPV6
	}

	newConn, err = net.DialTCP("tcp", nil, &net.TCPAddr{IP: req.addr, Port: int(req.port)})
	if err == nil {
		res.status = SOCKS_5_RES_SUCCEED
		_, err = conn.Write(encodeV5Res(res))
		if err != nil {
			conn.Close()
			newConn.Close()
			return err
		}
		go func() {
			defer conn.Close()
			io.Copy(conn, newConn)
		}()
		io.Copy(newConn, conn)
		newConn.Close()
		return nil
	} else {
		res.status = SOCKS_5_RES_HOST_UNREACHABLE
		conn.Write(encodeV5Res(res))
		conn.Close()
		return errors.New("SOCKS_5_RES_HOST_UNREACHABLE")
	}
}

func handleV5ConnectDomain(req *socksV5Req, conn net.Conn) error {
	var newConn *net.TCPConn
	var err error
	res := &socksV5Res{
		version: SOCKS_5_VERSION_NUMBER,
		port:    req.port,
	}

	ips, err := net.LookupIP(string(req.addr))
	if err != nil {
		conn.Close()
		return err
	}

	for _, ip := range ips {
		newConn, err = net.DialTCP("tcp", nil, &net.TCPAddr{IP: ip, Port: int(req.port)})
		if err == nil {
			res.status = SOCKS_5_RES_SUCCEED
			if len(ip) == 4 {
				res.addrType = ADDR_TYPE_IPV4
			} else {
				res.addrType = ADDR_TYPE_IPV6
			}
			res.addr = ip
			_, err = conn.Write(encodeV5Res(res))
			if err != nil {
				conn.Close()
				newConn.Close()
				return err
			}
			go func() {
				defer conn.Close()
				io.Copy(conn, newConn)
			}()
			io.Copy(newConn, conn)
			newConn.Close()
			return nil
		}
	}

	res.status = SOCKS_5_RES_HOST_UNREACHABLE
	conn.Write(encodeV5Res(res))
	conn.Close()
	return errors.New("SOCKS_5_RES_HOST_UNREACHABLE")
}

func handleV5Connect(req *socksV5Req, conn net.Conn) error {
	switch req.addrType {
	case ADDR_TYPE_DOMAIN:
		return handleV5ConnectDomain(req, conn)
	case ADDR_TYPE_IPV4:
		return handleV5ConnectIP(req, conn, true)
	case ADDR_TYPE_IPV6:
		return handleV5ConnectIP(req, conn, false)
	default:
		return fmt.Errorf("bad address type: %v", req.addrType)
	}
}

func decodeV5Req(conn net.Conn) (*socksV5Req, error) {
	msg := make([]byte, MSG_MAX_lEN)
	num, err := conn.Read(msg)
	if err != nil {
		return nil, err
	}

	return &socksV5Req{
		version:  SOCKS_5_VERSION_NUMBER,
		cmd:      msg[1],
		addrType: msg[3],
		addr:     msg[4 : num-2],
		port:     binary.BigEndian.Uint16(msg[num-2 : num]),
	}, nil
}

func handleV5AuthNone(conn net.Conn) error {
	// err := binary.Write(conn, binary.BigEndian, &socksV5AuthRes{version: SOCKS_5_VERSION_NUMBER, method: SOCKS_5_AUTH_NONE})
	_, err := conn.Write([]byte{SOCKS_5_VERSION_NUMBER, SOCKS_5_AUTH_NONE})
	if err != nil {
		conn.Close()
		return err
	}

	req, err := decodeV5Req(conn)
	if err != nil {
		conn.Close()
		return err
	}

	switch req.cmd {
	case SOCKS_5_REQ_CMD_CONNECT:
		return handleV5Connect(req, conn)
	case SOCKS_5_REQ_CMD_BIND:
		return nil
	case SOCKS_5_REQ_CMD_UDP:
		return nil
	default:
		return fmt.Errorf("bad cmd: %v", req.cmd)
	}
}

func parseV5AuthReq(msg []byte) *socksV5AuthReq {
	var req socksV5AuthReq
	req.version = msg[0]
	req.numMethod = msg[1]
	req.methods = msg[2 : 2+req.numMethod]

	return &req
}

func handleV5(conn net.Conn, msg []byte) error {
	authReq := parseV5AuthReq(msg)
	for _, v := range authReq.methods {
		switch v {
		case SOCKS_5_AUTH_NONE:
			return handleV5AuthNone(conn)
		}
	}

	conn.Write([]byte{SOCKS_5_VERSION_NUMBER, SOCKS_5_AUTH_NO_METHODS})
	conn.Close()

	return errors.New("there is no suitable authentication method")
}
