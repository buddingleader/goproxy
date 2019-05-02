package service

import (
	"fmt"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_TCPProxySessionService(t *testing.T) {
	tcpPort := "11101"
	assert.NotPanics(t, func() { go StartService(tcpPort) })

	remoteAddress := "127.0.0.1:11122"
	go startRemoteForTests(remoteAddress)

	stopChan := make(chan struct{})

	//发送心跳的goroutine
	go func() {
		err := SendKeepAlivePackage(tcpPort, remoteAddress)
		assert.NoError(t, err)

		heartBeatTick := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-heartBeatTick.C:
				SendKeepAlivePackage(tcpPort, remoteAddress)
			case <-stopChan:
				return
			}
		}
	}()

	time.Sleep(200 * time.Millisecond)

	//测试用的，开300个goroutine每秒发送一个包
	for i := 0; i < 300; i++ {
		clientConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", tcpPort))
		assert.NoError(t, err)
		if err != nil {
			continue
		}
		go func() {
			sendTimer := time.After(1 * time.Second)
			for {
				select {
				case <-sendTimer:
					clientConn.Write([]byte("666"))
					sendTimer = time.After(1 * time.Second)
				case <-stopChan:
					return
				}
			}
		}()
		go func() {
			for {
				buffer := make([]byte, 1024)
				n, err := clientConn.Read(buffer)
				if err == io.EOF {
					return
				}
				assert.NoError(t, err)

				fmt.Println("Receive remote server data:", string(buffer[:n]))
			}
		}()
	}

	go func() {
		time.Sleep(15 * time.Second)
		close(stopChan)
	}()
	//等待退出
	<-stopChan
}

func startRemoteForTests(address string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Panicf("Error to listen tcp service, address: %s, err: %s", address, err)
	}
	defer lis.Close()
	log.Println("Start to listen remote address:", address)

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("Error to establish connection :%v\n", err)
			continue
		}
		// log.Printf("[Remote Server] Receive a client connection, localAddr: %s, remoteAddr: %s\n", conn.LocalAddr(), conn.RemoteAddr())

		go func() {
			for {
				buffer := make([]byte, 1024)
				n, err := conn.Read(buffer)
				if err == io.EOF {
					log.Printf("[Remote Server] readed EOF %s\n", string(buffer[:n]))
					return
				}
				if err != nil {
					log.Printf("[Remote Server] readed error:%v\n", err)
					return
				}

				log.Printf("[Remote Server] readed %s\n", string(buffer[:n]))
			}
		}()

		go func() {
			conn.Write([]byte("Hello world! Client: " + conn.RemoteAddr().String()))
		}()

		time.Sleep(1 * time.Second)
		conn.Close()
	}
}
