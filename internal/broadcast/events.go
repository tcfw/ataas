package broadcast

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"

	nats "github.com/nats-io/nats.go"
)

const (
	defaultNatsHost      = "nats:8222"
	defaultChannelPrefix = "ataas."
)

//TODO(tcfw) support https://github.com/cloudevents/spec/blob/v0.3/spec.md

//Event is the basic structure all events should include
type Event struct {
	Channel  string
	Type     string
	Version  string
	Source   string
	ID       string
	Time     time.Time
	Metadata map[string]interface{}
}

//EventInterface provides required funcs to identify common structured events
type EventInterface interface {
	GetType() string
	SetSource(string)
	SetTime(time.Time)
	SetID()
	SetMetadata(map[string]interface{})
	GetMetadata() map[string]interface{}
	GetChannel() string
	SetChannel(string)
}

//GetType returns event type
func (e *Event) GetType() string {
	return e.Type
}

//SetSource applies source
func (e *Event) SetSource(source string) {
	e.Source = source
}

//SetTime sets event time from timestamp
func (e *Event) SetTime(eventTime time.Time) {
	e.Time = eventTime
}

//SetID generates a new UUID for the event id
func (e *Event) SetID() {
	e.ID = uuid.New().String()
}

//SetMetadata overrides the existing metadata string map
func (e *Event) SetMetadata(md map[string]interface{}) {
	e.Metadata = md
}

//GetMetadata returns the current metadata
func (e *Event) GetMetadata() map[string]interface{} {
	return e.Metadata
}

//GetChannel returns the event channel
func (e *Event) GetChannel() string {
	return e.Channel
}

//SetChannel overwrites the event channel
func (e *Event) SetChannel(channel string) {
	e.Channel = channel
}

//AuthenticateEvent publishes to events.auth.auth
type AuthenticateEvent struct {
	*Event
	AuthType string `json:"authType"`
	Success  bool   `json:"success"`
	User     string `json:"user"`
	IP       string `json:"ip"`
	Err      string `json:"error,omitempty"`
}

//ListenForBroadcast creates a new NATS connection and watches for internal events
//based on a channel & type using ListenForBroadcastOnNC
func ListenForBroadcast(serviceName string, eventType string, channel string) (<-chan []byte, func(), error) {
	nc, err := nats.Connect(viper.GetString("nats.url"))
	if err != nil {
		log.Printf("Failed to connect to nats: %s", err)
		return nil, func() {}, err
	}

	out, subCloseFunc, err := ListenForBroadcastOnNC(nc, serviceName, eventType, channel)
	if err != nil {
		return out, subCloseFunc, err
	}

	closeFunc := func() {
		subCloseFunc()
		nc.Close()
	}

	return out, closeFunc, err
}

//MultiListenForBroadcast listens for multiple events
func MultiListenForBroadcast(serviceName string, eventTypes ...string) (<-chan []byte, func(), error) {
	nc, err := nats.Connect(viper.GetString("nats.url"))
	if err != nil {
		log.Printf("Failed to connect to nats: %s", err)
		return nil, func() {}, err
	}

	outs := []<-chan []byte{}
	closeFns := []func(){}

	for _, eventType := range eventTypes {
		out, subCloseFunc, err := ListenForBroadcastOnNC(nc, serviceName, eventType, "")
		if err != nil {
			return nil, nil, err
		}

		outs = append(outs, out)
		closeFns = append(closeFns, subCloseFunc)
	}

	closer := func() {
		for _, closer := range closeFns {
			closer()
		}
		nc.Close()
	}

	return Merge(outs...), closer, nil
}

//ListenForBroadcastOnNC is the same as ListenForBroadcast but on an existing nats connection
func ListenForBroadcastOnNC(nc *nats.Conn, serviceName, eventType, channel string) (chan []byte, func(), error) {
	if channel == "" {
		channel = defaultChannelPrefix + "broadcast"
	}

	inbound := make(chan *nats.Msg, 64)
	outbound := make(chan []byte, 64)
	closeCh := make(chan struct{})

	closeFunc := func() {
		close(closeCh)
	}

	var err error
	var sub *nats.Subscription
	if serviceName != "" {
		sub, err = nc.ChanQueueSubscribe(channel, serviceName, inbound)
	} else {
		sub, err = nc.ChanSubscribe(channel, inbound)
	}
	if err != nil {
		return nil, func() {}, err
	}

	go func() {
		for {
			select {
			case msg := <-inbound:
				ev := &Event{}
				json.Unmarshal(msg.Data, ev)
				if ev.Type == eventType {
					outbound <- msg.Data
				}
			case <-closeCh:
				//assume closed
				sub.Unsubscribe()
				return
			}
		}
	}()

	return outbound, closeFunc, nil
}

//BroadcastEvent attempts to connect to nats server to pub any event and saves to stream
func BroadcastEvent(ctx context.Context, event EventInterface) error {
	hostname, _ := os.Hostname()
	nc, err := nats.Connect(viper.GetString("nats.url"))
	if err != nil {
		log.Printf("Failed to broadcast event: %s", err)
		return err
	}
	defer nc.Close()

	event.SetSource(hostname)
	event.SetID()
	event.SetTime(time.Now())

	if event.GetChannel() == "" {
		event.SetChannel(defaultChannelPrefix + "broadcast")
	}

	channel := event.GetChannel()

	event = appendContextUserInfo(ctx, event)

	buf, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return nc.Publish(channel, buf)
}

//BroadcastNonStreamingEvent broadcasts an event like BroadcastEvent but uses the non-streaming engine
func BroadcastNonStreamingEvent(ctx context.Context, event EventInterface) error {
	nc, err := nats.Connect(viper.GetString("nats.url"))
	if err != nil {
		log.Printf("Failed to broadcast event: %s", err)
		return err
	}
	defer nc.Close()

	hostname, _ := os.Hostname()
	event.SetSource(hostname)
	event.SetID()
	event.SetTime(time.Now())

	event = appendContextUserInfo(ctx, event)

	buf, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return nc.Publish(event.GetType(), buf)
}
