package api

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/cloudflare/circl/kem/kyber/kyber768"
	"golang.org/x/crypto/chacha20poly1305"
)

type chaCha20Cipher struct {
	session *QuantumSafeSession
	enabled bool
}

func EstablishChCha20Cipher(publicServerKey *kyber768.PublicKey) (Cipher, []byte, error) {
	session, kemCiphertext, err := EstablishQuantumSafeSession(publicServerKey)
	if err != nil {
		return nil, nil, err
	}
	return &chaCha20Cipher{
		session: session,
		enabled: true,
	}, kemCiphertext, nil
}

func RecoverCHaCha20Cipher(privateServerKey *KyberPrivateKey, kemCiphertext []byte) (Cipher, error) {
	session, err := RecoverQuantumSafeSession(privateServerKey.key, kemCiphertext)
	if err != nil {
		return nil, err
	}
	return &chaCha20Cipher{
		session: session,
		enabled: true,
	}, nil
}

func (c *chaCha20Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	if !c.enabled {
		return plaintext, nil
	}
	return c.session.EncryptMessage(plaintext)
}

func (c *chaCha20Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if !c.enabled {
		return ciphertext, nil
	}
	return c.session.DecryptMessage(ciphertext)
}

func (c *chaCha20Cipher) Enable(enable bool) {
	c.enabled = enable
}

// QuantumSafeSession represents an encrypted session
type QuantumSafeSession struct {
	key []byte
}

// EstablishQuantumSafeSession creates a session key using Kyber KEM
// Client side: Returns session key + encrypted key material to send to server
func EstablishQuantumSafeSession(publicKey *kyber768.PublicKey) (*QuantumSafeSession, []byte, error) {
	scheme := kyber768.Scheme()

	// Encapsulate: generate shared secret + ciphertext
	kemCiphertext, sharedSecret, err := scheme.Encapsulate(publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("encapsulation failed: %w", err)
	}

	// Use first 32 bytes of shared secret as ChaCha20 key
	sessionKey := make([]byte, chacha20poly1305.KeySize)
	copy(sessionKey, sharedSecret[:32])

	return &QuantumSafeSession{key: sessionKey}, kemCiphertext, nil
}

// RecoverQuantumSafeSession decrypts the session key using Kyber private key
// Server side: Recovers the session key from encrypted key material
func RecoverQuantumSafeSession(privateKey *kyber768.PrivateKey, kemCiphertext []byte) (*QuantumSafeSession, error) {
	scheme := kyber768.Scheme()

	// Decapsulate: recover shared secret
	sharedSecret, err := scheme.Decapsulate(privateKey, kemCiphertext)
	if err != nil {
		return nil, fmt.Errorf("decapsulation failed: %w", err)
	}

	sessionKey := make([]byte, chacha20poly1305.KeySize)
	copy(sessionKey, sharedSecret[:32])

	return &QuantumSafeSession{key: sessionKey}, nil
}

// EncryptMessage encrypts a message with ChaCha20-Poly1305
// Format: [nonce (12 bytes)][ciphertext + authentication tag]
func (s *QuantumSafeSession) EncryptMessage(plaintext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, aead.NonceSize()) // 12 bytes for ChaCha20-Poly1305
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate in one operation
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	// Prepend nonce to ciphertext
	result := make([]byte, 0, len(nonce)+len(ciphertext))
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// DecryptMessage decrypts a message encrypted with EncryptMessage
func (s *QuantumSafeSession) DecryptMessage(ciphertext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and encrypted data
	nonce := ciphertext[:nonceSize]
	encryptedData := ciphertext[nonceSize:]

	// Decrypt and verify authentication tag
	plaintext, err := aead.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption or authentication failed: %w", err)
	}

	return plaintext, nil
}

// EncryptMessageWithAAD encrypts with Additional Authenticated Data
// Use for binding metadata (topic, messageID, timestamp, etc.) to the message
// The AAD is authenticated but NOT encrypted
func (s *QuantumSafeSession) EncryptMessageWithAAD(plaintext, additionalData []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// additionalData is authenticated but not encrypted
	// Any tampering with AAD will fail decryption
	ciphertext := aead.Seal(nil, nonce, plaintext, additionalData)

	result := make([]byte, 0, len(nonce)+len(ciphertext))
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// DecryptMessageWithAAD decrypts with Additional Authenticated Data
func (s *QuantumSafeSession) DecryptMessageWithAAD(ciphertext, additionalData []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	encryptedData := ciphertext[nonceSize:]

	// Will fail if AAD doesn't match what was used during encryption
	plaintext, err := aead.Open(nil, nonce, encryptedData, additionalData)
	if err != nil {
		return nil, fmt.Errorf("decryption or authentication failed: %w", err)
	}

	return plaintext, nil
}
