package broker

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// SafeChannel 是线程安全的 RabbitMQ Channel 包装器
// AMQP Channel 不是并发安全的，多个 goroutine 同时使用会导致协议错误
// TODO 改用 channel 池
type SafeChannel struct {
	mu      sync.Mutex
	channel *amqp.Channel
}

// NewSafeChannel 创建一个线程安全的 Channel 包装器
func NewSafeChannel(ch *amqp.Channel) *SafeChannel {
	return &SafeChannel{
		channel: ch,
	}
}

// PublishWithContext 线程安全地发布消息
func (sc *SafeChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	return sc.channel.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}

// QueueDeclare 线程安全地声明队列
func (sc *SafeChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	return sc.channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

// QueueBind 线程安全地绑定队列
func (sc *SafeChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	return sc.channel.QueueBind(name, key, exchange, noWait, args)
}

// Consume 线程安全地消费消息
func (sc *SafeChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	return sc.channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

// ExchangeDeclare 线程安全地声明交换机
func (sc *SafeChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	return sc.channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
}

// Close 关闭 channel
func (sc *SafeChannel) Close() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	return sc.channel.Close()
}

// Channel 返回底层的 AMQP Channel（仅供特殊场景使用，不保证线程安全）
// 注意：直接使用底层 channel 会破坏线程安全保证，应尽量避免
func (sc *SafeChannel) Channel() *amqp.Channel {
	return sc.channel
}
