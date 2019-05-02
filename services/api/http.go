package api

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wangff15386/goproxy/services/service"
)

// KeepAliveServer worker-keepalive.do?group=<监听端口>&server=<host>:<port> 接收服务器注册和心跳 更新在线服务器列表
func KeepAliveServer(c *gin.Context) {
	tcpPort, remoteAddress := c.Query("group"), c.Query("server")
	err := service.SendKeepAlivePackage(tcpPort, remoteAddress)
	if err != nil {
		response(c, gin.H{"ok": false, "msg": err.Error()})
		return
	}

	response(c, gin.H{"ok": true})
}

// GetList worker-list.do?group=<监听端口> 查看在线服务器列表
func GetList(c *gin.Context) {
	tcpPort := c.Query("group")
	list, err := service.SendGetAllAliveServerAddressesPackage(tcpPort)
	if err != nil {
		response(c, gin.H{"ok": false, "msg": err.Error()})
		return
	}

	response(c, gin.H{"ok": true, "list": list})
}

// OpenGroup group-open.do?group=<监听端口> 打开端口监听
func OpenGroup(c *gin.Context) {
	tcpPort := c.Query("group")
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", tcpPort))
	if err != nil {
		response(c, gin.H{"ok": false, "msg": err.Error()})
		return
	}
	lis.Close()

	go service.StartService(tcpPort)
	response(c, gin.H{"ok": true})
}

// CloseGroup group-close.do?group=<监听端口> 关闭端口监听
func CloseGroup(c *gin.Context) {
	tcpPort := c.Query("group")
	err := service.SendStopListenPackage(tcpPort)
	if err != nil {
		response(c, gin.H{"ok": false, "msg": err.Error()})
		return
	}

	response(c, gin.H{"ok": true})
}

func response(c *gin.Context, result interface{}) {
	log.Println(c.Request.RequestURI, "response:", result)
	c.JSON(http.StatusOK, result)
}
