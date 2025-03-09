package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
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
		{"Downloaded At", time.Now().Format(time.RFC3339)},
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
		{"Downloaded At", time.Now().Format(time.RFC3339)},
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

	// Get text width to calculate center position
	textWidth := pdf.GetStringWidth(text)

	pdf.TransformBegin()
	pdf.TransformRotate(45, x/2, y/2)
	pdf.SetTextColor(200, 200, 200)
	// Position text centered on page
	pdf.Text(x/2-textWidth/2, y/2, text)
	pdf.TransformEnd()

	pdf.SetAlpha(1, "Normal")
	pdf.SetTextColor(currR, currG, currB)
}

func setupPDFHeader(pdf *gofpdf.Fpdf, resourceFolder string, authorName string) {
	pdf.SetHeaderFunc(func() {
		currentR, currentG, currentB := pdf.GetTextColor()
		pageWidth, _ := pdf.GetPageSize()

		// add smaller logo to header
		logoWidth := 10.0
		pdf.Image(filepath.Join(resourceFolder, "prism-header-logo.png"), 20, 10, logoWidth, 0, false, "", 0, "")

		// prism report text
		pdf.SetFont("Arial", "B", 12)
		pdf.SetTextColor(0, 0, 200)
		pdf.Text(pageWidth-pdf.GetStringWidth(fmt.Sprintf("%s Report", authorName))-20, 17, fmt.Sprintf("%s Report", authorName))

		// divider line
		pdf.SetDrawColor(200, 200, 200)
		pdf.Line(20, 25, pageWidth-20, 25)

		// watermark
		printWatermark(pdf, "PRISM")

		// restore original settings
		// assume we're using Arial 14 font
		pdf.SetFont("Arial", "", 14)
		pdf.SetTextColor(currentR, currentG, currentB)
	})
}

func setupPDFFooter(pdf *gofpdf.Fpdf) {
	pdf.SetFooterFunc(func() {
		currentR, currentG, currentB := pdf.GetTextColor()
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetY(-15)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d of {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
		pdf.SetTextColor(currentR, currentG, currentB)
	})
}

func setupPDFCoverPage(pdf *gofpdf.Fpdf, report api.Report, resourceFolder string) {
	pdf.AddPage()

	// add prism logo to the front page
	logoPath := filepath.Join(resourceFolder, "prism-logo.png")
	logoWidth := 100.0 // Width in mm, adjust as needed

	// Get page width to center the logo
	pageWidth, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usableWidth := pageWidth - left - right

	// Calculate x position to center the logo
	xPos := (usableWidth-logoWidth)/2 + left

	// Add the logo at position (xPos, 20) with width logoWidth
	pdf.Image(logoPath, xPos, 50, logoWidth, 0, false, "", 0, "")
	pdf.Ln(40)

	pdf.SetY(200)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(200, 200, 255)
	pdf.CellFormat(0, 10, "Individual Report", "0", 1, "C", true, 0, "")
	pdf.Ln(2)

	details := [][]string{
		{"Report ID", report.Id.String()},
		{"Downloaded At", time.Now().Format(time.DateOnly)},
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
}

func setupPDFFlagGroup(pdf *gofpdf.Fpdf, flags []api.Flag) error {
	if len(flags) == 0 {
		return nil
	}
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(200, 200, 255)

	pageWidth, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	headerWidth := pageWidth - left - right

	pdf.SetX(left)
	pdf.CellFormat(headerWidth, 10, flags[0].GetHeading(), "0", 1, "C", true, 0, "")
	pdf.Ln(3)

	for flagIndex, flag := range flags {
		pdf.SetFont("Arial", "B", 13)
		pdf.SetFillColor(230, 230, 230)
		pdf.CellFormat(0, 10, fmt.Sprintf("Issue %d", flagIndex+1), "", 1, "L", true, 0, "")
		pdf.Ln(3)

		for _, kv := range flag.GetDetailsFieldsForReport() {
			keyWidth := 50.0
			pageWidth, _ := pdf.GetPageSize()
			left, _, right, _ := pdf.GetMargins()
			valueWidth := pageWidth - left - right - keyWidth

			pdf.SetFont("Arial", "B", 11)
			pdf.SetTextColor(80, 80, 80)
			pdf.CellFormat(keyWidth, 8, kv.Key, "", 0, "L", false, 0, "")

			pdf.SetFont("Arial", "", 11)
			pdf.SetTextColor(0, 0, 0)

			if kv.Url != "" {
				pdf.SetTextColor(0, 0, 200)
				startX := pdf.GetX()
				startY := pdf.GetY()
				pdf.MultiCell(valueWidth, 8, kv.Value, "", "L", false)
				pdf.LinkString(startX, startY, valueWidth, pdf.GetY()-startY, kv.Url)
				pdf.SetTextColor(0, 0, 0)
			} else {
				pdf.MultiCell(valueWidth, 8, kv.Value, "", "L", false)
			}
			pdf.Ln(2)
		}
		pdf.Ln(5)
	}
	return nil
}

func generatePDF(report api.Report, resourceFolder string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 30, 20)
	pdf.SetAutoPageBreak(true, 20)

	setupPDFCoverPage(pdf, report, resourceFolder)

	// we set the footer here so that cover also has a page number
	setupPDFFooter(pdf)

	pdf.AliasNbPages("{nb}")

	// Prepare heading and flags data
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

	// Reserve page for TOC (page 2)
	// we set the header after the cover page so that header logo starts from page 2
	setupPDFHeader(pdf, resourceFolder, report.AuthorName)
	pdf.AddPage()
	tocPage := pdf.PageNo()

	// Track section start pages
	type sectionInfo struct {
		heading   string
		startPage int
	}
	var sectionPages []sectionInfo

	// Generate all content pages
	for _, group := range headingsAndFlags {
		if len(group.flags) > 0 {
			startPage := pdf.PageNo()
			sectionPages = append(sectionPages, sectionInfo{
				heading:   group.flags[0].GetHeading(),
				startPage: startPage,
			})

			if err := setupPDFFlagGroup(pdf, group.flags); err != nil {
				return nil, err
			}
		}
	}

	// Now create TOC with correct page references
	// Important: We need to save the last page number before going back to TOC
	lastPageBeforeTOC := pdf.PageNo()

	// Go back to TOC page and fill it in
	pdf.SetPage(tocPage)
	pdf.SetY(30)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Table of Contents", "0", 1, "C", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(120, 8, "Category", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 8, "Page", "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "", 12)

	for _, section := range sectionPages {
		link := pdf.AddLink()
		pdf.SetLink(link, 0, section.startPage+1)
		pdf.SetTextColor(0, 0, 255)
		pdf.CellFormat(120, 8, section.heading, "", 0, "L", false, link, "")
		pdf.CellFormat(0, 8, fmt.Sprintf("%d", section.startPage+1), "", 1, "R", false, link, "")
		pdf.SetTextColor(0, 0, 0)
	}

	// Critical fix: Return to the last content page to ensure all pages are included
	// without this, content pages might be lost
	pdf.SetPage(lastPageBeforeTOC)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
