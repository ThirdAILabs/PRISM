package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/reports"
	"prism/prism/schema"
	"prism/prism/services/auth"
	"prism/prism/services/licensing"
	"strings"
	"time"

	"github.com/bbalet/stopwords"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ReportService struct {
	manager *reports.ReportManager
	db      *gorm.DB
}

type FileResponse struct {
	Content     []byte
	ContentType string
	Filename    string
}

func (s *ReportService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/list", WrapRestHandler(s.List))
	r.Post("/create", WrapRestHandler(s.CreateReport))
	r.Get("/{report_id}", WrapRestHandler(s.GetReport))
	r.Post("/{report_id}/check-disclosure", WrapRestHandler(s.CheckDisclosure))
	r.Get("/{report_id}/download", WrapRestHandler(s.DownloadReport))

	r.Post("/activate-license", WrapRestHandler(s.UseLicense))

	return r
}

func (s *ReportService) List(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	reports, err := s.manager.ListReports(userId)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return reports, nil
}

func (s *ReportService) CreateReport(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	params, err := ParseRequestBody[api.CreateReportRequest](r)
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	if params.AuthorId == "" {
		return nil, CodedError(errors.New("AuthorId must be specified"), http.StatusUnprocessableEntity)
	}

	if params.AuthorId == "" {
		return nil, CodedError(errors.New("AuthorName must be specified"), http.StatusUnprocessableEntity)
	}

	switch params.Source {
	case api.OpenAlexSource, api.GoogleScholarSource, api.ScopusSource, api.UnstructuredSource:
		// ok
	default:
		return nil, CodedError(errors.New("invalid Source"), http.StatusUnprocessableEntity)
	}

	if params.StartYear == 0 {
		params.StartYear = time.Now().Year() - 4
	}

	if params.EndYear == 0 {
		params.EndYear = time.Now().Year()
	}

	licenseId, err := licensing.VerifyLicenseForReport(s.db, userId)
	if err != nil {
		slog.Error("cannot create new report, unable to verify license", "error", err)
		switch {
		case errors.Is(err, licensing.ErrMissingLicense):
			return nil, CodedError(err, http.StatusUnprocessableEntity)
		case errors.Is(err, licensing.ErrExpiredLicense):
			return nil, CodedError(err, http.StatusForbidden)
		case errors.Is(err, licensing.ErrDeactivatedLicense):
			return nil, CodedError(err, http.StatusForbidden)
		default:
			return nil, CodedError(err, http.StatusInternalServerError)
		}
	}

	id, err := s.manager.CreateReport(licenseId, userId, params.AuthorId, params.AuthorName, params.Source, params.StartYear, params.EndYear)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return api.CreateReportResponse{Id: id}, nil
}

func (s *ReportService) GetReport(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	param := chi.URLParam(r, "report_id")
	id, err := uuid.Parse(param)
	if err != nil {
		return nil, CodedError(fmt.Errorf("invalid uuid '%v' provided: %w", param, err), http.StatusBadRequest)
	}

	report, err := s.manager.GetReport(userId, id)
	if err != nil {
		switch {
		case errors.Is(err, reports.ErrReportNotFound):
			return nil, CodedError(err, http.StatusNotFound)
		case errors.Is(err, reports.ErrUserCannotAccessReport):
			return nil, CodedError(err, http.StatusForbidden)
		default:
			return nil, CodedError(err, http.StatusInternalServerError)
		}
	}

	return report, nil
}

