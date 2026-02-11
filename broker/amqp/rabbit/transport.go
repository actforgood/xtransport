// Package rabbit provides a AMQP transport and some utitities.
package rabbit

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/actforgood/xerr"
	"github.com/actforgood/xlog"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/actforgood/xtransport"
	"github.com/actforgood/xtransport/broker"
)

type rabbitmqTransport struct {
	connFac           ConnectionFactory
	consumers         []broker.Consumer
	shutDown          bool
	consummersStopped chan struct{}
	mu                *sync.RWMutex
}

// NewRabbitMQTransport instantiates a new RabbitMQ transport.
func NewRabbitMQTransport(
	connFac ConnectionFactory,
	consumers ...broker.Consumer,
) xtransport.Transport {
	return &rabbitmqTransport{
		connFac:           connFac,
		consumers:         consumers,
		consummersStopped: make(chan struct{}),
		mu:                new(sync.RWMutex),
	}
}

// StartAsync starts the HTTP server. It listens for new connections and messages.
func (rt *rabbitmqTransport) StartAsync(ctx context.Context, errorsChan chan<- error) {
	xlog.LoggerFromContext(ctx).Info(xlog.MsgKey, "AMQP (RabbitMQ) transport starting")

	go func() {
		var err error
		var ch *amqp.Channel
		var errorsCount = 0
		var wg sync.WaitGroup
		for _, consumer := range rt.consumers {
			errorsCount = 0
			for errorsCount < 3 {
				if err != nil {
					errorsCount++
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Duration(errorsCount) * time.Second):

				}
				ch, err = rt.connFac.Channel(consumer.Props().GetString(PropConsumerConsumeName))
				if err != nil {
					xlog.LoggerFromContext(ctx).Warn(
						xlog.MsgKey, "could not make channel",
						xlog.ErrorKey, err,
						"correlationId", xtransport.CorrelationIDFromContext(ctx),
					)
					continue
				}
				ch.Qos(50, 0, false) // TODO: make it configurable

				err = rt.setUpQueue(ctx, ch, consumer)
				if err != nil {
					continue
				}

				var argsConsume map[string]any
				if argsFromProps, ok := consumer.Props()[PropConsumerConsumeArgs].(map[string]any); ok {
					argsConsume = amqp.Table(argsFromProps)
				}

				deliveryChan, err := ch.Consume(
					consumer.Props().GetString(PropConsumerQueueName),
					consumer.Props().GetString(PropConsumerConsumeName),
					consumer.Props().GetBool(PropConsumerConsumeAutoAck),
					consumer.Props().GetBool(PropConsumerConsumeExclusive),
					consumer.Props().GetBool(PropConsumerConsumeNoLocal),
					consumer.Props().GetBool(PropConsumerConsumeNoWait),
					argsConsume,
				)
				if err != nil {
					xlog.LoggerFromContext(ctx).Warn(
						xlog.MsgKey, "could not get messages chan",
						xlog.ErrorKey, err,
						"correlationId", xtransport.CorrelationIDFromContext(ctx),
					)
					continue
				} else {
					xlog.LoggerFromContext(ctx).Info(
						xlog.MsgKey, "consumer started consuming messages",
						"correlationId", xtransport.CorrelationIDFromContext(ctx),
						"consumer", consumer.Props().GetString(PropConsumerConsumeName),
						"queue", consumer.Props().GetString(PropConsumerQueueName),
					)
				}
				wg.Add(1)
				go rt.consumeMessages(ctx, deliveryChan, &wg, consumer)
				break
			}
		}
		wg.Wait()
		if rt.isShutDown() {
			rt.consummersStopped <- struct{}{}
			err = nil
		}
		if err != nil {
			errorsChan <- err
		}
	}()
}

