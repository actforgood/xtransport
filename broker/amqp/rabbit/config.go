package rabbit

const (
	ExchangeTypeFanout = "fanout"
	ExchangeTypeDirect = "direct"
)

const (
	DeliveryModeTransient  byte = 0
	DeliveryModePersistent byte = 2
)

const (
	// Message, common, consumer - producer.
	PropMsgContentType     = "amqp.property.msg.contentType"
	PropMsgContentEncoding = "amqp.property.msg.contentEncoding"
	PropMsgDeliveryMode    = "amqp.property.msg.deliveryMode"
	PropMsgPriority        = "amqp.property.msg.priority"
	PropMsgCorrelationID   = "amqp.property.msg.correlationId"
	PropMsgReplyTo         = "amqp.property.msg.replyTo"
	PropMsgExpiration      = "amqp.property.msg.expiration"
	PropMsgMessageID       = "amqp.property.msg.messageId"
	PropMsgTimestamp       = "amqp.property.msg.timestamp"
	PropMsgType            = "amqp.property.msg.type"
	PropMsgUserID          = "amqp.property.msg.userId"
	PropMsgAppID           = "amqp.property.msg.appId"
	PropMsgHeaders         = "amqp.property.msg.headers"

	// Message, consumer.
	PropMsgConsumerTag  = "amqp.property.msg.consumerTag"
	PropMsgMessageCount = "amqp.property.msg.messageCount"
	PropMsgDeliveryTag  = "amqp.property.msg.deliveryTag"
	PropMsgRedelivered  = "amqp.property.msg.redelivered"
	PropMsgExchange     = "amqp.property.msg.exchange"
	PropMsgRoutingKey   = "amqp.property.msg.routingKey"

	// Publishing.
	PropPublishRoutingKey = "amqp.property.publish.routingKey"
	PropPublishImmediate  = "amqp.property.publish.immediate"
	PropPublishMandatory  = "amqp.property.publish.mandatory"

	// Consuming.
	PropConsumerExchangeName                 = "amqp.property.consume.exchange.name"
	PropConsumerExchangeType                 = "amqp.property.consume.exchange.type"
	PropConsumerExchangeDurable              = "amqp.property.consume.exchange.durable"
	PropConsumerQueueName                    = "amqp.property.consume.queue.name"
	PropConsumerQueueDurable                 = "amqp.property.consume.queue.durable"
	PropConsumerQueueAutoDelete              = "amqp.property.consume.queue.autoDelete"
	PropConsumerQueueExclusive               = "amqp.property.consume.queue.exclusive"
	PropConsumerQueueNoWait                  = "amqp.property.consume.queue.noWait"
	PropConsumerQueueArgs                    = "amqp.property.consume.queue.args"
	PropConsumerConsumeName                  = "amqp.property.consume.consumerName"
	PropConsumerConsumeAutoAck               = "amqp.property.consume.autoAck"
	PropConsumerConsumeExclusive             = "amqp.property.consume.exclusive"
	PropConsumerConsumeNoLocal               = "amqp.property.consume.noLocal"
	PropConsumerConsumeNoWait                = "amqp.property.consume.noWait"
	PropConsumerConsumeArgs                  = "amqp.property.consume.args"
	PropConsumerConsumeInternalRetryInterval = "amqp.property.consume.retryInterval"
	PropConsumerConsumeInternalRetryMax      = "amqp.property.consume.retryMax"
)

// Config holds the configuration for an Exchange-Queue-Bind setup.
type Config struct {
	Exchange ExchangeDefinition
	Queue    *QueueDefinition
	Bind     BindDefinition
}

// ExchangeDefinition holds the configuration for an AMQP exchange.
type ExchangeDefinition struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]any
}

// QueueDefinition holds the configuration for an AMQP queue.
type QueueDefinition struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       map[string]any
}

// BindDefinition holds the configuration for an AMQP queue binding.
type BindDefinition struct {
	RoutingKey string
	NoWait     bool
	Args       map[string]any
}

// NewDurableExchange creates a new durable exchange definition with the given name and kind.
func NewDurableExchange(name, kind string) ExchangeDefinition {
	return ExchangeDefinition{
		Name:    name,
		Kind:    kind,
		Durable: true,
	}
}

// NewDurableQueue creates a new durable queue definition with the given name.
func NewDurableQueue(name string) QueueDefinition {
	return QueueDefinition{
		Name:    name,
		Durable: true,
	}
}

// NewRoutingKeyBind creates a new bind definition with the given routing key.
func NewRoutingKeyBind(routingKey string) BindDefinition {
	return BindDefinition{
		RoutingKey: routingKey,
	}
}
