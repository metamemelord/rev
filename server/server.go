package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/PlanckProject/go-commons/logger"
)

func Init(port string) {
	addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:"+port)
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Server running on port ", port)
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			logger.Error(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn *net.TCPConn) {
	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		logger.Error(err)
		conn.Close()
		return
	}

	if strings.HasPrefix(request.RequestURI, "/register") {
		conn.SetKeepAlive(true)
		ctx, cancel := context.WithCancel(context.Background())
		segments := strings.Split(strings.TrimPrefix(request.RequestURI, "/register/"), "/")
		if len(segments) < 2 ||
			(len(segments[0]) == 0 || len(segments[1]) == 0) {
			conn.Write([]byte("HTTP/1.1 400 Bad request\n\nInvalid request"))
			conn.Close()
			cancel()
			return
		}
		serviceConn := &ServiceConnection{Ctx: ctx, CancelFunc: cancel, Conn: *conn}
		AddConnection(fmt.Sprintf("%s/%s", segments[0], segments[1]), serviceConn)
	} else if strings.HasPrefix(request.RequestURI, "/messages") {
		// TODO: HANDLE ERRORS in this section
		segments := strings.Split(strings.TrimPrefix(request.RequestURI, "/messages/"), "/")
		if len(segments) < 2 ||
			(len(segments[0]) == 0 || len(segments[1]) == 0) {
			conn.Write([]byte("HTTP/1.1 400 Bad request\n\nInvalid request"))
			conn.Close()
			return
		}
		serviceConn := GetConnection(fmt.Sprintf("%s/%s", segments[0], segments[1]))
		if serviceConn == nil {
			conn.Write([]byte("HTTP/1.1 404 Not Found\n\n"))
			conn.Close()
			return
		}
		logger.Infof("%s %s", request.Method, "/"+
			strings.Join(strings.Split(request.RequestURI, "/")[4:], "/"))
		rawRequest := getRawRequest(request)
		logger.WithField("raw_request", string(rawRequest)).Debug("Raw request")
		_, err := serviceConn.Conn.Write(rawRequest)
		if err != nil {
			logger.WithField("error", err).Error("Error while calling the downstream service")
			conn.Write([]byte(fmt.Sprintf("%s %d %s", "HTTP/1.1", http.StatusServiceUnavailable,
				"Service unavailable")))
			conn.Close()
			// Inspect and cleanup service here
			return
		}
		_, err = io.Copy(&serviceConn.Conn, request.Body)
		if err != nil {
			logger.WithField("error", err).Error("Error while calling the downstream service")
			conn.Write([]byte(fmt.Sprintf("%s %d %s", "HTTP/1.1", http.StatusServiceUnavailable,
				"Service unavailable")))
			conn.Close()
			// Inspect and cleanup service here
			return
		}
		res, err := http.ReadResponse(bufio.NewReader(&serviceConn.Conn), request)
		if err != nil {
			logger.WithField("error", err).Error("Error while calling the downstream service")
			conn.Write([]byte(fmt.Sprintf("%s %d %s", "HTTP/1.1", http.StatusServiceUnavailable, "Service unavailable")))
			conn.Close()
			// Inspect and cleanup service here
			return
		}
		logger.Info("Response status code: ", res.StatusCode)
		conn.Write([]byte(fmt.Sprintf("%s %d %s\n%s\n\n", res.Proto, res.StatusCode, res.Status, processHeaders(res.Header))))
		io.Copy(conn, res.Body)
		conn.Close()
	} else {
		logger.Warn("Ignored request: ", request.RequestURI)
		logger.Debug(request)
		conn.Write([]byte("HTTP/1.1 404 Not Found\n\n"))
		conn.Close()
	}
}

func getRawRequest(request *http.Request) []byte {
	requestUri := "/" + strings.Join(strings.Split(request.RequestURI, "/")[4:], "/")
	logger.WithField("request_uri", requestUri).Debug("Request URI")
	return []byte(fmt.Sprintf("%s %s %s\n%s\n\n", request.Method, requestUri, request.Proto, processHeaders(request.Header)))
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