// Shutdown closes the connection to RabbitMQ server.
func (rt *rabbitmqTransport) Shutdown(ctx context.Context) error {
	rt.mu.Lock()
	rt.shutDown = true
	rt.mu.Unlock()
	for _, consumer := range rt.consumers {
		ch, err := rt.connFac.Channel(consumer.Props().GetString(PropConsumerConsumeName))
		if err != nil || ch.IsClosed() {
			continue
		}
		if err := ch.Cancel(consumer.Props().GetString(PropConsumerConsumeName), false); err != nil {
			xlog.LoggerFromContext(ctx).Warn(
				xlog.MsgKey, "could not cancel consumer",
				xlog.ErrorKey, err,
				"correlationId", xtransport.CorrelationIDFromContext(ctx),
				"consumer", consumer.Props().GetString(PropConsumerConsumeName),
				"queue", consumer.Props().GetString(PropConsumerQueueName),
			)
		} else {
			xlog.LoggerFromContext(ctx).Info(
				xlog.MsgKey, "consumer cancelled",
				"correlationId", xtransport.CorrelationIDFromContext(ctx),
				"consumer", consumer.Props().GetString(PropConsumerConsumeName),
				"queue", consumer.Props().GetString(PropConsumerQueueName),
			)
		}
	}

	var mErr xerr.MultiError
	select {
	case <-rt.consummersStopped:
		// all good, consumers stopped
		// close the channel
		for _, consumer := range rt.consumers {
			ch, err := rt.connFac.Channel(consumer.Props().GetString(PropConsumerConsumeName))
			if err == nil {
				if err := ch.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
					mErr.Add(xerr.Wrapf(err,
						"could not close RabbitMQ channel for consumer %s",
						consumer.Props().GetString(PropConsumerConsumeName),
					))
				}
			}
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return mErr.ErrOrNil()
}

func (rt *rabbitmqTransport) isShutDown() bool {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	return rt.shutDown
}

func (rt *rabbitmqTransport) setUpQueue(ctx context.Context, ch *amqp.Channel, consumer broker.Consumer) error {
	var argsQueue amqp.Table
	if argsFromProps, ok := consumer.Props()[PropConsumerQueueArgs].(map[string]any); ok {
		argsQueue = amqp.Table(argsFromProps)
	}
	if _, err := ch.QueueDeclare(
		consumer.Props().GetString(PropConsumerQueueName),
		consumer.Props().GetBool(PropConsumerQueueDurable),
		consumer.Props().GetBool(PropConsumerQueueAutoDelete),
		consumer.Props().GetBool(PropConsumerQueueExclusive),
		consumer.Props().GetBool(PropConsumerQueueNoWait),
		argsQueue,
	); err != nil {
		xlog.LoggerFromContext(ctx).Warn(
			xlog.MsgKey, "could not initialize queue",
			"queue", consumer.Props().GetString(PropConsumerQueueName),
			xlog.ErrorKey, err,
			"correlationId", xtransport.CorrelationIDFromContext(ctx),
		)
		return err
	}

	if consumer.Props().GetString(PropConsumerExchangeName) != "" {
		if err := ch.ExchangeDeclare(
			consumer.Props().GetString(PropConsumerExchangeName),
			consumer.Props().GetString(PropConsumerExchangeType),
			consumer.Props().GetBool(PropConsumerExchangeDurable),
			false,
			false,
			false,
			nil,
		); err != nil {
			xlog.LoggerFromContext(ctx).Warn(
				xlog.MsgKey, "could not initialize exchange",
				"exchange", consumer.Props().GetString(PropConsumerExchangeName),
				"queue", consumer.Props().GetString(PropConsumerQueueName),
				xlog.ErrorKey, err,
				"correlationId", xtransport.CorrelationIDFromContext(ctx),
			)
			return err
		}

		if err := ch.QueueBind(
			consumer.Props().GetString(PropConsumerQueueName),
			consumer.Props().GetString(PropMsgRoutingKey),
			consumer.Props().GetString(PropConsumerExchangeName),
			false,
			nil,
		); err != nil {
			xlog.LoggerFromContext(ctx).Warn(
				xlog.MsgKey, "could not bind queue",
				"exchange", consumer.Props().GetString(PropConsumerExchangeName),
				"queue", consumer.Props().GetString(PropConsumerQueueName),
				xlog.ErrorKey, err,
				"correlationId", xtransport.CorrelationIDFromContext(ctx),
			)
			return err
		}

		// dlx setup
		dlxExchangeName := argsQueue["x-dead-letter-exchange"]
		if dlxExchangeNameStr, ok := dlxExchangeName.(string); ok && dlxExchangeNameStr != "" {
			if err := ch.ExchangeDeclare(
				dlxExchangeNameStr,
				consumer.Props().GetString(PropConsumerExchangeType),
				consumer.Props().GetBool(PropConsumerExchangeDurable),
				false,
				false,
				false,
				nil,
			); err != nil {
				xlog.LoggerFromContext(ctx).Warn(
					xlog.MsgKey, "could not initialize DLX exchange",
					"exchange", dlxExchangeNameStr,
					"queue", consumer.Props().GetString(PropConsumerQueueName),
					xlog.ErrorKey, err,
					"correlationId", xtransport.CorrelationIDFromContext(ctx),
				)
				return err
			}

			retryInterval := int(consumer.Props().GetDuration(PropConsumerConsumeInternalRetryInterval) / time.Millisecond)
			retryQueueName := dlxExchangeNameStr + "_" + consumer.Props().GetString(PropMsgRoutingKey) + "_retry_" + strconv.Itoa(retryInterval)
			if _, err := ch.QueueDeclare(
				retryQueueName,
				consumer.Props().GetBool(PropConsumerQueueDurable),
				consumer.Props().GetBool(PropConsumerQueueAutoDelete),
				consumer.Props().GetBool(PropConsumerQueueExclusive),
				consumer.Props().GetBool(PropConsumerQueueNoWait),
				amqp.Table{
					"x-dead-letter-exchange":    consumer.Props().GetString(PropConsumerExchangeName),
					"x-dead-letter-routing-key": consumer.Props().GetString(PropMsgRoutingKey),
					"x-message-ttl":             retryInterval,
				},
			); err != nil {
				xlog.LoggerFromContext(ctx).Warn(
					xlog.MsgKey, "could not initialize retry queue",
					"queue", retryQueueName,
					xlog.ErrorKey, err,
					"correlationId", xtransport.CorrelationIDFromContext(ctx),
				)
				return err
			}

			dlxRoutingKey := ""
			if dlxRoutingKeyStr, ok := argsQueue["x-dead-letter-routing-key"].(string); !ok {
				dlxRoutingKey = consumer.Props().GetString(PropMsgRoutingKey)
			} else {
				dlxRoutingKey = dlxRoutingKeyStr
			}
			if err := ch.QueueBind(
				retryQueueName,
				dlxRoutingKey,
				dlxExchangeNameStr,
				false,
				nil,
			); err != nil {
				xlog.LoggerFromContext(ctx).Warn(
					xlog.MsgKey, "could not bind retry queue",
					"exchange", dlxExchangeNameStr,
					"queue", retryQueueName,
					xlog.ErrorKey, err,
					"correlationId", xtransport.CorrelationIDFromContext(ctx),
				)
				return err
			}
		}
	}

	return nil
}

func (rt *rabbitmqTransport) consumeMessages(ctx context.Context, msgChan <-chan amqp.Delivery, wg *sync.WaitGroup, consumer broker.Consumer) {
	skippedCount := 0
	var ackResult byte
	for msg := range msgChan {
		var newCtx context.Context
		if msg.CorrelationId != "" {
			newCtx = xtransport.ContextWithCorrelationID(ctx, msg.CorrelationId)
		} else {
			newCtx = ctx
		}

		if rt.isShutDown() {
			// transport is shutting down, skip processing the message and requeue it
			if err := msg.Nack(false, true); err != nil {
				xlog.LoggerFromContext(newCtx).Error(
					xlog.MsgKey, "could not NACK-Requeue message at shutdown",
					xlog.ErrorKey, xerr.Wrap(err, ""),
					"correlationId", xtransport.CorrelationIDFromContext(newCtx),
				)
			} else {
				skippedCount++
			}
			continue
		}

		if IsRetried(msg) && GetOriginQueue(msg) != consumer.Props().GetString(PropConsumerQueueName) {
			// skip retried messages that did not originate from this consumer's queue
			// multiple consumers might share the same DLX and routing key.
			ackResult = broker.ConsumeResultAck
		} else {
			ackResult = consumer.Consume(newCtx, ConvertToMessage(msg))
			if IsRetried(msg) &&
				consumer.Props().GetInt(PropConsumerConsumeInternalRetryMax) > 0 &&
				RetryCount(msg) >= consumer.Props().GetInt(PropConsumerConsumeInternalRetryMax) &&
				ackResult == broker.ConsumeResultNack {
				xlog.LoggerFromContext(newCtx).Warn(
					xlog.MsgKey, "message exceeded max retry count, acknowledging it",
					"maxRetry", consumer.Props().GetInt(PropConsumerConsumeInternalRetryMax),
					"correlationId", xtransport.CorrelationIDFromContext(newCtx),
					"consumer", consumer.Props().GetString(PropConsumerConsumeName),
					"queue", consumer.Props().GetString(PropConsumerQueueName),
				)
				ackResult = broker.ConsumeResultAck
			}
		}
		switch ackResult {
		case broker.ConsumeResultAck:
			if err := msg.Ack(false); err != nil {
				xlog.LoggerFromContext(newCtx).Error(
					xlog.MsgKey, "could not ACK message",
					xlog.ErrorKey, xerr.Wrap(err, ""),
					"correlationId", xtransport.CorrelationIDFromContext(newCtx),
				)
			}
		case broker.ConsumeResultNack:
			if err := msg.Nack(false, false); err != nil {
				xlog.LoggerFromContext(newCtx).Error(
					xlog.MsgKey, "could not NACK message",
					xlog.ErrorKey, xerr.Wrap(err, ""),
					"correlationId", xtransport.CorrelationIDFromContext(newCtx),
				)
			}
		case broker.ConsumeResultNackRequeue:
			if err := msg.Nack(false, true); err != nil {
				xlog.LoggerFromContext(newCtx).Error(
					xlog.MsgKey, "could not NACK-Requeue message",
					xlog.ErrorKey, xerr.Wrap(err, ""),
					"correlationId", xtransport.CorrelationIDFromContext(newCtx),
				)
			}
		}
	}

	if skippedCount > 0 {
		xlog.LoggerFromContext(ctx).Info(
			xlog.MsgKey, "skipped processing messages due to shutdown",
			"skippedCount", skippedCount,
			"correlationId", xtransport.CorrelationIDFromContext(ctx),
			"consumer", consumer.Props().GetString(PropConsumerConsumeName),
			"queue", consumer.Props().GetString(PropConsumerQueueName),
		)
	}

	wg.Done()
}
