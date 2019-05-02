package config

import (
	"testing"
	"time"

	"github.com/wangff15386/goproxy/services/lb"

	"github.com/stretchr/testify/assert"
)

func Test_Config(t *testing.T) {
	assert.NotPanics(t, InitConfig)
	assert.Equal(t, "8080", conf.HTTPPort)
	assert.Equal(t, "8081", conf.TCPPort)
	assert.Equal(t, "8082", conf.PProfPort)
	assert.Equal(t, lb.HA, lb.PolicyStatus(conf.LBPolicy))
	assert.Equal(t, 3*time.Second, conf.RWTimeout)
	assert.Equal(t, 5*time.Second, conf.PrintInterval)
	assert.Equal(t, 5*time.Second, conf.HeartbeatKeepAlive)
	assert.Equal(t, 25*time.Second/10, conf.AliveCheckInterval)
	assert.Equal(t, 1024, conf.HandleBuffer)

	conf1 := GetConfig()
	assert.Equal(t, conf, conf1)

	conf1.HTTPPort = "7777"
	assert.Equal(t, "8080", conf.HTTPPort)
}
