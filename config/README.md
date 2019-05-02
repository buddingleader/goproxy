# http接口
HTTPPort: "8080"

# tcp端口转发
TCPPort: "8081"

# import "net/http/pprof"以方便内存诊断  
PProfPort: "8082"

# lb策略

> 1: ha - 热备，总是把所有请求转发到在线服务器列表的首台服务器，直至其掉线移除  
> 2: round-robin - 循环，每一次把来自用户的请求轮流分配给所有在线服务器，从1开始，直到N(内部服务器个数)，然后重新开始循环。  
> 3: ip_hash - 根据客户端ip计算hash code，然后取在线服务器数量的模得到N，然后转发到第N台服务器  

LBPolicy: 1

# 长时间(3s)cleint/server无读无写时 注销TCPProxySession  
RWTimeout: "3s"

# 每隔5秒定时打印日志
PrintInterval: "5s"

# 在线服务器心跳保活时间
HeartbeatKeepAlive: 5s

# 每隔2.5秒检查在线服务器是否活着, 超过HeartbeatKeepAlive之后从在线服务器列表删除
AliveCheckInterval: 2.5s

# 代理服务每次处理客户端读取的缓冲区大小, 单位：字节(B)
HandleBuffer: 1024