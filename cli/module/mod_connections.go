package module

import (
	"errors"
	"fmt"

	api "github.com/oo-developer/mmq/pkg"
	"github.com/oo-developer/mmq/src/common"
	"github.com/vmihailenco/msgpack/v5"
)

type modClients struct {
	client   *api.Client
	commands map[string]Command
}

func NewModClients() Module {
	m := &modClients{
		commands: make(map[string]Command),
	}
	m.commands["list"] = m.List
	return m
}

func (m *modClients) Execute(client *api.Client, commandName string, args ...string) error {
	command, ok := m.commands[commandName]
	if !ok {
		return m.Help(client, args...)
	}
	return command(client, args...)
}

func (m *modClients) List(client *api.Client, args ...string) error {
	request := common.ListConnectionsReq{
		CliRequest: common.CliRequest{
			Type: common.COMMAND_LIST_CONNECTIONS,
		},
	}
	requestBytes, _ := msgpack.Marshal(request)
	responseBytes, err := client.SendCommand(requestBytes)
	if err != nil {
		return err
	}
	response := common.ListConnectionsResp{}
	if err := msgpack.Unmarshal(responseBytes, &response); err != nil {
		return err
	}
	if response.Error {
		return errors.New(response.ErrorMessage)
	}
	fmt.Printf("%-36s %-20s %-8s\n", "ID", "USER NAME", "ROLE")
	for _, entry := range response.Connections {
		role := "user"
		if entry.Admin {
			role = "admin"
		}
		fmt.Printf("%-36s %-20s %s\n", entry.Id, entry.Username, role)
	}
	return nil
}

func (m *modClients) Help(client *api.Client, args ...string) error {
	return nil
}
