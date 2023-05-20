package httpp

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"proxy/myerror"
)

func ProxyServer(servPort string) {
	handler := func(cRespW http.ResponseWriter, cReq *http.Request) {
		fmt.Println(cReq.URL.String())
		// buf := new(bytes.Buffer)
		buf := make([]byte, 32*1024)
		servClient := new(http.Client)
		servReq, err := http.NewRequest(cReq.Method, cReq.URL.String(), cReq.Body)
		myerror.CheckError(err)
		defer cReq.Body.Close()
		servReq.Header = cReq.Header.Clone()
		servResp, err := servClient.Do(servReq)
		myerror.CheckError(err)
		defer servReq.Body.Close()
		for k, v := range servResp.Header {
			for i := 0; i < len(v); i++ {
				cRespW.Header().Add(k, v[i])
			}
		}
		for {
			// num, err := servResp.Body.Read(buf)
			// cRespW.Write(buf)
			num, err := io.ReadFull(servResp.Body, buf)
			if num == 0 && err != nil {
				break
			}
			num, err = cRespW.Write(buf)
			if num == 0 && err != nil {
				break
			}
		}
		defer servResp.Body.Close()
		// buf.Reset()
	}
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+servPort, nil))
}
