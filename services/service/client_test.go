package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_SendKeepAlivePackage(t *testing.T) {
	tcpPort := "9999"
	go StartService(tcpPort)

	remoteAddress := "127.0.0.1:11120"
	go startRemoteForTests(remoteAddress)

	time.Sleep(200 * time.Millisecond)
	err := SendKeepAlivePackage(tcpPort, remoteAddress)
	assert.NoError(t, err)
}

func Test_SendGetAllAliveServerAddressesPackage(t *testing.T) {
	tcpPort := "9998"
	go StartService(tcpPort)

	remoteAddress := "127.0.0.1:11121"
	go startRemoteForTests(remoteAddress)

	time.Sleep(200 * time.Millisecond)
	err := SendKeepAlivePackage(tcpPort, remoteAddress)
	assert.NoError(t, err)

	addresses, err := SendGetAllAliveServerAddressesPackage(tcpPort)
	assert.NoError(t, err)
	assert.NotEmpty(t, addresses)

	index := 0
	for i, address := range addresses {
		if address == remoteAddress {
			index = i
			continue
		}
	}
	assert.Equal(t, remoteAddress, addresses[index])
}

func Test_SendStopListenPackage(t *testing.T) {
	tcpPort := "9997"
	go StartService(tcpPort)

	remoteAddress := "127.0.0.1:11122"
	go startRemoteForTests(remoteAddress)

	time.Sleep(200 * time.Millisecond)
	err := SendKeepAlivePackage(tcpPort, remoteAddress)
	assert.NoError(t, err)

	err = SendStopListenPackage(tcpPort)
	assert.NoError(t, err)

	err = SendKeepAlivePackage(tcpPort, remoteAddress)
	assert.Error(t, err)

	// restart
	go StartService(tcpPort)

	time.Sleep(200 * time.Millisecond)
	err = SendKeepAlivePackage(tcpPort, remoteAddress)
	assert.NoError(t, err)
}
