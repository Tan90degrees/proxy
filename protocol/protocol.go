package protocol

type BasicRequest struct {
	Cmd      uint8
	AddrType uint8
	Addr     []byte
	Port     uint16 // 大端
}

type BasicResponse struct {
}

type ProtocolConn interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}
