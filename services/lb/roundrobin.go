package lb

import (
	"sync"
)

// RoundRobin  循环，每一次把来自用户的请求轮流分配给所有在线服务器，从1开始，直到N(内部服务器个数)，然后重新开始循环。
type RoundRobin struct {
	index int // local read address index
	lock  sync.RWMutex
}

// GetAddress implements IBalancePolicy
func (rr *RoundRobin) GetAddress(localAddress string, remoteAddresses []string) string {
	if len(remoteAddresses) == 0 {
		return ""
	}

	rr.lock.Lock()
	defer func() {
		rr.index++
		rr.lock.Unlock()
	}()

	// Check if the index is beyond the addresses range
	if rr.index >= len(remoteAddresses) {
		rr.index = 0
	}

	return remoteAddresses[rr.index]
}
