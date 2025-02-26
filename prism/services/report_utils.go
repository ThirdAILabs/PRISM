package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"prism/prism/api"
	"slices"
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
	}

	if err := writer.WriteAll(rows); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{}); err != nil {
		return nil, err
	}

	for _, flags := range report.Content {
		for _, flag := range flags {
			if err := writer.Write([]string{"Flag Title", flag.GetHeading()}); err != nil {
				return nil, err
			}
			for _, kv := range flag.GetDetailFields() {
				if err := writer.Write([]string{kv.Key, kv.Value}); err != nil {
					return nil, err
				}
			}
			if err := writer.Write([]string{"", ""}); err != nil {
				return nil, err
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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

	for _, flags := range report.Content {
		if len(flags) == 0 {
			continue
		}
		groupName := flags[0].GetHeading()
		sheetName := sanitizeSheetName(groupName)
		if _, err := f.NewSheet(sheetName); err != nil {
			return nil, err
		}

		details := flags[0].GetDetailFields()
		headers := make([]string, len(details))
		for i, kv := range details {
			headers[i] = kv.Key
		}

		for i, header := range headers {
			cell, err := excelize.CoordinatesToCellName(i+1, 1)
			if err != nil {
				return nil, err
			}
			if err := f.SetCellValue(sheetName, cell, header); err != nil {
				return nil, err
			}
		}

		for j, flag := range flags {
			rowData := flag.GetDetailFields()
			for i, kv := range rowData {
				cell, err := excelize.CoordinatesToCellName(i+1, j+2)
				if err != nil {
					return nil, err
				}
				if err := f.SetCellValue(sheetName, cell, kv.Value); err != nil {
					return nil, err
				}
			}
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

func sanitizeSheetName(name string) string {
	invalidChars := []string{":", "\\", "/", "?", "*", "[", "]"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "")
	}
	if len(name) > 31 {
		name = name[:31]
	}
	return name
}

func printWatermark(pdf *gofpdf.Fpdf, text string) {
	currR, currG, currB := pdf.GetTextColor()
	pdf.SetFont("Arial", "B", 50)
	pdf.SetAlpha(0.2, "Normal")
	x, y := pdf.GetPageSize()
	pdf.TransformBegin()
	pdf.TransformRotate(45, x/2, y/2)
	pdf.SetTextColor(200, 200, 200)
	pdf.Text(x/4, y/2, text)
	pdf.TransformEnd()

	pdf.SetAlpha(1, "Normal")
	pdf.SetTextColor(currR, currG, currB)
}

func generatePDF(report api.Report) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d of {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AliasNbPages("{nb}")

	pdf.AddPage()
	printWatermark(pdf, "PRISM")
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(200, 200, 255)
	pdf.CellFormat(0, 10, "PRISM REPORT", "0", 1, "C", true, 0, "")
	pdf.Ln(2)

	details := [][]string{
		{"Report ID", report.Id.String()},
		{"Created At", report.CreatedAt.Format(time.RFC3339)},
		{"Author Name", report.AuthorName},
	}
	pdf.SetFont("Arial", "", 12)

	for _, row := range details {
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(40, 8, row[0], "1", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 12)
		pdf.CellFormat(0, 8, row[1], "1", 1, "L", false, 0, "")
	}
	pdf.Ln(5)

	type headingAndFlag struct {
		heading string
		flags   []api.Flag
	}

	var headingsAndFlags []headingAndFlag
	for _, flags := range report.Content {
		if len(flags) > 0 {
			headingsAndFlags = append(headingsAndFlags, headingAndFlag{
				heading: flags[0].GetHeading(),
				flags:   flags,
			})
		}

	}

	slices.SortFunc(headingsAndFlags, func(a, b headingAndFlag) int {
		return strings.Compare(a.heading, b.heading)
	})

	currentPage := 3
	var tocLines []string
	var tocPages []struct {
		Start int
		End   int
	}

	for _, group := range headingsAndFlags {
		if len(group.flags) > 0 {
			start := currentPage
			end := currentPage + len(group.flags) - 1
			tocLines = append(tocLines, fmt.Sprintf("%s: %d - %d", group.heading, start, end))
			tocPages = append(tocPages, struct{ Start, End int }{start, end})
			currentPage += len(group.flags)
		}
	}

	pdf.AddPage()
	printWatermark(pdf, "PRISM")
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Table of Contents", "0", 1, "C", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(120, 8, "Type", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 8, "Pages", "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	for i, line := range tocLines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		category := strings.TrimSpace(parts[0])
		pageRange := strings.TrimSpace(parts[1])
		link := pdf.AddLink()
		pdf.SetLink(link, 0, tocPages[i].Start)
		pdf.SetTextColor(0, 0, 255)
		pdf.CellFormat(120, 8, category, "", 0, "L", false, link, "")
		pdf.CellFormat(0, 8, pageRange, "", 1, "R", false, link, "")
		pdf.SetTextColor(0, 0, 0)
	}

	addFlagPage := func(flag api.Flag) error {
		pdf.AddPage()
		printWatermark(pdf, "PRISM")
		pdf.SetFont("Arial", "B", 14)
		pdf.CellFormat(0, 10, flag.GetHeading(), "0", 1, "C", true, 0, "")
		pdf.Ln(3)
		for _, kv := range flag.GetDetailFields() {
			pdf.SetFont("Arial", "B", 12)
			pdf.CellFormat(41, 8, kv.Key, "", 0, "L", false, 0, "")
			pageWidth, _ := pdf.GetPageSize()
			left, _, right, _ := pdf.GetMargins() // returns left, top, right, bottom
			valueWidth := pageWidth - left - right - 41
			pdf.SetFont("Arial", "", 12)
			if strings.EqualFold(kv.Key, "URL") || strings.HasSuffix(strings.ToLower(kv.Key), "url") {
				pdf.SetTextColor(0, 0, 255)
				pdf.CellFormat(0, 8, "link", "", 1, "L", false, 0, kv.Value)
				pdf.SetTextColor(0, 0, 0)
			} else {
				pdf.MultiCell(valueWidth, 8, kv.Value, "", "L", false)
			}
		}
		pdf.Ln(5)
		return nil
	}

	for _, group := range headingsAndFlags {
		for _, flag := range group.flags {
			if err := addFlagPage(flag); err != nil {
				return nil, err
			}
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
