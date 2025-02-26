package licensing

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"log/slog"
	"prism/prism/api"
	"prism/prism/schema"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrLicenseNotFound        = errors.New("license not found")
	ErrLicenseRetrievalFailed = errors.New("error retrieving license")
	ErrLicenseCreationFailed  = errors.New("license generation failed")
	ErrInvalidLicense         = errors.New("invalid license")
	ErrExpiredLicense         = errors.New("expired license")
	ErrDeactivatedLicense     = errors.New("license is deactivated")
	ErrMissingLicense         = errors.New("user does not have registered license")
)

type LicensePayload struct { // This is only exported to use to create a malformed license for a test.
	Id     uuid.UUID
	Secret []byte
}

const (
	versionPrefix = "V1-"
)

func (l *LicensePayload) Serialize() (string, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(l); err != nil {
		slog.Error("error serializing license", "error", err)
		return "", ErrLicenseCreationFailed
	}

	return versionPrefix + base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}

func parseLicense(licenseKey string) (*LicensePayload, error) {
	if !strings.HasPrefix(licenseKey, versionPrefix) {
		slog.Error("license is invalid, missing version prefix")
		return nil, ErrInvalidLicense
	}

	data, err := base64.URLEncoding.DecodeString(licenseKey[len(versionPrefix):])
	if err != nil {
		slog.Error("error decoding base64 encoding of license", "error", err)
		return nil, ErrInvalidLicense
	}

	var license LicensePayload
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&license); err != nil {
		slog.Error("error parsing license bytes", "error", err)
		return nil, ErrInvalidLicense
	}

	return &license, nil
}

func (l *LicensePayload) verifySecret(hashedSecret []byte) error {
	err := bcrypt.CompareHashAndPassword(hashedSecret, l.Secret)
	if err != nil {
		slog.Info("license verification failed", "error", err)
		return ErrInvalidLicense
	}
	return nil
}

func CreateLicense(txn *gorm.DB, name string, expiration time.Time) (api.CreateLicenseResponse, error) {
	secret := make([]byte, 48)
	if _, err := rand.Read(secret); err != nil {
		slog.Error("error generating license secret", "error", err)
		return api.CreateLicenseResponse{}, ErrLicenseCreationFailed
	}

	license := LicensePayload{Id: uuid.New(), Secret: secret}

	licenseKey, err := license.Serialize()
	if err != nil {
		return api.CreateLicenseResponse{}, err
	}

	hashedSecret, err := bcrypt.GenerateFromPassword(license.Secret, bcrypt.DefaultCost)
	if err != nil {
		slog.Error("error hashing license secret", "error", err)
		return api.CreateLicenseResponse{}, ErrLicenseCreationFailed
	}

	licenseEntry := schema.License{
		Id:          license.Id,
		Secret:      hashedSecret,
		Name:        name,
		Expiration:  expiration.UTC(),
		Deactivated: false,
	}

	if err := txn.Create(&licenseEntry).Error; err != nil {
		slog.Error("error saving license entry to db", "error", err)
		return api.CreateLicenseResponse{}, ErrLicenseCreationFailed
	}

	return api.CreateLicenseResponse{Id: licenseEntry.Id, License: licenseKey}, nil
}

func getLicenseForUpdate(txn *gorm.DB, id uuid.UUID) (schema.License, error) {
	var license schema.License
	if err := txn.Clauses(clause.Locking{Strength: "UPDATE"}).First(&license, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return license, ErrLicenseNotFound
		}
		slog.Error("error retrieving license from db", "license_id", id, "error", err)
		return license, ErrLicenseRetrievalFailed
	}
	return license, nil
}

func DeactivateLicense(txn *gorm.DB, id uuid.UUID) error {
	license, err := getLicenseForUpdate(txn, id)
	if err != nil {
		return err
	}

	license.Deactivated = true

	if err := txn.Save(&license).Error; err != nil {
		slog.Error("error deactivating license in db", "error", err)
		return errors.New("license deactivation failed")
	}

	return nil
}

func AddLicenseUser(txn *gorm.DB, licenseKey string, userId uuid.UUID) error {
	licenseData, err := parseLicense(licenseKey)
	if err != nil {
		return err
	}

	license, err := getLicenseForUpdate(txn, licenseData.Id)
	if err != nil {
		return err
	}

	if err := licenseData.verifySecret(license.Secret); err != nil {
		return err
	}

	if license.Expiration.Before(time.Now().UTC()) {
		slog.Info("license is expired")
		return ErrExpiredLicense
	}

	if license.Deactivated {
		slog.Info("license is deactivated")
		return ErrDeactivatedLicense
	}

	licenseUser := schema.LicenseUser{LicenseId: license.Id, UserId: userId}
	if err := txn.Save(&licenseUser).Error; err != nil {
		slog.Error("db error saving license user", "error", err)
		return errors.New("error adding license user")
	}

	return nil
}

func VerifyLicenseForReport(txn *gorm.DB, userId uuid.UUID) (uuid.UUID, error) {
	var licenseUser schema.LicenseUser

	if err := txn.Preload("License").First(&licenseUser, "user_id = ?", userId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, ErrMissingLicense
		}
		slog.Error("error retreiving user license from db", "error", err)
		return uuid.Nil, ErrLicenseRetrievalFailed
	}

	if licenseUser.License.Expiration.Before(time.Now().UTC()) {
		slog.Info("license is expired")
		return uuid.Nil, ErrExpiredLicense
	}

	if licenseUser.License.Deactivated {
		slog.Info("license is deactivated")
		return uuid.Nil, ErrDeactivatedLicense
	}

	return licenseUser.LicenseId, nil
}

func ListLicenses(txn *gorm.DB) ([]api.License, error) {
	var licenses []schema.License
	if err := txn.Find(&licenses).Error; err != nil {
		slog.Error("error retrieving licenses", "error", err)
		return nil, ErrLicenseRetrievalFailed
	}

	results := make([]api.License, 0, len(licenses))
	for _, license := range licenses {
		results = append(results, api.License{
			Id:          license.Id,
			Name:        license.Name,
			Expiration:  license.Expiration,
			Deactivated: license.Deactivated,
		})
	}

	return results, nil
}
