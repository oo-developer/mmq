package common

import api "github.com/oo-developer/mmq/pkg"

type StorageService interface {
	Service
	GetAllMessages() []*api.Message
	AddMessageChannel() chan *api.Message
	RemoveMessageChannel() chan string
	GetAllUsers() []User
	AddUser(user User) error
	RemoveUserByName(userName string) error
}
