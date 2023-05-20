package httpp

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"proxy/myerror"
	"sync"
	"time"
)

func TunnelServer(servPort string) {
	var twg sync.WaitGroup
	listener, err := net.Listen("tcp", ":"+servPort)
	myerror.CheckErrorExit(err)
	defer listener.Close()

	handler := func(clientConn net.Conn) {
		var iwg sync.WaitGroup
		twg.Add(1)
		defer clientConn.Close()

		// 解析HTTP请求头
		header := make([]byte, 1024)
		// buf := make([]byte, 4096)
		cNum, _ := clientConn.Read(header)
		if cNum == 0 {
			return
		}
		methodIndex := bytes.IndexByte(header, 32)
		method := string(header[:methodIndex])

		//建立到目的服务器的隧道
		if method == http.MethodConnect {
			hNoMethod := header[methodIndex+1:]
			urlIndex := bytes.IndexByte(hNoMethod, 32)
			url := string(hNoMethod[:urlIndex])
			clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
			servConn, err := net.DialTimeout("tcp", url, time.Second*5)
			if err != nil {
				return
			}
			defer servConn.Close()
			log.Println(url)

			// 通过隧道传输信息
			iwg.Add(1)
			go func() {
				// num, _ := io.CopyBuffer(servConn, conn, buf)
				num, err := io.Copy(servConn, clientConn)
				if num == 0 && err != nil {
					iwg.Done()
					return
				}
			}()

			// iwg.Add(1)
			// go func() {
			io.Copy(clientConn, servConn)
			// 		if num == 0 && err != nil {
			// 			iwg.Done()
			// 			return
			// 		}
			// 	}()
			// 	iwg.Wait()
		}

		twg.Done()
	}

	for {
		conn, _ := listener.Accept()
		go handler(conn)
	}
}
