package rabbit

import (
	"context"
	"errors"

	"github.com/actforgood/xerr"
	"github.com/actforgood/xrand"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/actforgood/xtransport/broker"
)

type publisher struct {
	config    Config
	connFac   ConnectionFactory
	channelID string
}

func NewPublisher(connFac ConnectionFactory, Config Config) (broker.Publisher, error) {
	pub := &publisher{config: Config, connFac: connFac, channelID: xrand.String(6)}
	if err := pub.initialize(); err != nil {
		return nil, err
	}

	return pub, nil
}

func (p *publisher) initialize() error {
	ch, err := p.connFac.Channel(p.channelID)
	if err != nil {
		return xerr.Wrap(err, "could not initialize channel")
	}

	if p.config.Exchange.Name != "" {
		if err = ch.ExchangeDeclare(
			p.config.Exchange.Name,
			p.config.Exchange.Kind,
			p.config.Exchange.Durable,
			p.config.Exchange.AutoDelete,
			p.config.Exchange.Internal,
			p.config.Exchange.NoWait,
			p.config.Exchange.Args,
		); err != nil {
			return xerr.Wrap(err, "could not initialize exchange")
		}
	}

	if p.config.Queue != nil {
		q, err := ch.QueueDeclare(
			p.config.Queue.Name,
			p.config.Queue.Durable,
			p.config.Queue.AutoDelete,
			p.config.Queue.Exclusive,
			p.config.Queue.NoWait,
			p.config.Queue.Args,
		)
		if err != nil {
			return xerr.Wrap(err, "could not initialize queue")
		}

		if err := ch.QueueBind(
			q.Name,
			p.config.Bind.RoutingKey,
			p.config.Exchange.Name,
			p.config.Bind.NoWait,
			p.config.Bind.Args,
		); err != nil {
			return xerr.Wrap(err, "could not bind queue")
		}
	}

	return nil
}

func (p *publisher) Publish(ctx context.Context, msg broker.Message) error {
	pubMsg := amqp.Publishing{
		ContentType:     msg.Props.GetString(PropMsgContentType),
		ContentEncoding: msg.Props.GetString(PropMsgContentEncoding),
		DeliveryMode:    uint8(msg.Props.GetInt(PropMsgDeliveryMode)),
		Priority:        uint8(msg.Props.GetInt(PropMsgPriority)),
		CorrelationId:   msg.Props.GetString(PropMsgCorrelationId),
		ReplyTo:         msg.Props.GetString(PropMsgReplyTo),
		Expiration:      msg.Props.GetString(PropMsgExpiration),
		MessageId:       msg.Props.GetString(PropMsgMessageId),
		Timestamp:       msg.Props.GetTime(PropMsgTimestamp),
		Type:            msg.Props.GetString(PropMsgType),
		UserId:          msg.Props.GetString(PropMsgUserId),
		AppId:           msg.Props.GetString(PropMsgAppId),
		Body:            msg.Body,
	}

	ch, err := p.connFac.Channel(p.channelID)
	if err != nil {
		return xerr.Wrap(err, "could not publish message")
	}
	if err := ch.PublishWithContext(
		ctx,
		p.config.Exchange.Name,
		msg.Props.GetString(PropPublishRoutingKey),
		msg.Props.GetBool(PropPublishMandatory),
		msg.Props.GetBool(PropPublishImmediate),
		pubMsg,
	); err != nil {
		return xerr.Wrap(err, "could not publish message")
	}

	return nil
}

func (p *publisher) Close() error {
	ch, _ := p.connFac.Channel(p.channelID)
	if ch != nil {
		if err := ch.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			return xerr.Wrap(err, "could not close amqp channel")
		}
	}

	return nil
}