func (s *ReportService) UseLicense(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	params, err := ParseRequestBody[api.ActivateLicenseRequest](r)
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

func updateFlagDisclosure(flag api.Flag, allFileTexts []string) {
	entities := filterTokens(flag.GetEntities())

DisclosureCheck:
	for _, entity := range entities {
		for _, txt := range allFileTexts {
			if strings.Contains(strings.ToLower(txt), entity) {
				flag.MarkDisclosed()
				break DisclosureCheck
			}
		}
	}
}

func updateDisclosures[T api.Flag](flags []T, allTexts []string) {
	for _, flag := range flags {
		updateFlagDisclosure(flag, allTexts)
	}
}

func (s *ReportService) CheckDisclosure(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	reportIdParam := chi.URLParam(r, "report_id")
	reportId, err := uuid.Parse(reportIdParam)
	if err != nil {
		return nil, CodedError(fmt.Errorf("invalid report id '%v': %w", reportIdParam, err), http.StatusBadRequest)
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	fileHeaders := r.MultipartForm.File["files"]
	if len(fileHeaders) == 0 {
		return nil, CodedError(errors.New("no files uploaded"), http.StatusBadRequest)
	}

	var allFileTexts []string
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			slog.Error("error opening uploaded file", "filename", fileHeader.Filename, "error", err)
			continue
		}

		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			slog.Error("error reading file bytes", "filename", fileHeader.Filename, "error", err)
			continue
		}

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		text, err := parseFileContent(ext, fileBytes)
		if err != nil {
			slog.Error("error parsing file content", "filename", fileHeader.Filename, "error", err)
			continue
		}

		allFileTexts = append(allFileTexts, text)
	}

	report, err := s.manager.GetReport(userId, reportId)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	if report.Status != schema.ReportCompleted {
		return nil, CodedError(errors.New("cannot process disclosures for report unless report status is complete"), http.StatusUnprocessableEntity)
	}

	updateDisclosures(report.Content.TalentContracts, allFileTexts)
	updateDisclosures(report.Content.AssociationsWithDeniedEntities, allFileTexts)
	updateDisclosures(report.Content.HighRiskFunders, allFileTexts)
	updateDisclosures(report.Content.AuthorAffiliations, allFileTexts)
	updateDisclosures(report.Content.PotentialAuthorAffiliations, allFileTexts)
	updateDisclosures(report.Content.MiscHighRiskAssociations, allFileTexts)
	updateDisclosures(report.Content.CoauthorAffiliations, allFileTexts)

	updatedContentBytes, err := json.Marshal(report.Content)
	if err != nil {
		slog.Error("error serializing updated report content", "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	if err := s.manager.UpdateReport(report.Id, "complete", updatedContentBytes); err != nil {
		slog.Error("error updating report with disclosure information", "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return report, nil
}

func parseFileContent(ext string, fileBytes []byte) (string, error) {
	switch ext {
	case ".txt":
		return string(fileBytes), nil
	case ".pdf":
		return extractTextFromPDF(fileBytes)
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func extractTextFromPDF(fileBytes []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return "", err
	}

	var textBuilder strings.Builder
	numPages := reader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		textBuilder.WriteString(pageText)
	}

	return textBuilder.String(), nil
}

func removeStopwords(text string) string {
	return stopwords.CleanString(text, "en", false)
}

func filterTokens(tokens []string) []string {
	var filtered []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		// If the token contains spaces, assume it’s a multi-word phrase and keep it as-is.
		if strings.Contains(token, " ") {
			filtered = append(filtered, strings.ToLower(token))
		} else {
			cleaned := removeStopwords(token)
			if cleaned != "" {
				filtered = append(filtered, cleaned)
			}
		}
	}
	return filtered
}

func (s *ReportService) DownloadReport(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	reportIdParam := chi.URLParam(r, "report_id")
	reportId, err := uuid.Parse(reportIdParam)
	if err != nil {
		return nil, CodedError(fmt.Errorf("invalid report id '%v': %w", reportIdParam, err), http.StatusBadRequest)
	}

	report, err := s.manager.GetReport(userId, reportId)
	if err != nil {
		switch {
		case errors.Is(err, reports.ErrReportNotFound):
			return nil, CodedError(err, http.StatusNotFound)
		case errors.Is(err, reports.ErrUserCannotAccessReport):
			return nil, CodedError(err, http.StatusForbidden)
		default:
			return nil, CodedError(err, http.StatusInternalServerError)
		}
	}

	if report.Status != schema.ReportCompleted {
		return nil, CodedError(errors.New("cannot download report unless report status is complete"), http.StatusUnprocessableEntity)
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	var fileBytes []byte
	var contentType, filename string
	switch format {
	case "csv":
		fileBytes, err = generateCSV(report)
		if err != nil {
			return nil, CodedError(err, http.StatusInternalServerError)
		}
		contentType = "text/csv"
		filename = "report.csv"
	case "pdf":
		fileBytes, err = generatePDF(report)
		if err != nil {
			return nil, CodedError(err, http.StatusInternalServerError)
		}
		contentType = "application/pdf"
		filename = "report.pdf"
	case "excel", "xlsx":
		fileBytes, err = generateExcel(report)
		if err != nil {
			return nil, CodedError(err, http.StatusInternalServerError)
		}
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "report.xlsx"
	default:
		return nil, CodedError(errors.New("unsupported format"), http.StatusBadRequest)
	}

	return FileResponse{
		Content:     fileBytes,
		ContentType: contentType,
		Filename:    filename,
	}, nil
}

func generateCSV(report api.Report) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	header := []string{"Field", "Value"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	rows := [][]string{
		{"Report ID", report.Id.String()},
		{"Created At", report.CreatedAt.Format(time.RFC3339)},
		{"Author ID", report.AuthorId},
		{"Author Name", report.AuthorName},
		{"Source", report.Source},
		{"Start Year", fmt.Sprintf("%d", report.StartYear)},
		{"End Year", fmt.Sprintf("%d", report.EndYear)},
		{"Status", report.Status},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	if err := writer.Write([]string{}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{"Content Flags", "", ""}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{"Category", "Flag Key", "Details (JSON)"}); err != nil {
		return nil, err
	}

	appendFlags := func(category string, flags []api.Flag) error {
		for _, flag := range flags {
			flagJSON, err := json.Marshal(flag)
			if err != nil {
				return err
			}
			row := []string{category, flag.Key(), string(flagJSON)}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
		return nil
	}

	content := report.Content
	if err := appendFlags("TalentContracts", castFlags(content.TalentContracts)); err != nil {
		return nil, err
	}
	if err := appendFlags("AssociationsWithDeniedEntities", castFlags(content.AssociationsWithDeniedEntities)); err != nil {
		return nil, err
	}
	if err := appendFlags("HighRiskFunders", castFlags(content.HighRiskFunders)); err != nil {
		return nil, err
	}
	if err := appendFlags("AuthorAffiliations", castFlags(content.AuthorAffiliations)); err != nil {
		return nil, err
	}
	if err := appendFlags("PotentialAuthorAffiliations", castFlags(content.PotentialAuthorAffiliations)); err != nil {
		return nil, err
	}
	if err := appendFlags("MiscHighRiskAssociations", castFlags(content.MiscHighRiskAssociations)); err != nil {
		return nil, err
	}
	if err := appendFlags("CoauthorAffiliations", castFlags(content.CoauthorAffiliations)); err != nil {
		return nil, err
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func castFlags[T api.Flag](flags []T) []api.Flag {
	result := make([]api.Flag, len(flags))
	for i, flag := range flags {
		result[i] = flag
	}
	return result
}

func generatePDF(report api.Report) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	pdf.Cell(40, 10, "Report Details")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	addLine := func(field, value string) {
		pdf.CellFormat(40, 10, fmt.Sprintf("%s:", field), "", 0, "", false, 0, "")
		pdf.CellFormat(0, 10, value, "", 1, "", false, 0, "")
	}

	addLine("Report ID", report.Id.String())
	addLine("Created At", report.CreatedAt.Format(time.RFC3339))
	addLine("Author ID", report.AuthorId)
	addLine("Author Name", report.AuthorName)
	addLine("Source", report.Source)
	addLine("Start Year", fmt.Sprintf("%d", report.StartYear))
	addLine("End Year", fmt.Sprintf("%d", report.EndYear))
	addLine("Status", report.Status)

	printFlagPage := func(category string, flag api.Flag) error {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.CellFormat(0, 10, fmt.Sprintf("Content Flag (%s)", category), "", 1, "C", false, 0, "")
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 12)
		pdf.CellFormat(40, 10, "Flag Key:", "", 0, "", false, 0, "")
		pdf.CellFormat(0, 10, flag.Key(), "", 1, "", false, 0, "")

		flagJSON, err := json.Marshal(flag)
		if err != nil {
			return err
		}

		pdf.MultiCell(0, 10, fmt.Sprintf("Details (JSON): %s", string(flagJSON)), "", "", false)
		return nil
	}

	content := report.Content
	for _, flag := range castFlags(content.TalentContracts) {
		if err := printFlagPage("TalentContracts", flag); err != nil {
			return nil, err
		}
	}
	for _, flag := range castFlags(content.AssociationsWithDeniedEntities) {
		if err := printFlagPage("AssociationsWithDeniedEntities", flag); err != nil {
			return nil, err
		}
	}
	for _, flag := range castFlags(content.HighRiskFunders) {
		if err := printFlagPage("HighRiskFunders", flag); err != nil {
			return nil, err
		}
	}
	for _, flag := range castFlags(content.AuthorAffiliations) {
		if err := printFlagPage("AuthorAffiliations", flag); err != nil {
			return nil, err
		}
	}
	for _, flag := range castFlags(content.PotentialAuthorAffiliations) {
		if err := printFlagPage("PotentialAuthorAffiliations", flag); err != nil {
			return nil, err
		}
	}
	for _, flag := range castFlags(content.MiscHighRiskAssociations) {
		if err := printFlagPage("MiscHighRiskAssociations", flag); err != nil {
			return nil, err
		}
	}
	for _, flag := range castFlags(content.CoauthorAffiliations) {
		if err := printFlagPage("CoauthorAffiliations", flag); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func generateExcel(report api.Report) ([]byte, error) {
	f := excelize.NewFile()

	summarySheet := "Summary"
	f.SetSheetName("Sheet1", summarySheet)

	summaryData := [][]interface{}{
		{"Report ID", report.Id.String()},
		{"Created At", report.CreatedAt.Format(time.RFC3339)},
		{"Author ID", report.AuthorId},
		{"Author Name", report.AuthorName},
		{"Source", report.Source},
		{"Start Year", report.StartYear},
		{"End Year", report.EndYear},
		{"Status", report.Status},
	}

	for i, row := range summaryData {
		rowIndex := i + 1
		if err := f.SetCellValue(summarySheet, fmt.Sprintf("A%d", rowIndex), row[0]); err != nil {
			return nil, err
		}
		if err := f.SetCellValue(summarySheet, fmt.Sprintf("B%d", rowIndex), row[1]); err != nil {
			return nil, err
		}
	}

	if len(report.Content.TalentContracts) > 0 {
		if err := createSheetForTalentContracts(f, report.Content.TalentContracts); err != nil {
			return nil, err
		}
	}
	if len(report.Content.AssociationsWithDeniedEntities) > 0 {
		if err := createSheetForAssociations(f, report.Content.AssociationsWithDeniedEntities); err != nil {
			return nil, err
		}
	}
	if len(report.Content.HighRiskFunders) > 0 {
		if err := createSheetForHighRiskFunders(f, report.Content.HighRiskFunders); err != nil {
			return nil, err
		}
	}
	if len(report.Content.AuthorAffiliations) > 0 {
		if err := createSheetForAuthorAffiliations(f, report.Content.AuthorAffiliations); err != nil {
			return nil, err
		}
	}
	if len(report.Content.PotentialAuthorAffiliations) > 0 {
		if err := createSheetForPotentialAuthorAffiliations(f, report.Content.PotentialAuthorAffiliations); err != nil {
			return nil, err
		}
	}
	if len(report.Content.MiscHighRiskAssociations) > 0 {
		if err := createSheetForMiscHighRiskAssociations(f, report.Content.MiscHighRiskAssociations); err != nil {
			return nil, err
		}
	}
	if len(report.Content.CoauthorAffiliations) > 0 {
		if err := createSheetForCoauthorAffiliations(f, report.Content.CoauthorAffiliations); err != nil {
			return nil, err
		}
	}

	index, err := f.GetSheetIndex(summarySheet)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func createSheetForTalentContracts(f *excelize.File, flags []*api.TalentContractFlag) error {
	sheetName := "Talent Contracts"
	f.NewSheet(sheetName)

	headers := []string{"Key", "Message", "Disclosed", "Work ID", "Display Name", "Work URL", "Publication Year", "Raw Acknowledgements"}
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key())
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), flag.Work.WorkId)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.Work.DisplayName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), flag.Work.WorkUrl)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.Work.PublicationYear)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), strings.Join(flag.RawAcknowledements, ", "))
	}
	return nil
}

