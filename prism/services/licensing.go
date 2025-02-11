package services

import (
	"errors"
	"net/http"
	"prism/api"
	"prism/services/licensing"

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

	license, err := licensing.CreateLicense(s.db, params.Name, params.Expiration)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return api.CreateLicenseResponse{License: license}, nil
}

func (s *LicenseService) DeactivateLicense(r *http.Request) (any, error) {
	licenseId, err := URLParamUUID(r, "license_key")
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	if err := s.db.Transaction(func(txn *gorm.DB) error {
		return licensing.DeactivateLicense(txn, licenseId)
	}); err != nil {
		if errors.Is(err, licensing.ErrLicenseNotFound) {
			return nil, CodedError(err, http.StatusNotFound)
		}
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return nil, nil
}

func (s *LicenseService) ListLicenses(r *http.Request) (any, error) {
	licenses, err := licensing.ListLicenses(s.db)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return licenses, nil
}
