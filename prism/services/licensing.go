package services

import (
	"errors"
	"log/slog"
	"net/http"
	"prism/prism/api"
	"prism/prism/services/licensing"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type LicenseService struct {
	db *gorm.DB
}

func (s *LicenseService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/create", WrapRestHandler(s.CreateLicense))
	r.Delete("/{license_id}", WrapRestHandler(s.DeactivateLicense))
	r.Get("/list", WrapRestHandler(s.ListLicenses))

	return r
}

func (s *LicenseService) CreateLicense(r *http.Request) (any, error) {
	params, err := ParseRequestBody[api.CreateLicenseRequest](r)
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	slog.Info("creating license", "name", params.Name)

	license, err := licensing.CreateLicense(s.db, params.Name, params.Expiration)
	if err != nil {
		slog.Error("error creating license", "name", params.Name, "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	slog.Info("license created", "name", params.Name, "license_id", license.Id)

	return license, nil
}

func (s *LicenseService) DeactivateLicense(r *http.Request) (any, error) {
	licenseId, err := URLParamUUID(r, "license_id")
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	slog.Info("deactivating license", "license_id", licenseId)

	if err := s.db.Transaction(func(txn *gorm.DB) error {
		return licensing.DeactivateLicense(txn, licenseId)
	}); err != nil {
		slog.Error("error deleting license", "license_id", licenseId, "error", err)
		if errors.Is(err, licensing.ErrLicenseNotFound) {
			return nil, CodedError(err, http.StatusNotFound)
		}
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	slog.Info("license deactivated", "license_id", licenseId)

	return nil, nil
}

func (s *LicenseService) ListLicenses(r *http.Request) (any, error) {
	licenses, err := licensing.ListLicenses(s.db)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return licenses, nil
}
