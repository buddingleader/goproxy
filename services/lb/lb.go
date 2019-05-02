package lb

import (
	"github.com/pkg/errors"
)

// IBalancePolicy 支持ha Round-Robin ip_hash 多种LB策略，可根据系统配置切换
type IBalancePolicy interface {
	GetAddress(localAddress string, remoteAddresses []string) string
}

// PolicyFactory 使用工厂模式创建实例
type PolicyFactory struct {
	policyMap map[PolicyStatus]IBalancePolicy
}

// InitFactory initialize the load balance factory
func InitFactory() *PolicyFactory {
	factory := &PolicyFactory{make(map[PolicyStatus]IBalancePolicy)}

	factory.policyMap[HA] = &Ha{}
	factory.policyMap[ROUNDROBIN] = &RoundRobin{}
	factory.policyMap[IPHASH] = &IPHash{}
	return factory
}

// GetLBPolicy returns a lb policy
func (factory *PolicyFactory) GetLBPolicy(policy PolicyStatus) (IBalancePolicy, error) {
	lbPolicy, ok := factory.policyMap[policy]
	if !ok {
		return nil, errors.Errorf("Could not find IBalancePolicy, no '%s' provider", policy)
	}

	return lbPolicy, nil
}

// PolicyStatus LB策略算法
type PolicyStatus int

func (status PolicyStatus) String() string {
	switch status {
	case HA:
		return "ha"
	case ROUNDROBIN:
		return "round-robin"
	case IPHASH:
		return "ip_hash"
	default:
		return "unknown"
	}
}

// ha - 热备，总是把所有请求转发到在线服务器列表的首台服务器，直至其掉线移除
// round-robin - 循环，每一次把来自用户的请求轮流分配给所有在线服务器，从1开始，直到N(内部服务器个数)，然后重新开始循环。
// ip_hash - 根据客户端ip计算hash code，然后取在线服务器数量的模得到N，然后转发到第N台服务器
const (
	UNKNOWN PolicyStatus = iota
	HA
	ROUNDROBIN
	IPHASH
)

// PolicyNames for int convert to PolicyStatus
var PolicyNames = map[int]PolicyStatus{
	0: UNKNOWN,
	1: HA,
	2: ROUNDROBIN,
	3: IPHASH,
}
