package licensing

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-resty/resty/v2"
)

var (
	ErrLicenseVerificationFailed = errors.New("license verification failed")
	ErrLicenseNotFound           = errors.New("license not found")
	ErrInvalidLicense            = errors.New("invalid license")
	ErrExpiredLicense            = errors.New("expired license")
)

type LicenseVerifier struct {
	client     *resty.Client
	licenseKey string
}

func NewLicenseVerifier(licenseKey string) (*LicenseVerifier, error) {
	verifier := &LicenseVerifier{
		client:     resty.New().SetBaseURL("https://api.keygen.sh"),
		licenseKey: licenseKey,
	}

	if err := verifier.VerifyLicense(); err != nil {
		return nil, err
	}

	return verifier, nil
}

type licenseVerifyResponse struct {
	Meta struct {
		Valid    bool   `json:"valid"`
		Constant string `json:"constant"`
		Detail   string `json:"detail"`
	} `json:"meta"`
}

func (verifier *LicenseVerifier) VerifyLicense() error {
	rb := map[string]map[string]any{
		"meta": {
			"key": verifier.licenseKey,
			"scope": map[string]any{
				"entitlements": []string{"FULL_ACCESS", "PRISM"},
			},
		},
	}

	endpoint := "/v1/accounts/thirdai/licenses/actions/validate-key"
	res, err := verifier.client.R().
		SetHeader("Content-Type", "application/vnd.api+json").
		SetHeader("Accept", "application/vnd.api+json").
		SetBody(rb).
		Post(endpoint)

	if err != nil {
		slog.Error("unable to verify license with keygen", "error", err)
		return ErrLicenseVerificationFailed
	}

	if !res.IsSuccess() {
		slog.Error("keygen returned error", "status_code", res.StatusCode(), "body", res.String())
		return ErrLicenseVerificationFailed
	}

	body := res.Body()

	var verified licenseVerifyResponse
	if err := json.Unmarshal(body, &verified); err != nil {
		slog.Error("error parsing response from keygen", "error", err)
		return ErrLicenseVerificationFailed
	}

	if !verified.Meta.Valid {
		switch verified.Meta.Constant {
		case "NOT_FOUND":
			return ErrLicenseNotFound
		case "EXPIRED", "SUSPENDED":
			return ErrExpiredLicense
		default:
			return ErrInvalidLicense
		}
	}

	if err := verifyResponseSignature(res, endpoint); err != nil {
		slog.Error("unable to verify response signature", "error", err)
		return ErrLicenseVerificationFailed
	}

	return nil
}

func parseSignature(res *resty.Response) ([]byte, error) {
	header := res.Header().Get("keygen-signature")
	_, sig, _ := strings.Cut(header, `signature="`)
	sig, _, _ = strings.Cut(sig, `", `)

	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return nil, fmt.Errorf("error response decoding signature: %w", err)
	}
	return sigBytes, nil
}

func parsePublicKey() (ed25519.PublicKey, error) {
	keyDER, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("error decoding public key: %w", err)
	}

	pubKey, err := x509.ParsePKIXPublicKey(keyDER)
	if err != nil {
		return nil, fmt.Errorf("error parsing DER encoded public key: %w", err)
	}

	ed25519PubKey, ok := pubKey.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not of type ed25519")
	}
	return ed25519PubKey, nil
}

const signingDataTemplate = `(request-target): %s %s
host: api.keygen.sh
date: %s
digest: sha-256=%s`

const publicKeyBase64 = "MCowBQYDK2VwAyEAmtv9iB02PTHBVsNImWiS3QGDp+RUDcABy3wu7Fp5Zq4="

// This verifies that the response originated at the keygen server by verifying
// the response signature using the public key corresponding to the private key
// used by keygen.
// https://keygen.sh/docs/api/signatures/#response-signatures
func verifyResponseSignature(res *resty.Response, endpoint string) error {
	bodyHash := sha256.Sum256(res.Body())
	bodyHashBase64 := base64.StdEncoding.EncodeToString(bodyHash[:])

	signingData := fmt.Sprintf(signingDataTemplate, strings.ToLower(res.Request.Method), endpoint, res.Header().Get("date"), bodyHashBase64)

	publicKey, err := parsePublicKey()
	if err != nil {
		return err
	}

	signature, err := parseSignature(res)
	if err != nil {
		return err
	}

	if !ed25519.Verify(publicKey, []byte(signingData), signature) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}
