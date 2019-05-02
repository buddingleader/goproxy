package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wangff15386/goproxy/config"
	"github.com/wangff15386/goproxy/services/api"
	"github.com/wangff15386/goproxy/services/service"
)

func main() {
	conf := config.GetConfig()

	go service.StartService(conf.TCPPort)
	gracefulStartHTTP(conf, setupRouter())
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// Allows all origins
	r.Use(cors.Default())

	// worker-keepalive.do?group=<监听端口>&server=<host>:<port> 接收服务器注册和心跳 更新在线服务器列表
	r.POST("/worker-keepalive.do", api.KeepAliveServer)
	// worker-list.do?group=<监听端口> 查看在线服务器列表
	r.GET("worker-list.do", api.GetList)
	// group-open.do?group=<监听端口> 打开端口监听
	r.POST("group-open.do", api.OpenGroup)
	// group-close.do?group=<监听端口> 关闭端口监听
	r.POST("group-close.do", api.CloseGroup)
	return r
}

func gracefulStartHTTP(conf config.ProxyConfig, router *gin.Engine) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", conf.HTTPPort),
		Handler: router,
	}

	go func() {
		log.Println("Start to listen http address:", srv.Addr)

		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Increase pprof to facilitate memory diagnostics
	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf(":%s", conf.PProfPort), nil))
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")
	service.SendStopListenPackage(conf.TCPPort)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}
