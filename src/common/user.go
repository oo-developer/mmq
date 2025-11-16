package common

import (
	api "github.com/oo-developer/tinymq/pkg"
)

type User interface {
	Name() string
	PublicKey() *api.KyberPublicKey
}

type UserService interface {
	Service
	LookupUserByName(name string) (User, bool)
	AddUser(userName string) error
	RemoveUserByName(userName string) error
}
