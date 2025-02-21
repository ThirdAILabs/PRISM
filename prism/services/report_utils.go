package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"prism/prism/api"
	"strings"
	"time"

	"github.com/bbalet/stopwords"
	"github.com/jung-kurt/gofpdf"
	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
)

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
		{"Author Name", report.AuthorName},
		{"Start Year", fmt.Sprintf("%d", report.StartYear)},
		{"End Year", fmt.Sprintf("%d", report.EndYear)},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	if err := writer.Write([]string{}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{"Content Flags", ""}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{"Category", "Summary"}); err != nil {
		return nil, err
	}

	appendFlags := func(flags []api.Flag) error {
		for _, flag := range flags {
			row := []string{flag.GetMessage(), flag.Summary()}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
		return nil
	}

	content := report.Content
	if err := appendFlags(castFlags(content.TalentContracts)); err != nil {
		return nil, err
	}
	if err := appendFlags(castFlags(content.AssociationsWithDeniedEntities)); err != nil {
		return nil, err
	}
	if err := appendFlags(castFlags(content.HighRiskFunders)); err != nil {
		return nil, err
	}
	if err := appendFlags(castFlags(content.AuthorAffiliations)); err != nil {
		return nil, err
	}
	if err := appendFlags(castFlags(content.PotentialAuthorAffiliations)); err != nil {
		return nil, err
	}
	if err := appendFlags(castFlags(content.MiscHighRiskAssociations)); err != nil {
		return nil, err
	}
	if err := appendFlags(castFlags(content.CoauthorAffiliations)); err != nil {
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

func generateExcel(report api.Report) ([]byte, error) {
	f := excelize.NewFile()

	summarySheet := "Summary"
	if err := f.SetSheetName("Sheet1", summarySheet); err != nil {
		return nil, err
	}

	summaryData := [][]interface{}{
		{"Report ID", report.Id.String()},
		{"Created At", report.CreatedAt.Format(time.RFC3339)},
		{"Author Name", report.AuthorName},
		{"Start Year", report.StartYear},
		{"End Year", report.EndYear},
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
	sheetName := flags[0].GetMessage()
	if _, err := f.NewSheet(sheetName); err != nil {
		return err
	}

	headers := []string{"Key", "Message", "Disclosed", "Display Name", "Work URL", "Publication Year", "Raw Acknowledgements"}
	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key()); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.Work.DisplayName); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), flag.Work.WorkUrl); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.Work.PublicationYear); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), strings.Join(flag.RawAcknowledements, ", ")); err != nil {
			return err
		}
	}
	return nil
}

func createSheetForAssociations(f *excelize.File, flags []*api.AssociationWithDeniedEntityFlag) error {
	sheetName := flags[0].GetMessage()
	if _, err := f.NewSheet(sheetName); err != nil {
		return err
	}

	headers := []string{"Key", "Message", "Disclosed", "Display Name", "Work URL", "Publication Year", "Raw Acknowledgements"}
	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key()); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.Work.DisplayName); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), flag.Work.WorkUrl); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.Work.PublicationYear); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), strings.Join(flag.RawAcknowledements, ", ")); err != nil {
			return err
		}

	}
	return nil
}

func createSheetForHighRiskFunders(f *excelize.File, flags []*api.HighRiskFunderFlag) error {
	sheetName := flags[0].GetMessage()
	if _, err := f.NewSheet(sheetName); err != nil {
		return err
	}

	headers := []string{"Key", "Message", "Disclosed", "Display Name", "Work URL", "Publication Year", "Funders", "From Acknowledgements"}
	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key()); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.Work.DisplayName); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), flag.Work.WorkUrl); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.Work.PublicationYear); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), strings.Join(flag.Funders, ", ")); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), flag.FromAcknowledgements); err != nil {
			return err
		}
	}
	return nil
}

func createSheetForAuthorAffiliations(f *excelize.File, flags []*api.AuthorAffiliationFlag) error {
	sheetName := flags[0].GetMessage()
	if _, err := f.NewSheet(sheetName); err != nil {
		return err
	}

	headers := []string{"Key", "Message", "Disclosed", "Display Name", "Work URL", "Publication Year", "Affiliations"}
	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key()); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.Work.DisplayName); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), flag.Work.WorkUrl); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.Work.PublicationYear); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), strings.Join(flag.Affiliations, ", ")); err != nil {
			return err
		}
	}
	return nil
}

