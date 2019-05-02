package discovery

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ServiceDiscovery(t *testing.T) {
	stopChan := make(chan struct{})
	disc := NewServiceDiscovery(stopChan)

	assert.Equal(t, 0, len(disc.aliveLastSeen))
	assert.Equal(t, 0, len(disc.remoteAddresses))

	// check aliveLastSeen and remoteAddresses
	remoteAddress := "127.0.0.1:11110"
	disc.HandleAliveMessage("127.0.0.1:11110")
	assert.Equal(t, 1, len(disc.aliveLastSeen))
	assert.Equal(t, 1, len(disc.remoteAddresses))
	now, known := disc.aliveLastSeen[remoteAddress]
	assert.True(t, known)
	_, known = disc.remoteAddresses[remoteAddress]
	assert.True(t, known)

	// update lastseen
	disc.HandleAliveMessage("127.0.0.1:11110")
	now1, known1 := disc.aliveLastSeen[remoteAddress]
	assert.True(t, known1)
	assert.True(t, now1.After(now))
	_, known1 = disc.remoteAddresses[remoteAddress]
	assert.True(t, known1)

	// alive timeout
	time.Sleep(disc.conf.HeartbeatKeepAlive + disc.conf.AliveCheckInterval)
	_, known2 := disc.aliveLastSeen[remoteAddress]
	assert.False(t, known2)
	_, known2 = disc.remoteAddresses[remoteAddress]
	assert.False(t, known2)
}
