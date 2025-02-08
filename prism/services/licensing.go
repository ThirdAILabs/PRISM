package services

import (
	"errors"
	"net/http"
	"prism/api"
	"prism/services/auth"
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
	r.Post("/add-user", WrapRestHandler(s.AddLicenseUser))
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

func (s *LicenseService) AddLicenseUser(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	params, err := ParseRequestBody[api.AddLicenseUserRequest](r)
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	if err := s.db.Transaction(func(txn *gorm.DB) error {
		return licensing.AddLicenseUser(txn, params.License, userId)
	}); err != nil {
		switch {
		case errors.Is(err, licensing.ErrLicenseNotFound):
			return nil, CodedError(err, http.StatusNotFound)
		case errors.Is(err, licensing.ErrInvalidLicense):
			return nil, CodedError(err, http.StatusUnprocessableEntity)
		case errors.Is(err, licensing.ErrExpiredLicense):
			return nil, CodedError(err, http.StatusForbidden)
		case errors.Is(err, licensing.ErrDeactivatedLicense):
			return nil, CodedError(err, http.StatusForbidden)
		default:
			return nil, CodedError(err, http.StatusInternalServerError)
		}
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