func createSheetForPotentialAuthorAffiliations(f *excelize.File, flags []*api.PotentialAuthorAffiliationFlag) error {
	sheetName := flags[0].GetMessage()
	if _, err := f.NewSheet(sheetName); err != nil {
		return err
	}

	headers := []string{"Key", "Message", "Disclosed", "University", "University URL"}
	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key()); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), flag.University); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.UniversityUrl); err != nil {
			return err
		}
	}
	return nil
}

func createSheetForMiscHighRiskAssociations(f *excelize.File, flags []*api.MiscHighRiskAssociationFlag) error {
	sheetName := flags[0].GetMessage()
	if _, err := f.NewSheet(sheetName); err != nil {
		return err
	}

	headers := []string{"Key", "Message", "Disclosed", "Doc Title", "Doc URL", "Doc Entities", "Entity Mentioned", "Frequent Coauthor"}
	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key()); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), flag.DocTitle); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.DocUrl); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), strings.Join(flag.DocEntities, ", ")); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.EntityMentioned); err != nil {
			return err
		}
		if flag.FrequentCoauthor != nil {
			if err := f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), *flag.FrequentCoauthor); err != nil {
				return err
			}
		}
	}
	return nil
}

func createSheetForCoauthorAffiliations(f *excelize.File, flags []*api.CoauthorAffiliationFlag) error {
	sheetName := flags[0].GetMessage()
	if _, err := f.NewSheet(sheetName); err != nil {
		return err
	}

	headers := []string{"Key", "Message", "Disclosed", "Display Name", "Work URL", "Publication Year", "Coauthors", "Affiliations"}
	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	for i, flag := range flags {
		rowNum := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), flag.Key()); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), flag.Message); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), flag.Disclosed); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), flag.Work.DisplayName); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), flag.Work.WorkUrl); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), flag.Work.PublicationYear); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), strings.Join(flag.Coauthors, ", ")); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), strings.Join(flag.Affiliations, ", ")); err != nil {
			return err
		}
	}
	return nil
}

func generatePDF(report api.Report) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "B", 14)
		pdf.SetFillColor(200, 200, 255)
		pdf.CellFormat(0, 10, "Report Details", "0", 1, "C", true, 0, "")
		pdf.Ln(2)
	})

	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	details := [][]string{
		{"Report ID", report.Id.String()},
		{"Created At", report.CreatedAt.Format(time.RFC3339)},
		{"Author Name", report.AuthorName},
		{"Start Year", fmt.Sprintf("%d", report.StartYear)},
		{"End Year", fmt.Sprintf("%d", report.EndYear)},
	}

	for _, row := range details {
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(40, 8, row[0], "1", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 12)
		pdf.CellFormat(0, 8, row[1], "1", 1, "L", false, 0, "")
	}
	pdf.Ln(5)

	addFlagPage := func(flag api.Flag) error {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.SetFillColor(220, 220, 220)
		pdf.CellFormat(0, 10, flag.GetMessage(), "0", 1, "C", true, 0, "")
		pdf.Ln(3)

		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(40, 8, "Summary:", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 12)
		pdf.MultiCell(0, 7, flag.Summary(), "1", "L", false)
		return nil
	}

	content := report.Content

	addFlagPagesForCategory := func(flags []api.Flag) error {
		for _, flag := range flags {
			if err := addFlagPage(flag); err != nil {
				return err
			}
		}
		return nil
	}

	if err := addFlagPagesForCategory(castFlags(content.TalentContracts)); err != nil {
		return nil, err
	}
	if err := addFlagPagesForCategory(castFlags(content.AssociationsWithDeniedEntities)); err != nil {
		return nil, err
	}
	if err := addFlagPagesForCategory(castFlags(content.HighRiskFunders)); err != nil {
		return nil, err
	}
	if err := addFlagPagesForCategory(castFlags(content.AuthorAffiliations)); err != nil {
		return nil, err
	}
	if err := addFlagPagesForCategory(castFlags(content.PotentialAuthorAffiliations)); err != nil {
		return nil, err
	}
	if err := addFlagPagesForCategory(castFlags(content.MiscHighRiskAssociations)); err != nil {
		return nil, err
	}
	if err := addFlagPagesForCategory(castFlags(content.CoauthorAffiliations)); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
