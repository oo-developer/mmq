package common

import (
	api "github.com/oo-developer/mmq/pkg"
)

type User interface {
	Name() string
	IsAdmin() bool
	PublicKeyPem() string
	PublicKey() *api.KyberPublicKey
}

type UserService interface {
	Service
	LookupUserByName(name string) (User, bool)
	AddUser(userName string, admin bool) (string, error)
	RemoveUserByName(userName string) error
	AllUsers() []User
}