func createSheetForAssociations(f *excelize.File, flags []*api.AssociationWithDeniedEntityFlag) error {
	sheetName := "Associations With Denied Entities"
	f.NewSheet(sheetName)

	headers := []string{"Key", "Message", "Disclosed", "Work ID", "Display Name", "Work URL", "Publication Year", "Raw Acknowledgements"}
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key())
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), flag.Work.WorkId)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.Work.DisplayName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), flag.Work.WorkUrl)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.Work.PublicationYear)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), strings.Join(flag.RawAcknowledements, ", "))
	}
	return nil
}

func createSheetForHighRiskFunders(f *excelize.File, flags []*api.HighRiskFunderFlag) error {
	sheetName := "High Risk Funders"
	f.NewSheet(sheetName)

	headers := []string{"Key", "Message", "Disclosed", "Work ID", "Display Name", "Work URL", "Publication Year", "Funders", "From Acknowledgements"}
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key())
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), flag.Work.WorkId)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.Work.DisplayName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), flag.Work.WorkUrl)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.Work.PublicationYear)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), strings.Join(flag.Funders, ", "))
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), flag.FromAcknowledgements)
	}
	return nil
}

func createSheetForAuthorAffiliations(f *excelize.File, flags []*api.AuthorAffiliationFlag) error {
	sheetName := "Author Affiliations"
	f.NewSheet(sheetName)

	headers := []string{"Key", "Message", "Disclosed", "Work ID", "Display Name", "Work URL", "Publication Year", "Affiliations"}
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key())
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), flag.Work.WorkId)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.Work.DisplayName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), flag.Work.WorkUrl)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.Work.PublicationYear)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), strings.Join(flag.Affiliations, ", "))
	}
	return nil
}

