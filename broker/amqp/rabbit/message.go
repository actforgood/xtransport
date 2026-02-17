package rabbit

import (
	"github.com/actforgood/xtransport/broker"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConvertToMessage converts an AMQP message to internal Message format.
func ConvertToMessage(amqpMsg amqp.Delivery) broker.Message {
	return broker.Message{
		Body: amqpMsg.Body,
		Props: broker.Props{
			PropMsgHeaders:         amqpMsg.Headers,
			PropMsgContentType:     amqpMsg.ContentType,
			PropMsgContentEncoding: amqpMsg.ContentEncoding,
			PropMsgDeliveryMode:    amqpMsg.DeliveryMode,
			PropMsgPriority:        amqpMsg.Priority,
			PropMsgCorrelationID:   amqpMsg.CorrelationId,
			PropMsgReplyTo:         amqpMsg.ReplyTo,
			PropMsgExpiration:      amqpMsg.Expiration,
			PropMsgMessageID:       amqpMsg.MessageId,
			PropMsgTimestamp:       amqpMsg.Timestamp,
			PropMsgType:            amqpMsg.Type,
			PropMsgUserID:          amqpMsg.UserId,
			PropMsgAppID:           amqpMsg.AppId,
			PropMsgConsumerTag:     amqpMsg.ConsumerTag,
			PropMsgMessageCount:    amqpMsg.MessageCount,
			PropMsgDeliveryTag:     amqpMsg.DeliveryTag,
			PropMsgRedelivered:     amqpMsg.Redelivered,
			PropMsgExchange:        amqpMsg.Exchange,
			PropMsgRoutingKey:      amqpMsg.RoutingKey,
		},
	}
}

// IsRetried checks if the given AMQP message has been retried
// (based on the presence of the "x-death" header and its count).
func IsRetried(amqpMsg amqp.Delivery) bool {
	return RetryCount(amqpMsg) > 0
}

// RetryCount retrieves the retry count from the "x-death" header of the given AMQP message.
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

// GetOriginQueue retrieves the origin queue from the "x-first-death-queue" header of the given AMQP message.
func GetOriginQueue(amqpMsg amqp.Delivery) string {
	if xDeathFirstQueue, found := amqpMsg.Headers["x-first-death-queue"]; found {
		if queueName, ok := xDeathFirstQueue.(string); ok {
			return queueName
		}
	}

	return ""
}
