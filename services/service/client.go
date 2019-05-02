package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/wangff15386/goproxy/config"
)

// SendKeepAlivePackage send a keep alive package to proxy service
func SendKeepAlivePackage(listenPort, remoteAddress string) error {
	conn, err := newLocalClientConn(listenPort)
	if err != nil {
		return fmt.Errorf("Error to dail connects to proxy service, error: %s", err)
	}
	defer conn.Close()

	tcpPackage := &TCPPackage{
		Type:    HEARTBEAT,
		Content: []byte(remoteAddress),
	}

	errc := make(chan error)
	go writeTCPPackageToProxyService(conn, tcpPackage, errc)

	return <-errc
}

func newLocalClientConn(listenPort string) (net.Conn, error) {
	conf := config.GetConfig()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", listenPort), conf.RWTimeout)
	if err != nil {
		return nil, fmt.Errorf("Error to dail connects to proxy service, error: %s", err)
	}
	conn.SetDeadline(time.Now().Add(conf.RWTimeout))
	log.Printf("Create a client connection, localAddr: %s, remoteAddr: %s\n", conn.LocalAddr(), conn.RemoteAddr())

	return conn, nil
}

// SendGetAllAliveServerAddressesPackage send a get all alive server addresses package to proxy service
func SendGetAllAliveServerAddressesPackage(listenPort string) ([]string, error) {
	conn, err := newLocalClientConn(listenPort)
	if err != nil {
		return nil, fmt.Errorf("Error to dail connects to proxy service, error: %s", err)
	}
	defer conn.Close()

	tcpPackage := &TCPPackage{
		Type: GETALLALIVESERVERS,
	}

	recvBytes, errc := make(chan []byte), make(chan error)
	go writeTCPPackageToProxyService(conn, tcpPackage, errc)
	go readTCPPackageToProxyService(conn, recvBytes, errc)

	for {
		select {
		case err := <-errc:
			if err == nil {
				continue
			}

			return nil, err
		case data := <-recvBytes:
			var addresses []string
			if err = json.Unmarshal(data, &addresses); err != nil {
				return nil, fmt.Errorf("Error to unmarshal all alive server addresses, error: %s", err)
			}

			return addresses, nil
		}
	}
}

// SendStopListenPackage send a keep alive package to proxy service
func SendStopListenPackage(listenPort string) error {
	conn, err := newLocalClientConn(listenPort)
	if err != nil {
		return fmt.Errorf("Error to dail connects to proxy service, error: %s", err)
	}
	defer conn.Close()

	tcpPackage := &TCPPackage{
		Type: STOPLISTEN,
	}

	errc := make(chan error)
	go writeTCPPackageToProxyService(conn, tcpPackage, errc)

	return <-errc
}

func writeTCPPackageToProxyService(conn net.Conn, tcpPackage *TCPPackage, errc chan error) {
	data, err := json.Marshal(tcpPackage)
	if err != nil {
		errc <- fmt.Errorf("Error to marshal tcp package, package:%v, error: %s", tcpPackage, err)
		return
	}

	if _, err := conn.Write(data); err != nil {
		errc <- fmt.Errorf("Error to write tcp package to proxy service, error: %s", err)
		return
	}

	errc <- nil
}

func readTCPPackageToProxyService(conn net.Conn, recvBytes chan []byte, errc chan error) {
	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err == io.EOF {
			recvBytes <- buffer[:n]
			return
		}

		if err != nil {
			errc <- fmt.Errorf("Error to read tcp package from proxy service, error: %s", err)
			return
		}

		recvBytes <- buffer[:n]
		return
	}
}
