package module

import (
	"errors"
	"flag"
	"fmt"
	"os"

	api "github.com/oo-developer/mmq/pkg"
	"github.com/oo-developer/mmq/src/common"
	"github.com/vmihailenco/msgpack/v5"
)

type modUsers struct {
	commands map[string]Command
}

func NewModUsers() Module {
	m := &modUsers{
		commands: make(map[string]Command),
	}
	m.commands["list"] = m.List
	m.commands["add"] = m.Add
	m.commands["remove"] = m.Remove
	m.commands["help"] = m.Help
	return m
}

func (m *modUsers) Execute(client *api.Client, commandName string, args ...string) error {
	command, ok := m.commands[commandName]
	if !ok {
		return m.Help(client, args...)
	}
	return command(client, args...)
}

func (m *modUsers) List(client *api.Client, args ...string) error {
	request := common.ListUsersReq{
		CliRequest: common.CliRequest{
			Type: common.COMMAND_LIST_USERS,
		},
	}
	requestBytes, _ := msgpack.Marshal(request)
	responseBytes, err := client.SendCommand(requestBytes)
	if err != nil {
		return err
	}
	response := common.ListUsersResp{}
	if err := msgpack.Unmarshal(responseBytes, &response); err != nil {
		return err
	}
	if response.Error {
		return errors.New(response.ErrorMessage)
	}
	fmt.Printf("%-20s %-8s\n", "USER NAME", "ROLE")
	for _, entry := range response.Users {
		role := "user"
		if entry.Admin {
			role = "admin"
		}
		fmt.Printf("%-20s %s\n", entry.Name, role)
	}
	return nil
}

func (m *modUsers) Add(client *api.Client, args ...string) error {
	flagSet := flag.NewFlagSet("users add", flag.ContinueOnError)
	name := flagSet.String("name", "", "Name of the new user")
	admin := flagSet.Bool("admin", false, "Set to true if you want admin user")
	keyFile := flagSet.String("key-file", "", "The file to store the private key of the new user")
	flagSet.Parse(args)
	if *name == "" {
		return errors.New("name is required")
	}
	request := common.AddUserReq{
		CliRequest: common.CliRequest{
			Type: common.COMMAND_ADD_USER,
		},
		Name:  *name,
		Admin: *admin,
	}
	requestBytes, _ := msgpack.Marshal(request)
	responseBytes, err := client.SendCommand(requestBytes)
	if err != nil {
		return err
	}
	response := common.AddUserResp{}
	err = msgpack.Unmarshal(responseBytes, &response)
	if err != nil {
		return err
	}
	if response.Error {
		return errors.New(response.ErrorMessage)
	}
	fmt.Println("[OK] User created successfully")
	if *keyFile == "" {
		fmt.Println("[OK] The private key is stored no where and can not be recovered")
		fmt.Println(response.PrivateKeyPem)
	} else {
		err := os.WriteFile(*keyFile, []byte(response.PrivateKeyPem), 0600)
		if err != nil {
			return err
		}
		fmt.Printf("[OK] Private key written to '%s'\n", *keyFile)
	}
	return nil
}

func (m *modUsers) Remove(client *api.Client, args ...string) error {
	flagSet := flag.NewFlagSet("users add", flag.ContinueOnError)
	name := flagSet.String("name", "", "Name of the new user")
	flagSet.Parse(args)
	if *name == "" {
		return errors.New("--name is required")
	}
	request := common.RemoveUserReq{
		CliRequest: common.CliRequest{
			Type: common.COMMAND_REMOVE_USER,
		},
		Name: *name,
	}
	requestBytes, _ := msgpack.Marshal(request)
	responseBytes, err := client.SendCommand(requestBytes)
	if err != nil {
		return err
	}
	response := common.RemoveUserResp{}
	err = msgpack.Unmarshal(responseBytes, &response)
	if err != nil {
		return err
	}
	if response.Error {
		return errors.New(response.ErrorMessage)
	}
	fmt.Println("[OK] User removed successfully")
	return nil
}

func (m *modUsers) Help(client *api.Client, args ...string) error {
	return nil
}
