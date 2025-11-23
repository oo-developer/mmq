package common

const (
	COMMAND_ADD_USER byte = iota
	COMMAND_REMOVE_USER
	COMMAND_LIST_USERS
	COMMAND_LIST_CONNECTIONS
)

type CliService interface {
	Service
	Execute(clientId string, command []byte) []byte
}

type CliRequest struct {
	Type byte `json:"type"`
}

type CliResponse struct {
	Error        bool   `json:"error"`
	ErrorMessage string `json:"errorMessage"`
}

type AddUserReq struct {
	CliRequest
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}

type AddUserResp struct {
	CliResponse
	PrivateKeyPem string `json:"privateKeyPem"`
}

type RemoveUserReq struct {
	CliRequest
	Name string `json:"name"`
}

type RemoveUserResp struct {
	CliResponse
}

type ListUsersReq struct {
	CliRequest
}

type UserResp struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}

type ListUsersResp struct {
	CliResponse
	Users []UserResp `json:"users"`
}

type ListConnectionsReq struct {
	CliRequest
}

type ConnectionResp struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
}

type ListConnectionsResp struct {
	CliResponse
	Connections []ConnectionResp `json:"connections"`
}
