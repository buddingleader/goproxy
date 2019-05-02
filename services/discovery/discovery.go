package discovery

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/wangff15386/goproxy/config"
)

// Service to discover and manage remote addresses
type Service struct {
	// deadLastSeen  map[string]*timestamp     // H
	aliveLastSeen   map[string]time.Time // V
	remoteAddresses map[string]struct{}  // All known remote service addresses

	lock     sync.RWMutex
	conf     config.ProxyConfig
	stopChan chan struct{}
}

// NewServiceDiscovery returns a new discovery service
func NewServiceDiscovery(stopChan chan struct{}) *Service {
	disc := &Service{
		aliveLastSeen:   make(map[string]time.Time),
		remoteAddresses: make(map[string]struct{}),
		conf:            config.GetConfig(),
		stopChan:        stopChan,
	}

	go disc.periodicalCheckAlive()
	return disc
}

// HandleAliveMessage 接收服务器注册和心跳
func (disc *Service) HandleAliveMessage(remoteAddress string) {
	disc.lock.RLock()
	_, known := disc.remoteAddresses[remoteAddress]
	disc.lock.RUnlock()

	if !known {
		disc.learnNewRemoteAddress(remoteAddress)
		return
	}

	disc.lock.RLock()
	lastAliveTS, isAlive := disc.aliveLastSeen[remoteAddress]
	disc.lock.RUnlock()

	if !isAlive {
		log.Printf("Error remote Address %s is known but not found in alive maps, isAlive=%v", remoteAddress, isAlive)
		return
	}

	// Update the last heartbeat time only when the heartbeat message is newer than the last heartbeat
	if now := time.Now(); lastAliveTS.Before(now) {
		disc.learnExistedRemoteAddress(remoteAddress, now)
	}
}

func (disc *Service) learnNewRemoteAddress(remoteAddress string) {
	disc.lock.Lock()
	defer disc.lock.Unlock()

	disc.remoteAddresses[remoteAddress] = struct{}{}
	disc.aliveLastSeen[remoteAddress] = time.Now()

	log.Printf("Learning a new remote address: %s, lastSeen: %s", remoteAddress, disc.aliveLastSeen[remoteAddress])
}

func (disc *Service) learnExistedRemoteAddress(remoteAddress string, now time.Time) {
	disc.lock.Lock()
	defer disc.lock.Unlock()

	disc.aliveLastSeen[remoteAddress] = now

	log.Printf("Learning a existed remote address: %s, lastSeen: %s", remoteAddress, disc.aliveLastSeen[remoteAddress])
}

// GetAllAliveRemoteAddresses 获取所有在线服务器列表
func (disc *Service) GetAllAliveRemoteAddresses() []string {
	disc.lock.RLock()
	defer disc.lock.RUnlock()

	addresses := make([]string, 0)
	for address := range disc.remoteAddresses {
		addresses = append(addresses, address)
	}

	// Sort addresses in ascending order to ensure that the order of each acquisition is consistent
	// Because you know that the order of traversing the map is random.
	sort.Strings(addresses)
	return addresses
}

func (disc *Service) periodicalCheckAlive() {
	log.Println("Starting discovery periodical check alive")

	ticker := time.NewTicker(disc.conf.AliveCheckInterval)
	for {
		select {
		case <-disc.stopChan:
			log.Println("Stopped discovery periodical check alive")
			return
		case <-ticker.C:
			now := time.Now()
			for address, lastSeen := range disc.aliveLastSeen {
				if lastSeen.Add(disc.conf.HeartbeatKeepAlive).Before(now) {
					disc.expireDeadAddress(address, now)
				}
			}
		}
	}
}

func (disc *Service) expireDeadAddress(remoteAddress string, now time.Time) {
	disc.lock.Lock()
	defer disc.lock.Unlock()

	delete(disc.remoteAddresses, remoteAddress)
	delete(disc.aliveLastSeen, remoteAddress)

	log.Printf("Expired a dead remote address: %s ,at time: %s\n", remoteAddress, now)
}
