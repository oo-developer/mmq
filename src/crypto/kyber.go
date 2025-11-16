package crypto

import (
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/cloudflare/circl/kem/kyber/kyber768"
)

// KyberPublicKey wraps a Kyber public key
type KyberPublicKey struct {
	key *kyber768.PublicKey
}

// KyberPrivateKey wraps a Kyber private key
type KyberPrivateKey struct {
	key *kyber768.PrivateKey
}

// GenerateKyberKeyPair generates a new Kyber key pair
func GenerateKyberKeyPair() (*KyberPublicKey, *KyberPrivateKey, error) {
	pub, priv, err := kyber768.GenerateKeyPair(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Kyber key pair: %w", err)
	}
	return &KyberPublicKey{key: pub}, &KyberPrivateKey{key: priv}, nil
}

// SaveKyberPublicKey saves a Kyber public key to a PEM file
func SaveKyberPublicKey(filepath string, key *KyberPublicKey) error {
	pubBytes, err := key.key.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	pubBlock := &pem.Block{
		Type:  "KYBER768 PUBLIC KEY",
		Bytes: pubBytes,
	}

	pemData := pem.EncodeToMemory(pubBlock)
	if pemData == nil {
		return fmt.Errorf("failed to encode PEM")
	}

	if err := os.WriteFile(filepath, pemData, 0644); err != nil {
		return fmt.Errorf("failed to write public key file: %w", err)
	}

	return nil
}

// SaveKyberPrivateKey saves a Kyber private key to a PEM file
func SaveKyberPrivateKey(filepath string, key *KyberPrivateKey) error {
	privBytes, err := key.key.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	privBlock := &pem.Block{
		Type:  "KYBER768 PRIVATE KEY",
		Bytes: privBytes,
	}

	pemData := pem.EncodeToMemory(privBlock)
	if pemData == nil {
		return fmt.Errorf("failed to encode PEM")
	}

	// Private key should have restricted permissions (0600)
	if err := os.WriteFile(filepath, pemData, 0600); err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	return nil
}

// LoadKyberPublicKey loads a Kyber public key from a PEM file
func LoadKyberPublicKey(filepath string) (*KyberPublicKey, error) {
	pemData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "KYBER768 PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM type: expected 'KYBER768 PUBLIC KEY', got '%s'", block.Type)
	}

	scheme := kyber768.Scheme()
	pub, err := scheme.UnmarshalBinaryPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal public key: %v", err)
	}

	kyberPub, ok := pub.(*kyber768.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid public key type")
	}

	return &KyberPublicKey{key: kyberPub}, nil
}

// LoadKyberPrivateKey loads a Kyber private key from a PEM file
func LoadKyberPrivateKey(filepath string) (*KyberPrivateKey, error) {
	pemData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "KYBER768 PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM type: expected 'KYBER768 PRIVATE KEY', got '%s'", block.Type)
	}

	scheme := kyber768.Scheme()
	priv, err := scheme.UnmarshalBinaryPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %v", err)
	}

	kyberPriv, ok := priv.(*kyber768.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid private key type")
	}

	return &KyberPrivateKey{key: kyberPriv}, nil
}

// SaveKyberKeyPair saves both public and private keys
func SaveKyberKeyPair(publicPath, privatePath string, pub *KyberPublicKey, priv *KyberPrivateKey) error {
	if err := SaveKyberPublicKey(publicPath, pub); err != nil {
		return err
	}
	if err := SaveKyberPrivateKey(privatePath, priv); err != nil {
		return err
	}
	return nil
}
