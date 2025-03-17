package pdf_test

import (
	"bytes"
	"fmt"
	"os/exec"
	"prism/prism/openalex"
	"prism/prism/pdf"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func readPdf(pdfPath string) (string, error) {
	cmd := exec.Command("pdftotext", pdfPath, "-")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func TestDownloadWithoutCache(t *testing.T) {
	downloader := pdf.NewPDFDownloader("", false, false)

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

	if !strings.Contains(strings.ToLower(content), "attention") {
		t.Fatal("PDF does not contain the expected text: 'attention'")
	}
}

func TestCacheDownload(t *testing.T) {
	downloader := pdf.NewPDFDownloader("thirdai-prism", true, false)

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
	upload_cache_downloader := pdf.NewPDFDownloader("thirdai-prism", false, true)
	download_cache_downloader := pdf.NewPDFDownloader("thirdai-prism", true, false)

	DOI := fmt.Sprintf("https://doi.org/test_upload_%s", uuid.New().String())

	t.Cleanup(func() { upload_cache_downloader.DeleteFromCache(DOI) })

	// Check that a PDF is being uploaded to the cache
	work := openalex.Work{
		DownloadUrl: "https://arxiv.org/pdf/1706.03762",
		DOI:         DOI,
	}

	_, err := upload_cache_downloader.DownloadWork(work)
	if err != nil {
		t.Fatal(err)
	}

	pdfPath, err := download_cache_downloader.DownloadWork(work)
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
		DOI:         DOI,
	}

	_, err = upload_cache_downloader.DownloadWork(quantum_work)
	if err != nil {
		t.Fatal(err)
	}

	pdfPath, err = download_cache_downloader.DownloadWork(quantum_work)
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
