package rabbit

import "github.com/chi3ndd/library/clock"

type (
	Service interface {
		DeclareExchange(name, kind string, durable bool) error
		DeclareQueue(name string, durable bool, priority int, ttl clock.Duration) error
		BindQueue(queue, exchange string) error
		Publish(exchange, queue string, message Message) error
		Consume(queue string, auto bool, prefetchCount int, callback Consumer) error
	}
	Consumer func([]byte) error
)
