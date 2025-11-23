package cli

import (
	"fmt"

	"github.com/oo-developer/mmq/src/common"
	"github.com/oo-developer/mmq/src/config"
	log "github.com/oo-developer/mmq/src/logging"
	"github.com/vmihailenco/msgpack/v5"
)

type cli struct {
	config        *config.Config
	userService   common.UserService
	brokerService common.BrokerService
}

func NewCliService(config *config.Config, userService common.UserService, brokerService common.BrokerService) common.CliService {
	c := &cli{
		config:        config,
		userService:   userService,
		brokerService: brokerService,
	}
	return c
}

func (c *cli) Start() {

}

func (c *cli) Shutdown() {
}

func (c *cli) Execute(clientId string, payload []byte) []byte {
	request := common.CliRequest{}
	err := msgpack.Unmarshal(payload, &request)
	if err != nil {
		return c.returnError(err)
	}
	client := c.brokerService.Client(clientId)
	switch request.Type {
	case common.COMMAND_ADD_USER:
		return c.addUser(client, payload)
	case common.COMMAND_REMOVE_USER:
		return c.removeUser(client, payload)
	case common.COMMAND_LIST_USERS:
		return c.allUsers(client, payload)
	case common.COMMAND_LIST_CONNECTIONS:
		return c.allConnections(client, payload)
	default:
		log.Errorf("Unknown cli command type: %v", request.Type)
		c.returnError(fmt.Errorf("unknown cli command type: %v", request.Type))
	}
	return c.returnError(fmt.Errorf("unknown cli command type: %v", request.Type))
}

func (c *cli) addUser(client common.BrokerClient, command []byte) []byte {
	if !client.User().IsAdmin() {
		return c.returnError(fmt.Errorf("user '%s' is not admin", client.User().Name()))
	}
	request := common.AddUserReq{}
	err := msgpack.Unmarshal(command, &request)
	if err != nil {
		return c.returnError(err)
	}
	privateKeyPem, err := c.userService.AddUser(request.Name, request.Admin)
	if err != nil {
		return c.returnError(err)
	}
	response := &common.AddUserResp{
		PrivateKeyPem: privateKeyPem,
	}
	value, err := msgpack.Marshal(response)
	if err != nil {
		return c.returnError(err)
	}
	return value
}

func (c *cli) removeUser(client common.BrokerClient, command []byte) []byte {
	if !client.User().IsAdmin() {
		return c.returnError(fmt.Errorf("user '%s' is not admin", client.User().Name()))
	}
	request := common.RemoveUserReq{}
	err := msgpack.Unmarshal(command, &request)
	if err != nil {
		return c.returnError(err)
	}
	if request.Name == client.User().Name() {
		return c.returnError(fmt.Errorf("user '%s' cannot be removed", client.User().Name()))
	}
	err = c.userService.RemoveUserByName(request.Name)
	if err != nil {
		return c.returnError(err)
	}
	response := &common.RemoveUserResp{}
	value, err := msgpack.Marshal(response)
	if err != nil {
		return c.returnError(err)
	}
	return value
}

func (c *cli) allUsers(client common.BrokerClient, payload []byte) []byte {
	if !client.User().IsAdmin() {
		return c.returnError(fmt.Errorf("user '%s' is not admin", client.User().Name()))
	}
	resultList := &common.ListUsersResp{}
	resultList.Users = make([]common.UserResp, 0)
	for _, entry := range c.userService.AllUsers() {
		resultList.Users = append(resultList.Users, common.UserResp{
			Name:  entry.Name(),
			Admin: entry.IsAdmin(),
		})
	}
	value, err := msgpack.Marshal(resultList)
	if err != nil {
		return c.returnError(err)
	}
	return value
}

func (c *cli) allConnections(client common.BrokerClient, payload []byte) []byte {
	if !client.User().IsAdmin() {
		return c.returnError(fmt.Errorf("user '%s' is not admin", client.User().Name()))
	}
	resultList := &common.ListConnectionsResp{}
	resultList.Connections = make([]common.ConnectionResp, 0)
	for _, entry := range c.brokerService.AllClients() {
		resultList.Connections = append(resultList.Connections, common.ConnectionResp{
			Id:       entry.Id(),
			Username: entry.User().Name(),
			Admin:    entry.User().IsAdmin(),
		})
	}
	value, err := msgpack.Marshal(resultList)
	if err != nil {
		return c.returnError(err)
	}
	return value
}

func (c *cli) returnError(err error) []byte {
	response := common.CliResponse{
		Error:        true,
		ErrorMessage: err.Error(),
	}
	value, _ := msgpack.Marshal(response)
	return value
}
