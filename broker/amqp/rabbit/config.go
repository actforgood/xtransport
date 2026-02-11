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
	// Message, common, consumer - producer
	PropMsgContentType     = "amqp.property.msg.contentType"
	PropMsgContentEncoding = "amqp.property.msg.contentEncoding"
	PropMsgDeliveryMode    = "amqp.property.msg.deliveryMode"
	PropMsgPriority        = "amqp.property.msg.priority"
	PropMsgCorrelationId   = "amqp.property.msg.correlationId"
	PropMsgReplyTo         = "amqp.property.msg.replyTo"
	PropMsgExpiration      = "amqp.property.msg.expiration"
	PropMsgMessageId       = "amqp.property.msg.messageId"
	PropMsgTimestamp       = "amqp.property.msg.timestamp"
	PropMsgType            = "amqp.property.msg.type"
	PropMsgUserId          = "amqp.property.msg.userId"
	PropMsgAppId           = "amqp.property.msg.appId"
	PropMsgHeaders         = "amqp.property.msg.headers"

	// Message, consumer
	PropMsgConsumerTag  = "amqp.property.msg.consumerTag"
	PropMsgMessageCount = "amqp.property.msg.messageCount"
	PropMsgDeliveryTag  = "amqp.property.msg.deliveryTag"
	PropMsgRedelivered  = "amqp.property.msg.redelivered"
	PropMsgExchange     = "amqp.property.msg.exchange"
	PropMsgRoutingKey   = "amqp.property.msg.routingKey"

	// Publishing
	PropPublishRoutingKey = "amqp.property.publish.routingKey"
	PropPublishImmediate  = "amqp.property.publish.immediate"
	PropPublishMandatory  = "amqp.property.publish.mandatory"

	// Consuming
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

type Config struct {
	Exchange ExchangeDefinition
	Queue    *QueueDefinition
	Bind     BindDefinition
}

type ExchangeDefinition struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]any
}

type QueueDefinition struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       map[string]any
}

type BindDefinition struct {
	RoutingKey string
	NoWait     bool
	Args       map[string]any
}

func NewDurableExchange(name, kind string) ExchangeDefinition {
	return ExchangeDefinition{
		Name:    name,
		Kind:    kind,
		Durable: true,
	}
}

func NewDurableQueue(name string) QueueDefinition {
	return QueueDefinition{
		Name:    name,
		Durable: true,
	}
}

func NewRoutingKeyBind(routingKey string) BindDefinition {
	return BindDefinition{
		RoutingKey: routingKey,
	}
}
