package api

type noCipher struct {
}

func NewNoCipher() Cipher {
	return &noCipher{}
}

func (c *noCipher) Encrypt(plaintext []byte) ([]byte, error) {
	return plaintext, nil
}

func (c *noCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

func (c *noCipher) Enable(enable bool) {
}
