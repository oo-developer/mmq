package storage

import api "github.com/oo-developer/mmq/pkg"

type user struct {
	NameValue         string `json:"name"`
	AdminValue        bool   `json:"admin"`
	PublicKeyPemValue string `json:"publicKeyPem"`
}

func (n *user) Name() string {
	return n.NameValue
}

func (n *user) PublicKeyPem() string {
	return n.PublicKeyPemValue
}

func (n *user) IsAdmin() bool {
	return n.AdminValue
}

func (n *user) PublicKey() *api.KyberPublicKey {
	return nil
}
