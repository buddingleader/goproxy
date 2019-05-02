package lb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	remoteAddressesForTests = []string{"127.0.0.1:11110", "127.0.0.1:11111", "127.0.0.1:11111"}
)

func Test_HA(t *testing.T) {
	ha := Ha{}
	for index := 0; index < 10; index++ {
		address := ha.GetAddress("", remoteAddressesForTests)
		assert.Equal(t, remoteAddressesForTests[0], address)
	}

	newAddress := []string{remoteAddressesForTests[1]}
	for index := 0; index < 10; index++ {
		address := ha.GetAddress("", newAddress)
		assert.Equal(t, remoteAddressesForTests[1], address)
	}
}
