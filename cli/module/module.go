package module

import api "github.com/oo-developer/mmq/pkg"

var Modules = map[string]Module{
	"users":       NewModUsers(),
	"connections": NewModClients(),
}

type Command func(client *api.Client, args ...string) error

type Module interface {
	Execute(client *api.Client, command string, args ...string) error
}
