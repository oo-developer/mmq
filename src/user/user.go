package user

import (
	"fmt"
	"os"

	api "github.com/oo-developer/mmq/pkg"
	"github.com/oo-developer/mmq/src/common"
	"github.com/oo-developer/mmq/src/config"
	log "github.com/oo-developer/mmq/src/logging"
)

type user struct {
	name         string
	admin        bool
	publicKeyPem string
	publicKey    api.KyberPublicKey
}

func (n *user) Name() string {
	return n.name
}

func (n *user) PublicKeyPem() string {
	return n.publicKeyPem
}

func (n *user) IsAdmin() bool {
	return n.admin
}

func (n *user) PublicKey() *api.KyberPublicKey {
	return &n.publicKey
}

type users struct {
	config         *config.Config
	users          map[string]*user
	storageService common.StorageService
}

func NewUserService(config *config.Config, storageService common.StorageService) common.UserService {
	u := &users{
		config:         config,
		users:          make(map[string]*user),
		storageService: storageService,
	}
	return u
}

func (u *users) Start() {
	u.load()
	log.Info("UserService started")
}

func (u *users) load() {
	userList := u.storageService.GetAllUsers()
	for _, entry := range userList {
		userEntry := &user{
			name:         entry.Name(),
			admin:        entry.IsAdmin(),
			publicKeyPem: entry.PublicKeyPem(),
		}

		publicKey, err := api.LoadKyberPublicKey([]byte(userEntry.publicKeyPem))
		if err != nil {
			log.Errorf("load public key failed for user '%s': %s", entry.Name(), err.Error())
		} else {
			userEntry.publicKey = *publicKey
		}
		u.users[entry.Name()] = userEntry
	}
}

func (u *users) Shutdown() {
	log.Info("UserService shut down")
}

func (u *users) LookupUserByName(name string) (common.User, bool) {
	user, ok := u.users[name]
	return user, ok
}

func (u *users) AddUser(userName string, admin bool) (string, error) {
	if _, ok := u.users[userName]; ok {
		return "", fmt.Errorf("user '%s' already exists", userName)
	}
	publicKey, privateKey, err := api.GenerateKyberKeyPair()
	if err != nil {
		return "", err
	}
	publicKeyBytes, err := api.EncodeKyberPublicKeyPEM(publicKey)
	if err != nil {
		return "", err
	}
	privateKeyBytes, err := api.EncodeKyberPrivateKeyPEM(privateKey)
	if err != nil {
		return "", err
	}
	err = os.WriteFile(fmt.Sprintf("%s_private_key.pem", userName), privateKeyBytes, 0600)
	if err != nil {
		return "", err
	}
	userEntry := &user{
		name:         userName,
		admin:        admin,
		publicKeyPem: string(publicKeyBytes),
		publicKey:    *publicKey,
	}
	err = u.storageService.AddUser(userEntry)
	if err != nil {
		return "", err
	}
	u.users[userName] = userEntry
	return string(privateKeyBytes), nil
}

func (u *users) RemoveUserByName(userName string) error {
	if _, ok := u.users[userName]; !ok {
		return fmt.Errorf("user '%s' does not exists", userName)
	}
	err := u.storageService.RemoveUserByName(userName)
	if err != nil {
		return err
	}
	delete(u.users, userName)
	return nil
}

func (u *users) AllUsers() []common.User {
	userList := make([]common.User, 0, len(u.users))
	for _, user := range u.users {
		userList = append(userList, user)
	}
	return userList
}
