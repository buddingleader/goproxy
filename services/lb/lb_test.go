package lb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PolicyFactory(t *testing.T) {
	factory := InitFactory()

	for _, policy := range factory.policyMap {
		address := policy.GetAddress("", remoteAddressesForTests)
		assert.NotEmpty(t, address)
	}

	assert.Equal(t, 3, len(factory.policyMap))
	for index := 1; index <= 3; index++ {
		policy, err := factory.GetLBPolicy(PolicyStatus(index))
		assert.NoError(t, err)
		address := policy.GetAddress("", remoteAddressesForTests)
		assert.NotEmpty(t, address)

	}

	policy, err := factory.GetLBPolicy(PolicyStatus(0))
	assert.Error(t, err)
	assert.Nil(t, policy)
}
