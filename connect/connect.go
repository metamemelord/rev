package connect

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	httpReq "github.com/PlanckProject/go-commons/http/request"
	"github.com/PlanckProject/go-commons/logger"
)

func Init(serverHost, serverPort, serviceUser, serviceName, downstreamProtocol, downstreamHost, downstreamPort string) {
	defer func() {
		if err := recover(); err != nil {
			logger.Fatal("Unexpected error: ", err)
		}
	}()

	{
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", downstreamHost, downstreamPort))
		if err != nil {
			logger.Error("Downstream service is unreachable")
			return
		}
		_ = conn.Close()
	}
	addrStr := fmt.Sprintf("%s:%s", serverHost, serverPort)
	addr, _ := net.ResolveTCPAddr("tcp", addrStr)
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		logger.Fatal("Error connecting to network ", addrStr)
		return
	}
	conn.SetKeepAlive(true)
	conn.Write([]byte(fmt.Sprintf("POST /register/%s/%s HTTP/1.1\n\n\n", serviceUser, serviceName)))
	logger.Infof("Connected to %s:%s!", serverHost, serverPort)
	for {
		httpRequest, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			if err == io.EOF {
				logger.Warn("EOF received from server, disconnected")
				return
			}
			panic(err)
		}

		downstreamRequestURI := fmt.Sprintf("%s://%s:%s/%s", downstreamProtocol, downstreamHost,
			downstreamPort, strings.TrimPrefix(httpRequest.RequestURI, "/"))
		logger.Debug("Downstream request URI: ", downstreamRequestURI)

		downstreamHttpRequest := httpReq.New().
			SetMethod(httpRequest.Method).
			SetURI(downstreamRequestURI).
			SetRetries(0).
			SetPayloadFromReader(httpRequest.Body)
		for header, value := range httpRequest.Header {
			downstreamHttpRequest.SetHeader(header, strings.Join(value, ","))
		}
		downstreamHttpRequest.SetHeader("X-Tunneled-Via-Rev", "Rev (ver: 0.0.1)")
		res, err := downstreamHttpRequest.Do()
		if err != nil {
			logger.Error(err)
			conn.Write([]byte(fmt.Sprintf("%s %d %s\n%s\n\n", "HTTP/1.1", http.StatusServiceUnavailable, "Service unavailable", processHeaders(res.Header))))
			continue
		}
		res.Header.Add("X-Tunneled-Via-Rev", "Rev (ver: 0.0.1)")
		conn.Write([]byte(fmt.Sprintf("%s %d %s\n%s\n\n", res.Proto, res.StatusCode, res.Status, processHeaders(res.Header))))
		io.Copy(conn, res.Body)
	}
}

func processHeaders(headers http.Header) string {
	headerString := ""
	for header, value := range headers {
		if len(headerString) != 0 {
			headerString += "\n"
		}
		headerString += header + ":"
		headerString += strings.Join(value, ",")
	}
	return headerString
}