func createSheetForPotentialAuthorAffiliations(f *excelize.File, flags []*api.PotentialAuthorAffiliationFlag) error {
	sheetName := "Potential Author Affiliations"
	f.NewSheet(sheetName)

	headers := []string{"Key", "Message", "Disclosed", "University", "University URL"}
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key())
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), flag.University)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.UniversityUrl)
	}
	return nil
}

func createSheetForMiscHighRiskAssociations(f *excelize.File, flags []*api.MiscHighRiskAssociationFlag) error {
	sheetName := "Misc High Risk Associations"
	f.NewSheet(sheetName)

	headers := []string{"Key", "Message", "Disclosed", "Doc Title", "Doc URL", "Doc Entities", "Entity Mentioned", "Frequent Coauthor"}
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key())
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), flag.DocTitle)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.DocUrl)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), strings.Join(flag.DocEntities, ", "))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.EntityMentioned)
		if flag.FrequentCoauthor != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), *flag.FrequentCoauthor)
		}
	}
	return nil
}

func createSheetForCoauthorAffiliations(f *excelize.File, flags []*api.CoauthorAffiliationFlag) error {
	sheetName := "Coauthor Affiliations"
	f.NewSheet(sheetName)

	headers := []string{"Key", "Message", "Disclosed", "Work ID", "Display Name", "Work URL", "Publication Year", "Coauthors", "Affiliations"}
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key())
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), flag.Work.WorkId)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.Work.DisplayName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), flag.Work.WorkUrl)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.Work.PublicationYear)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), strings.Join(flag.Coauthors, ", "))
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), strings.Join(flag.Affiliations, ", "))
	}
	return nil
}
