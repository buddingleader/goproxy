package lb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IPHash(t *testing.T) {
	iphash := IPHash{}
	address := iphash.GetAddress("127.0.0.1:8080", remoteAddressesForTests)
	for index := 0; index < 10; index++ {
		address1 := iphash.GetAddress("127.0.0.1:8080", remoteAddressesForTests)
		assert.Equal(t, address, address1)
	}
}
