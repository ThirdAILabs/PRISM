package pdf_test

import (
	"fmt"
	"prism/prism/openalex"
	"prism/prism/pdf"
	"strings"
	"testing"

	"github.com/gen2brain/go-fitz"
	"github.com/google/uuid"
)

func readPdf(pdfPath string) (string, error) {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	pageCount := doc.NumPage()

	var textBuilder strings.Builder

	for i := 0; i < pageCount; i++ {
		text, err := doc.Text(i)
		if err != nil {
			return "", fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
		}
		textBuilder.WriteString(fmt.Sprintf("%s\n", text))
	}

	return textBuilder.String(), nil
}

func TestDownloadWithoutCache(t *testing.T) {
	downloader := pdf.NewPDFDownloader("thirdai-prism")

	doi := "nonexistent"
	doiURL := fmt.Sprintf("https://doi.org/%s", doi)

	t.Cleanup(func() {
		if err := downloader.DeleteFromCache(doi); err != nil {
			t.Logf("failed to delete %s from cache: %v", doi, err)
		}
	})

	work := openalex.Work{
		DownloadUrl: "https://arxiv.org/pdf/1706.03762",
		DOI:         doiURL,
	}
	pdfPath, err := downloader.DownloadWork(work)
	if err != nil {
		t.Fatal(err)
	}

	content, err := readPdf(pdfPath)
	if err != nil {
		t.Fatal("Could not extract text from PDF")
	}

	if !strings.Contains(strings.ToLower(content), "attention") {
		t.Fatal("PDF does not contain the expected text: 'attention'")
	}
}

func TestCacheDownload(t *testing.T) {
	downloader := pdf.NewPDFDownloader("thirdai-prism")

	// Check that we can retrieve a PDF from the cache
	work := openalex.Work{
		DownloadUrl: "https://arxiv.org/pdf/1706.03762",
		DOI:         "https://doi.org/test",
	}
	pdfPath, err := downloader.DownloadWork(work)
	if err != nil {
		t.Fatal(err)
	}

	content, err := readPdf(pdfPath)
	if err != nil {
		t.Fatal("Could not extract text from PDF")
	}

	if !strings.Contains(content, "This is a test pdf") {
		t.Fatal("PDF does not contain the expected text: 'This is a test pdf'")
	}
}

func TestCacheUpload(t *testing.T) {
	downloader := pdf.NewPDFDownloader("thirdai-prism")

	doi := fmt.Sprintf("test_upload_%s", uuid.New().String())
	doiURL := fmt.Sprintf("https://doi.org/%s", doi)

	t.Cleanup(func() {
		if err := downloader.DeleteFromCache(doi); err != nil {
			t.Logf("failed to delete %s from cache: %v", doi, err)
		}
	})

	// Check that a PDF is being uploaded to the cache
	work := openalex.Work{
		DownloadUrl: "https://arxiv.org/pdf/1706.03762",
		DOI:         doiURL,
	}

	_, err := downloader.DownloadWork(work)
	if err != nil {
		t.Fatal(err)
	}

	pdfPath, err := downloader.DownloadWork(work)
	if err != nil {
		t.Fatal(err)
	}

	content, err := readPdf(pdfPath)
	if err != nil {
		t.Fatal("Could not extract text from PDF")
	}

	if !strings.Contains(content, "attention") {
		t.Fatal("PDF does not contain the expected text: 'attention'")
	}

	// Check that an existing PDF doesn't get overwritten in the cache
	quantum_work := openalex.Work{
		DownloadUrl: "https://arxiv.org/pdf/1801.00862",
		DOI:         doiURL,
	}

	_, err = downloader.DownloadWork(quantum_work)
	if err != nil {
		t.Fatal(err)
	}

	pdfPath, err = downloader.DownloadWork(quantum_work)
	if err != nil {
		t.Fatal(err)
	}

	content, err = readPdf(pdfPath)
	if err != nil {
		t.Fatal("Could not extract text from PDF")
	}

	if strings.Contains(content, "quantum") {
		t.Fatal("PDF should not contain the text: 'quantum'")
	}

}
