package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/wangff15386/goproxy/config"
	"github.com/wangff15386/goproxy/services/discovery"
	"github.com/wangff15386/goproxy/services/lb"
)

// TCP package type 1: HeartBeat, 2: GetAllAliveServers, 3: StopListen, other: ReverseProxy
const (
	HEARTBEAT = iota + 1
	GETALLALIVESERVERS
	STOPLISTEN
)

// TCPPackage for proxy service
type TCPPackage struct {
	Type    int    `json:"type"`
	Content []byte `json:"content"`
}

// TCPProxySession 当有client连接进来时 创建TCPProxySession对象
type TCPProxySession struct {
	net.Conn
}

// TCPProxySessionService 所有TCPProxySession使用ProxySessionService进行状态监测和生命周期管理
type TCPProxySessionService struct {
	disc          *discovery.Service
	lbFactory     *lb.PolicyFactory
	proxySessions map[string]*TCPProxySession
	lock          sync.RWMutex
	listenr       net.Listener
	conf          config.ProxyConfig
	stopChan      chan struct{}
}

func newTCPProxyService() *TCPProxySessionService {
	stopChan := make(chan struct{})

	return &TCPProxySessionService{
		disc:          discovery.NewServiceDiscovery(stopChan),
		lbFactory:     lb.InitFactory(),
		proxySessions: make(map[string]*TCPProxySession, 0),
		conf:          config.GetConfig(),
		stopChan:      stopChan,
	}
}

// StartService start the TCP proxy session service
func StartService(tcpPort string) error {
	log.Println("Starting proxy service")

	service := newTCPProxyService()
	if tcpPort == "" {
		tcpPort = service.conf.TCPPort
	}

	var err error
	service.listenr, err = net.Listen("tcp", fmt.Sprintf("localhost:%s", tcpPort))
	if err != nil {
		return fmt.Errorf("Error to listen tcp service, port: %s, err: %s", tcpPort, err)
	}
	log.Println("Start to listen tcp port:", tcpPort)

	go service.periodicalPrint()

	for {
		select {
		case <-service.stopChan:
			return nil
		default:
		}

		conn, err := service.listenr.Accept()
		if err != nil {
			log.Printf("Error to establish connection :%v\n", err)
			continue
		}

		go service.handleConn(conn)
	}
}

// 每隔5秒定时打印日志：在线client，在线server
func (service *TCPProxySessionService) periodicalPrint() {
	log.Println("Starting proxy service periodical print")

	ticker := time.NewTicker(service.conf.PrintInterval)
	for {
		select {
		case <-service.stopChan:
			log.Println("Stopped proxy service periodical print")
			return
		case <-ticker.C:
			service.lock.RLock()
			sessions := service.proxySessions
			service.lock.RUnlock()

			clients := make([]string, 0)
			for client := range sessions {
				clients = append(clients, client)
			}
			sort.Strings(clients)

			servers := service.disc.GetAllAliveRemoteAddresses()
			log.Printf("在线clients: %v, 在线servers: %v", clients, servers)
		}
	}
}

func (service *TCPProxySessionService) handleConn(conn net.Conn) {
	clientProxySession := service.add(conn)
	defer service.close(clientProxySession)

	for {
		clientProxySession.SetDeadline(time.Now().Add(service.conf.RWTimeout))
		buffer := make([]byte, service.conf.HandleBuffer)
		n, err := clientProxySession.Read(buffer)
		if err == io.EOF {
			// log.Println("Successfully read the client data, address:", clientProxySession.RemoteAddr())
			return
		}

		if err != nil {
			log.Printf("Error to read client tcp package: %v\n", err)
			return
		}
		log.Printf("Successfully reading tcp package to proxy service, address: %v, package: %v\n", clientProxySession.RemoteAddr(), buffer[:n])

		var tcpPackage TCPPackage
		if err = json.Unmarshal(buffer[:n], &tcpPackage); err != nil {
			go service.handleReverseProxyPackage(clientProxySession, buffer[:n])
			continue
		}

		switch tcpPackage.Type {
		case HEARTBEAT:
			go service.handleKeepAlivePackage(string(tcpPackage.Content))
		case GETALLALIVESERVERS:
			go service.handleGetAllAliveRemoteAddressesPackage(clientProxySession)
		case STOPLISTEN:
			go service.handleStopListenPackage()
		default:
			go service.handleReverseProxyPackage(clientProxySession, buffer[:n])
		}
	}
}

