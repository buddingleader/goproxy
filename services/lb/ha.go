package lb

// Ha 热备，总是把所有请求转发到在线服务器列表的首台服务器，直至其掉线移除
type Ha struct {
}

// GetAddress implements IBalancePolicy
func (ha *Ha) GetAddress(localAddress string, remoteAddresses []string) string {
	if len(remoteAddresses) == 0 {
		return ""
	}

	return remoteAddresses[0]
}
