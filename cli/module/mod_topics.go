package module

import (
	"errors"
	"flag"
	"fmt"

	api "github.com/oo-developer/mmq/pkg"
	"github.com/oo-developer/mmq/src/common"
	"github.com/vmihailenco/msgpack/v5"
)

type modTopics struct {
	commands map[string]Command
}

func NewModTopics() Module {
	m := &modTopics{
		commands: make(map[string]Command),
	}
	m.commands["list"] = m.List
	m.commands["publish"] = m.Publish
	m.commands["help"] = m.Help
	return m
}

func (m *modTopics) Execute(client *api.Client, commandName string, args ...string) error {
	command, ok := m.commands[commandName]
	if !ok {
		return m.Help(client, args...)
	}
	return command(client, args...)
}

func (m *modTopics) List(client *api.Client, args ...string) error {
	request := common.ListTopicsReq{
		CliRequest: common.CliRequest{
			Type: common.COMMAND_LIST_TOPICS,
		},
	}
	requestBytes, _ := msgpack.Marshal(request)
	responseBytes, err := client.SendCommand(requestBytes)
	if err != nil {
		return err
	}
	response := common.ListTopicsResp{}
	if err := msgpack.Unmarshal(responseBytes, &response); err != nil {
		return err
	}
	if response.Error {
		return errors.New(response.ErrorMessage)
	}
	fmt.Printf("%-80s %-4s %-4s\n", "TOPIC", "RET", "PER")
	for _, entry := range response.Topics {
		retained := ""
		if entry.Retained {
			retained = " X"
		}
		persistent := ""
		if entry.Persistent {
			persistent = " X"
		}
		fmt.Printf("%-80s %-4s %-4s\n", entry.Topic, retained, persistent)
	}
	return nil
}

func (m *modTopics) Publish(client *api.Client, args ...string) error {
	flagSet := flag.NewFlagSet("topics publish", flag.ContinueOnError)
	topic := flagSet.String("topic", "", "The topic to send to")
	payload := flagSet.String("payload", "", "The payload")
	persistent := flagSet.Bool("persistent", false, "Whether or not the topic is persistent")
	retained := flagSet.Bool("retained", false, "Whether or not the topic is retained")
	flagSet.Parse(args)
	if topic == nil {
		return errors.New("--topic is required")
	}
	if payload == nil {
		return errors.New("--payload is required")
	}
	if retained != nil && *retained && persistent != nil && *persistent {
		err := client.Publish(*topic, []byte(*payload), api.Persistent, api.Retained)
		if err != nil {
			return err
		}
	} else if retained != nil && *retained {
		err := client.Publish(*topic, []byte(*payload), api.Retained)
		if err != nil {
			return err
		}
	} else if persistent != nil && *persistent {
		err := client.Publish(*topic, []byte(*payload), api.Persistent)
		if err != nil {
			return err
		}
	} else {
		err := client.Publish(*topic, []byte(*payload))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *modTopics) Help(client *api.Client, args ...string) error {
	return nil
}
