package lb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RoundRobin(t *testing.T) {
	rr := RoundRobin{}
	for index := 0; index < 10; index++ {
		address := rr.GetAddress("", remoteAddressesForTests)
		assert.Equal(t, remoteAddressesForTests[index%len(remoteAddressesForTests)], address)
	}

	newAddress := []string{remoteAddressesForTests[1], remoteAddressesForTests[2]}
	for index := 0; index < 10; index++ {
		address := rr.GetAddress("", newAddress)
		assert.Equal(t, newAddress[index%len(newAddress)], address)
		assert.Equal(t, remoteAddressesForTests[index%len(newAddress)+1], address)
	}
}