func (service *TCPProxySessionService) add(conn net.Conn) *TCPProxySession {
	service.lock.Lock()
	defer service.lock.Unlock()

	clientProxySession := &TCPProxySession{conn}
	service.proxySessions[clientProxySession.RemoteAddr().String()] = clientProxySession
	return clientProxySession
}

func (service *TCPProxySessionService) close(clientProxySession *TCPProxySession) error {
	service.lock.Lock()
	defer service.lock.Unlock()

	address := clientProxySession.RemoteAddr().String()
	delete(service.proxySessions, address)
	log.Println("Close the client connection, address:", address)
	return clientProxySession.Close()
}

func (service *TCPProxySessionService) handleReverseProxyPackage(clientProxySession *TCPProxySession, data []byte) {
	// log.Println("handle reverse proxy package from remote address:", clientProxySession.RemoteAddr())

	policyStatus := lb.PolicyNames[service.conf.LBPolicy]
	lbPolicy, err := service.lbFactory.GetLBPolicy(policyStatus)
	if err != nil {
		log.Printf("Error to get load balance policy, status: %s, error:%s\n", policyStatus, err)
		return
	}

	address := lbPolicy.GetAddress(clientProxySession.RemoteAddr().String(), service.disc.GetAllAliveRemoteAddresses())
	serverConn, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("Error to dial connects to the remote address: %s, error: %s\n", address, err)
		return
	}
	log.Printf("Create a server connetion, clientAddr: %s, proxyAddr: %s, remoteAddr: %s\n", clientProxySession.RemoteAddr(), serverConn.LocalAddr(), address)
	defer func() {
		serverConn.Close()
		log.Println("Close a server connetion, address:", address)
	}()

	errc := make(chan error)
	go service.sendPackageToRemoteServer(clientProxySession, serverConn, data, errc)
	go service.readPackageFromRemoteServer(clientProxySession, serverConn, errc)

	exit := <-errc
	switch exit {
	case io.EOF:
	default:
		log.Println(exit)
	}
}

func (service *TCPProxySessionService) sendPackageToRemoteServer(clientProxySession *TCPProxySession, serverConn net.Conn, data []byte, errc chan error) {
	serverConn.SetDeadline(time.Now().Add(service.conf.RWTimeout))

	if _, err := serverConn.Write(data); err != nil {
		errc <- fmt.Errorf("Error to write client data to remote server, error: %s", err)
	}
}

func (service *TCPProxySessionService) readPackageFromRemoteServer(clientProxySession *TCPProxySession, serverConn net.Conn, errc chan error) {
	for {
		serverConn.SetDeadline(time.Now().Add(service.conf.RWTimeout))

		buffer := make([]byte, 1024)
		n, err := serverConn.Read(buffer)
		if err == io.EOF {
			errc <- io.EOF
			return
		}

		if err != nil {
			errc <- fmt.Errorf("Error to read tcp package from remote server, error: %s", err)
			return
		}

		clientProxySession.SetDeadline(time.Now().Add(service.conf.RWTimeout))
		if _, err = clientProxySession.Write(buffer[:n]); err != nil {
			errc <- fmt.Errorf("Error to write tcp package to client, error: %s", err)
			return
		}
	}
}

func (service *TCPProxySessionService) handleKeepAlivePackage(address string) {
	// log.Println("Receive a keep alive package from remote address:", address)

	service.disc.HandleAliveMessage(address)
}

func (service *TCPProxySessionService) handleGetAllAliveRemoteAddressesPackage(clientProxySession *TCPProxySession) {
	// log.Println("Receive a get all alive remote server addresses package")

	addresses := service.disc.GetAllAliveRemoteAddresses()
	data, err := json.Marshal(addresses)
	if err != nil {
		log.Printf("Error to marshal all alive remote server addresses to []byte, addresses: %v, error: %v\n", addresses, err)
		return
	}

	if _, err = clientProxySession.Write(data); err != nil {
		log.Printf("Error to write all alive remote addresses to client, address: %v, addresses: %v\n", clientProxySession.RemoteAddr(), addresses)
	}
}

func (service *TCPProxySessionService) handleStopListenPackage() {
	log.Println("Stopping proxy service")
	defer log.Println("Stopped proxy service")

	close(service.stopChan)

	err := service.listenr.Close()
	if err != nil {
		log.Println("Error to close proxy service lisener, error:", err)
	}

	for _, clientProxySession := range service.proxySessions {
		if err = service.close(clientProxySession); err != nil {
			log.Println("Error to close client proxy session, error:", err)
		}
	}

}
