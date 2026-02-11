package rabbit

import (
	"github.com/actforgood/xtransport/broker"
	amqp "github.com/rabbitmq/amqp091-go"
)

func ConvertToMessage(amqpMsg amqp.Delivery) broker.Message {
	return broker.Message{
		Body: amqpMsg.Body,
		Props: broker.Props{
			PropMsgHeaders:         amqpMsg.Headers,
			PropMsgContentType:     amqpMsg.ContentType,
			PropMsgContentEncoding: amqpMsg.ContentEncoding,
			PropMsgDeliveryMode:    amqpMsg.DeliveryMode,
			PropMsgPriority:        amqpMsg.Priority,
			PropMsgCorrelationId:   amqpMsg.CorrelationId,
			PropMsgReplyTo:         amqpMsg.ReplyTo,
			PropMsgExpiration:      amqpMsg.Expiration,
			PropMsgMessageId:       amqpMsg.MessageId,
			PropMsgTimestamp:       amqpMsg.Timestamp,
			PropMsgType:            amqpMsg.Type,
			PropMsgUserId:          amqpMsg.UserId,
			PropMsgAppId:           amqpMsg.AppId,
			PropMsgConsumerTag:     amqpMsg.ConsumerTag,
			PropMsgMessageCount:    amqpMsg.MessageCount,
			PropMsgDeliveryTag:     amqpMsg.DeliveryTag,
			PropMsgRedelivered:     amqpMsg.Redelivered,
			PropMsgExchange:        amqpMsg.Exchange,
			PropMsgRoutingKey:      amqpMsg.RoutingKey,
		},
	}
}

func IsRetried(amqpMsg amqp.Delivery) bool {
	return RetryCount(amqpMsg) > 0
}

func RetryCount(amqpMsg amqp.Delivery) int {
	if xDeath, foundXDeath := amqpMsg.Headers["x-death"]; foundXDeath {
		if xDeathEntries, ok := xDeath.([]any); ok && len(xDeathEntries) > 0 {
			if lastXDeathEntry, foundLastXDeathEntry := xDeathEntries[0].(amqp.Table); foundLastXDeathEntry {
				if count, foundCount := lastXDeathEntry["count"]; foundCount {
					if retryCount, ok := count.(int64); ok {
						return int(retryCount)
					}
					if retryCount, ok := count.(int); ok {
						return retryCount
					}
				}
			}
		}
	}
	return 0
}

func GetOriginQueue(amqpMsg amqp.Delivery) string {
	if xDeathFirstQueue, found := amqpMsg.Headers["x-first-death-queue"]; found {
		if queueName, ok := xDeathFirstQueue.(string); ok {
			return queueName
		}
	}
	return ""
}
