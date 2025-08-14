package licensing

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidLicense = errors.New("invalid license")
	ErrExpiredLicense = errors.New("expired license")
)

type Payload struct {
	Expiration time.Time
}

type License struct {
	Payload   []byte
	Signature []byte
}

type OnPremLicenseVerifier struct {
	publicKeyPem []byte
	licenseStr   string
}

func NewOnPremLicenseVerifier(publicKeyPem []byte, licenseStr string) *OnPremLicenseVerifier {

	return &OnPremLicenseVerifier{
		publicKeyPem: publicKeyPem,
		licenseStr:   licenseStr,
	}
}

func (v *OnPremLicenseVerifier) VerifyLicense() error {
	publicKey, err := parseRsaPublicKey(v.publicKeyPem)
	if err != nil {
		return ErrInvalidLicense
	}

	payload, err := DecodeLicense(publicKey, v.licenseStr)
	if err != nil {
		return ErrInvalidLicense
	}

	if payload.Expiration.IsZero() {
		return ErrInvalidLicense
	}

	if payload.Expiration.Before(time.Now().UTC()) {
		return ErrExpiredLicense
	}

	return nil
}

func DecodeLicense(publicKey *rsa.PublicKey, licenseStr string) (Payload, error) {
	licenseBytes, err := base64.StdEncoding.DecodeString(licenseStr)
	if err != nil {
		return Payload{}, ErrInvalidLicense
	}

	var license License
	if err := gob.NewDecoder(bytes.NewReader(licenseBytes)).Decode(&license); err != nil {
		return Payload{}, ErrInvalidLicense
	}

	if err := verifySignature(publicKey, license.Payload, license.Signature); err != nil {
		return Payload{}, ErrInvalidLicense
	}

	var payload Payload
	if err := json.Unmarshal(license.Payload, &payload); err != nil {
		return Payload{}, ErrInvalidLicense
	}

	return payload, nil
}

func CreateLicense(privateKeyPem []byte, expiration time.Time) (string, error) {
	privateKey, err := parseRsaPrivateKey(privateKeyPem)
	if err != nil {
		return "", fmt.Errorf("error parsing private key: %w", err)
	}

	payload := Payload{
		Expiration: expiration,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshaling payload: %w", err)
	}

	signature, err := signMessage(privateKey, payloadBytes)
	if err != nil {
		return "", fmt.Errorf("error signing payload: %w", err)
	}

	license := License{
		Payload:   payloadBytes,
		Signature: signature,
	}

	licenseBytes := bytes.Buffer{}
	if err := gob.NewEncoder(&licenseBytes).Encode(license); err != nil {
		return "", fmt.Errorf("error encoding license: %w", err)
	}

	return base64.StdEncoding.EncodeToString(licenseBytes.Bytes()), nil
}

func GenerateKeys() ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating private key: %w", err)
	}

	publicKey := &privateKey.PublicKey

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling private key: %w", err)
	}

	privateKeyPem, err := encodeToPem("PRIVATE KEY", privateKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("error encoding private key to PEM: %w", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling public key: %w", err)
	}
	publicKeyPem, err := encodeToPem("PUBLIC KEY", publicKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("error encoding public key to PEM: %w", err)
	}

	return privateKeyPem, publicKeyPem, nil
}

func encodeToPem(blockType string, bytes []byte) ([]byte, error) {
	pemBlock := &pem.Block{
		Type:  blockType,
		Bytes: bytes,
	}

	pemBytes := pem.EncodeToMemory(pemBlock)
	if pemBytes == nil {
		return nil, fmt.Errorf("failed to encode PEM block")
	}

	return pemBytes, nil
}

func parseRsaPublicKey(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not of type RSA")
	}

	return rsaPublicKey, nil
}

func parseRsaPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err)
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not of type RSA")
	}

	return rsaPrivateKey, nil
}

func signMessage(privateKey *rsa.PrivateKey, message []byte) ([]byte, error) {
	hash := crypto.SHA256.New()
	hash.Write(message)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash.Sum(nil))
	if err != nil {
		return nil, fmt.Errorf("error signing message: %w", err)
	}

	return signature, nil
}

func verifySignature(publicKey *rsa.PublicKey, message []byte, signature []byte) error {
	hash := crypto.SHA256.New()
	hash.Write(message)

	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash.Sum(nil), signature); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}

const OnPremLicensePublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzKliSDC4qidOVSBJSeUh
t8no1R0gJSI+iYcv/xWsbKO4fLq2QI9DVwOjzg7oSkNNh06i6wS3kvzPFcRbLBIk
sNHkt+62Xlaj9SSGgqTDsmtyFJxRGdgD1Q0wqZ4RHXiddBXEt1uJ4spPXDVhmhRM
48flfmVIsPp8ruJ8yZ1JuOSoIkjfXud+bwfVTs81T5weUsHPl6q388IvmPnWi+16
OGH9VhkrSyrshDdVomen8pSFejuqEWNZYtf6EdaHLtWlNQ7/NmaHAT67YoPZXG9N
Selj6fiUSK7QfocrTd1PGg8N/UkFfg6ppJgzhj6lqdf6cOKWpUV7xNiAGk/x8/V0
LQIDAQAB
-----END PUBLIC KEY-----
`
