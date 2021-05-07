package broadcast

import (
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

type Broadcaster interface {
	Publish(topic string, data interface{}) error
	Subscribe(topic string, cb nats.Handler) (func() error, error)
}

var (
	natsEConn *nats.EncodedConn
	natsConn  *nats.Conn
)

func init() {
	viper.SetDefault("nats.url", nats.DefaultURL)
}

func Driver() (Broadcaster, error) {
	if natsEConn == nil {
		err := setupNats()
		if err != nil {
			return nil, err
		}
	}

	return newNatsBroadcaster(natsEConn), nil
}

func setupNats() error {
	var err error
	natsConn, err = nats.Connect(viper.GetString("nats.url"))
	if err != nil {
		return err
	}

	natsEConn, err = nats.NewEncodedConn(natsConn, nats.JSON_ENCODER)
	if err != nil {
		return err
	}

	return nil
}

func newNatsBroadcaster(nce *nats.EncodedConn) *NatsBroadcaster {
	return &NatsBroadcaster{nce}
}

type NatsBroadcaster struct {
	conn *nats.EncodedConn
}

func (nb *NatsBroadcaster) Publish(topic string, data interface{}) error {
	return nb.conn.Publish(topic, data)
}

func (nb *NatsBroadcaster) Subscribe(topic string, cb nats.Handler) (func() error, error) {
	sub, err := nb.conn.Subscribe(topic, cb)

	close := func() error {
		return sub.Unsubscribe()
	}

	return close, err
}
