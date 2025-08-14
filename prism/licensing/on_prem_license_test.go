package licensing_test

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"prism/prism/licensing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnPremLicensing(t *testing.T) {
	privateKey, publicKey, err := licensing.GenerateKeys()
	require.NoError(t, err)

	t.Run("valid license", func(t *testing.T) {
		expiry := time.Now().Add(time.Hour)
		goodLicense, err := licensing.CreateLicense([]byte(privateKey), expiry)
		require.NoError(t, err)

		verifier := licensing.NewOnPremLicenseVerifier([]byte(publicKey), goodLicense)
		assert.NoError(t, verifier.VerifyLicense())
	})

	t.Run("expired license", func(t *testing.T) {
		expiredLicense, err := licensing.CreateLicense([]byte(privateKey), time.Now().Add(-time.Hour))
		require.NoError(t, err)

		verifier := licensing.NewOnPremLicenseVerifier([]byte(publicKey), expiredLicense)
		assert.ErrorIs(t, licensing.ErrExpiredLicense, verifier.VerifyLicense())
	})

	t.Run("invalid license", func(t *testing.T) {
		original, err := licensing.CreateLicense([]byte(privateKey), time.Now().Add(time.Hour))
		require.NoError(t, err)

		licenseBytes, err := base64.StdEncoding.DecodeString(original)
		require.NoError(t, err)

		var corruptedLicense licensing.License
		require.NoError(t, gob.NewDecoder(bytes.NewReader(licenseBytes)).Decode(&corruptedLicense))

		corruptedLicense.Payload, err = json.Marshal(licensing.Payload{Expiration: time.Now().Add(2 * time.Hour)})
		require.NoError(t, err)

		buf := bytes.Buffer{}
		require.NoError(t, gob.NewEncoder(&buf).Encode(corruptedLicense))

		corruptedLicenseStr := base64.StdEncoding.EncodeToString(buf.Bytes())

		verifier := licensing.NewOnPremLicenseVerifier([]byte(publicKey), corruptedLicenseStr)
		assert.ErrorIs(t, licensing.ErrInvalidLicense, verifier.VerifyLicense())
	})
}
