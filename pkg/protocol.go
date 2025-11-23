package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type MessageType byte
type MessageProperty byte

const (
	TypeMessage MessageType = iota
	TypeMessageAck
	TypeConnect
	TypeConnectAck
	TypeAuthenticate
	TypeAuthenticateAck
	TypeSessionKey
	TypeSessionKeyAck
	TypePublish
	TypePublishAck
	TypeSubscribe
	TypeSubscribeAck
	TypeUnsubscribe
	TypeUnsubscribeAck
	TypePing
	TypePong
	TypeCliCommand
	TypeCliCommandAck
	TypeDisconnect
)

const (
	Retained   MessageProperty = 1 << 0
	Persistent MessageProperty = 1 << 1
)

var (
	MaxTopicLength   = 2048
	MaxPayloadLength = 10485760
)

const (
	MaxClientIdLength       = 40
	MaxSubscriptionIdLength = 40
)

type Message struct {
	Type           MessageType
	Properties     MessageProperty
	Topic          string
	Payload        []byte
	ClientId       string
	SubscriptionId string
}

func (m *Message) IsRetained() bool {
	return m.Properties&Retained != 0
}

func (m *Message) IsPersistent() bool {
	return m.Properties&Persistent != 0
}

func (m *Message) Send(w io.Writer, cypher Cipher) error {
	buffer := bytes.Buffer{}
	err := m.encode(&buffer)
	if err != nil {
		return err
	}
	payloadData := buffer.Bytes()
	encryptedData, err := cypher.Encrypt(payloadData)
	if err != nil {
		return fmt.Errorf("encrypt failed: %v", err)
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(encryptedData))); err != nil {
		return fmt.Errorf("failed to write data length: %w", err)
	}
	if _, err := w.Write(encryptedData); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	return nil
}

func Receive(r io.Reader, cypher Cipher) (*Message, error) {
	var dataLen uint16
	if err := binary.Read(r, binary.BigEndian, &dataLen); err != nil {
		return nil, fmt.Errorf("failed to read topic length: %w", err)
	}
	data := make([]byte, dataLen)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, fmt.Errorf("failed to read topic: %w", err)
	}
	decryptedData, err := cypher.Decrypt(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	buffer := bytes.NewBuffer(decryptedData)
	msg, err := decode(buffer)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (m *Message) encode(w io.Writer) error {
	// Message format: [Type:1][Properties:1][TopicLen:2][Topic:n][PayloadLen:4][Payload:n][ClientIDLen:2][ClientId:n][SubscriptionIdLen:2][SubscriptionId:n]
	if err := binary.Write(w, binary.BigEndian, m.Type); err != nil {
		return fmt.Errorf("failed to write type: %w", err)
	}
	if err := binary.Write(w, binary.BigEndian, m.Properties); err != nil {
		return fmt.Errorf("failed to write properties: %w", err)
	}

	topicBytes := []byte(m.Topic)
	if len(topicBytes) > MaxTopicLength {
		return fmt.Errorf("topic too long (%d > %d)", len(topicBytes), MaxTopicLength)
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(topicBytes))); err != nil {
		return fmt.Errorf("failed to write topic length: %w", err)
	}
	if _, err := w.Write(topicBytes); err != nil {
		return fmt.Errorf("failed to write topic: %w", err)
	}

	if len(m.Payload) > MaxPayloadLength {
		return fmt.Errorf("payload too long (%d > %d)", len(m.Payload), MaxPayloadLength)
	}
	if err := binary.Write(w, binary.BigEndian, uint32(len(m.Payload))); err != nil {
		return fmt.Errorf("failed to write payload length: %w", err)
	}
	if _, err := w.Write(m.Payload); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}

	clientIDBytes := []byte(m.ClientId)
	if len(clientIDBytes) > MaxClientIdLength {
		return fmt.Errorf("client id too long (%d > %d)", len(clientIDBytes), MaxClientIdLength)
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(clientIDBytes))); err != nil {
		return fmt.Errorf("failed to write client ID length: %w", err)
	}
	if _, err := w.Write(clientIDBytes); err != nil {
		return fmt.Errorf("failed to write client ID: %w", err)
	}

	subscriptionIdBytes := []byte(m.SubscriptionId)
	if len(subscriptionIdBytes) > MaxSubscriptionIdLength {
		return fmt.Errorf("subscription id too long (%d > %d)", len(subscriptionIdBytes), MaxSubscriptionIdLength)
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(subscriptionIdBytes))); err != nil {
		return fmt.Errorf("failed to write subscription ID length: %w", err)
	}
	if _, err := w.Write(subscriptionIdBytes); err != nil {
		return fmt.Errorf("failed to write subscription ID: %w", err)
	}

	return nil
}

func decode(r io.Reader) (*Message, error) {
	msg := &Message{}

	if err := binary.Read(r, binary.BigEndian, &msg.Type); err != nil {
		return nil, fmt.Errorf("failed to read type: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &msg.Properties); err != nil {
		return nil, fmt.Errorf("failed to read properties: %w", err)
	}
	var topicLen uint16
	if err := binary.Read(r, binary.BigEndian, &topicLen); err != nil {
		return nil, fmt.Errorf("failed to read topic length: %w", err)
	}
	if topicLen > uint16(MaxTopicLength) {
		return nil, fmt.Errorf("topic too long (%d > %d)", topicLen, MaxTopicLength)
	}
	topicBytes := make([]byte, topicLen)
	if _, err := io.ReadFull(r, topicBytes); err != nil {
		return nil, fmt.Errorf("failed to read topic: %w", err)
	}
	msg.Topic = string(topicBytes)

	var payloadLen uint32
	if err := binary.Read(r, binary.BigEndian, &payloadLen); err != nil {
		return nil, fmt.Errorf("failed to read payload length: %w", err)
	}
	if payloadLen > uint32(MaxPayloadLength) {
		return nil, fmt.Errorf("payload too long (%d > %d)", payloadLen, MaxPayloadLength)
	}
	msg.Payload = make([]byte, payloadLen)
	if _, err := io.ReadFull(r, msg.Payload); err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	var clientIDLen uint16
	if err := binary.Read(r, binary.BigEndian, &clientIDLen); err != nil {
		return nil, fmt.Errorf("failed to read client ID length: %w", err)
	}
	if clientIDLen > MaxClientIdLength {
		return nil, fmt.Errorf("client id too long (%d > %d)", clientIDLen, MaxClientIdLength)
	}
	clientIDBytes := make([]byte, clientIDLen)
	if _, err := io.ReadFull(r, clientIDBytes); err != nil {
		return nil, fmt.Errorf("failed to read client ID: %w", err)
	}
	msg.ClientId = string(clientIDBytes)

	var subscriptionIdLen uint16
	if err := binary.Read(r, binary.BigEndian, &subscriptionIdLen); err != nil {
		return nil, fmt.Errorf("failed to read subscription ID length: %w", err)
	}
	if subscriptionIdLen > MaxSubscriptionIdLength {
		return nil, fmt.Errorf("subscription id too long (%d > %d)", subscriptionIdLen, MaxSubscriptionIdLength)
	}
	subscriptionIdBytes := make([]byte, subscriptionIdLen)
	if _, err := io.ReadFull(r, subscriptionIdBytes); err != nil {
		return nil, fmt.Errorf("failed to read subscription ID: %w", err)
	}
	msg.SubscriptionId = string(subscriptionIdBytes)

	return msg, nil
}
