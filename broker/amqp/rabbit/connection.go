package rabbit

import (
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/actforgood/xconf"
	"github.com/actforgood/xerr"
	"github.com/actforgood/xver"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ConnectionFactory is an interface for managing AMQP connections and channels.
type ConnectionFactory interface {
	// Conn returns the underlying AMQP connection.
	Conn() *amqp.Connection
	// Channel returns an AMQP channel with the given internal ID.
	Channel(string) (*amqp.Channel, error)
	// Close closes the underlying AMQP connection and all channels.
	Close() error
}

type defaultConnectionFactory struct {
	cfg        xconf.Config
	cfgDSNKey  string
	dsn        string
	conn       *amqp.Connection
	connMu     sync.RWMutex
	channels   map[string]*amqp.Channel
	channelsMu sync.RWMutex
}

// NewXConfConnectionFactory creates a new ConnectionFactory using the provided xconf.Config
// and the key for the AMQP DSN.
func NewXConfConnectionFactory(config xconf.Config, amqpDSNKey string) (ConnectionFactory, error) {
	dsn := config.Get(amqpDSNKey, "").(string)
	conn, err := initAMQPConn(dsn)
	if err != nil {
		return nil, err
	}

	connFac := &defaultConnectionFactory{
		cfg:       config,
		cfgDSNKey: amqpDSNKey,
		conn:      conn,
		channels:  make(map[string]*amqp.Channel),
	}

	if defConfig, ok := config.(*xconf.DefaultConfig); ok {
		defConfig.RegisterObserver(connFac.onConfigChange)
	}

	connFac.reconnectOnUnexpectedClose()

	return connFac, nil
}

// NewDefaultConnectionFactory creates a new ConnectionFactory using the provided AMQP DSN.
func NewDefaultConnectionFactory(amqpDSN string) (ConnectionFactory, error) {
	conn, err := initAMQPConn(amqpDSN)
	if err != nil {
		return nil, err
	}

	connFac := &defaultConnectionFactory{
		dsn:  amqpDSN,
		conn: conn,
	}

	connFac.reconnectOnUnexpectedClose()

	return connFac, nil
}

// Conn...see [ConnectionFactory.Conn].
func (connFac *defaultConnectionFactory) Conn() *amqp.Connection {
	connFac.connMu.RLock()
	defer connFac.connMu.RUnlock()

	return connFac.conn
}

// Channel...see [ConnectionFactory.Channel].
func (connFac *defaultConnectionFactory) Channel(id string) (*amqp.Channel, error) {
	connFac.channelsMu.RLock()
	ch, found := connFac.channels[id]
	connFac.channelsMu.RUnlock()

	if !found || ch.IsClosed() {
		connFac.channelsMu.Lock()
		defer connFac.channelsMu.Unlock()
		newCh, err := connFac.Conn().Channel()
		if err != nil {
			return nil, xerr.Wrap(err, "could not initialize channel")
		}
		connFac.channels[id] = newCh

		return newCh, nil
	}

	return ch, nil
}

// Close...see [ConnectionFactory.Close].
func (connFac *defaultConnectionFactory) Close() error {
	if !connFac.Conn().IsClosed() {
		if err := connFac.Conn().Close(); err != nil {
			return xerr.Wrap(err, "could not close amqp connection")
		}
	}

	return nil
}

// onConfigChange is a callback to be registered to xconf.DefaultConfig which knows to reload configuration.
// In case connection config is changed, it is reinitialized with the new config.
func (connFac *defaultConnectionFactory) onConfigChange(config xconf.Config, changedKeys ...string) {
	if slices.Contains(changedKeys, connFac.cfgDSNKey) {
		dsn := connFac.dsn
		if connFac.cfg != nil {
			dsn = connFac.cfg.Get(connFac.cfgDSNKey, "").(string)
		}
		newConn, err := initAMQPConn(dsn)
		if err == nil {
			connFac.connMu.Lock()
			oldConn := connFac.conn
			connFac.conn = newConn
			for id := range connFac.channels {
				if ch, errCh := connFac.conn.Channel(); errCh == nil {
					connFac.channels[id] = ch
				}
			}
			connFac.reconnectOnUnexpectedClose()
			connFac.connMu.Unlock()
			_ = oldConn.Close()
		}
	}
}

func (connFac *defaultConnectionFactory) reconnectOnUnexpectedClose() {
	errsChan := make(chan *amqp.Error)
	connFac.conn.NotifyClose(errsChan)
	go func() {
		closeErr := <-errsChan
		if closeErr != nil {
			dsn := connFac.dsn
			if connFac.cfg != nil {
				dsn = connFac.cfg.Get(connFac.cfgDSNKey, "").(string)
			}
			newConn, err := initAMQPConn(dsn)
			if err == nil {
				connFac.connMu.Lock()
				oldConn := connFac.conn
				connFac.conn = newConn
				connFac.channelsMu.Lock()
				for id := range connFac.channels {
					if ch, errCh := connFac.conn.Channel(); errCh == nil {
						connFac.channels[id] = ch
					}
				}
				connFac.channelsMu.Unlock()
				connFac.reconnectOnUnexpectedClose()
				connFac.connMu.Unlock()
				_ = oldConn.Close()
			}
		}
	}()
}

func initAMQPConn(dsn string) (*amqp.Connection, error) {
	if dsn == "" {
		return nil, xerr.New(`amqp dsn is required`)
	}

	// conn, err := amqp.Dial(dsn)
	props := amqp.NewConnectionProperties()
	hostname, _ := os.Hostname()
	props.SetClientConnectionName(xver.Information().App.Name + "/" + xver.Information().App.Version + "@" + hostname)
	conn, err := amqp.DialConfig(dsn, amqp.Config{
		Locale:     "en_US",
		Properties: props,
	})
	if err != nil {
		if uri, parseErr := amqp.ParseURI(dsn); parseErr == nil {
			if uri.Password != "" {
				if len(uri.Password) > 2 {
					uri.Password = fmt.Sprintf("%c***%c", uri.Password[0], uri.Password[len(uri.Password)-1])
				} else {
					uri.Password = "***"
				}
			}

			return nil, xerr.Wrapf(err, `could not open amqp connection (%s)`, uri.String())
		}

		return nil, xerr.Wrapf(err, `could not open amqp connection`)
	}

	return conn, nil
}
