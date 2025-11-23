package common

import (
	"github.com/oo-developer/mmq/pkg"
)

type BrokerClient interface {
	Id() string
	User() User
	MessageChan() <-chan *api.Message
}

type Topic struct {
	Topic      string
	Persistent bool
	Retained   bool
}

type BrokerService interface {
	Service
	RegisterClient(clientID string, user User) BrokerClient
	UnregisterClient(clientID string)
	Client(clientId string) BrokerClient
	AllClients() []BrokerClient
	AllTopics() []*Topic
	Subscribe(clientID, topic string) (string, error)
	Unsubscribe(clientID, topic string, subscriptionId string) error
	Publish(properties api.MessageProperty, topic string, payload []byte, publisherID string)
}
