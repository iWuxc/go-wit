package broker

import (
	"github.com/iWuxc/go-wit/queue/broker/redis"
	"github.com/iWuxc/go-wit/queue/contract"
)

// Broker . get one contract.BrokerInterface
func Broker() contract.BrokerInterface {
	return redis.NewRedisBroker()
}
