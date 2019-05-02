package lb

import (
	"crypto/md5"
	"encoding/binary"
)

// IPHash 根据客户端ip计算hash code，然后取在线服务器数量的模得到N，然后转发到第N台服务器
type IPHash struct {
}

// GetAddress implements IBalancePolicy
func (iphash *IPHash) GetAddress(localAddress string, remoteAddresses []string) string {
	if len(remoteAddresses) == 0 {
		return ""
	}

	hash := md5.Sum([]byte(localAddress))
	index := binary.BigEndian.Uint32(hash[:]) % uint32(len(remoteAddresses))
	return remoteAddresses[index]
}
