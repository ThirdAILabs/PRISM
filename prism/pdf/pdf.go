package pdf

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"prism/prism/openalex"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/go-resty/resty/v2"
	"github.com/playwright-community/playwright-go"
)

type PDFDownloader struct {
	downloadClient      *resty.Client
	s3CacheBucket       string
	downloadFromS3Cache bool
	uploadToS3Cache     bool
	s3Client            *s3.Client
}

var headers = map[string]string{
	"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
	"accept-language":           "en-US,en;q=0.9",
	"cache-control":             "max-age=0",
	"user-agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
	"upgrade-insecure-requests": "1",
	"sec-ch-ua":                 `"Not/A)Brand";v="99", "Google Chrome";v="115", "Chromium";v="115"`,
	"sec-ch-ua-mobile":          "?0",
	"sec-ch-ua-platform":        `"Windows"`,
}

func NewPDFDownloader(s3CacheBucket string, downloadFromS3Cache, uploadToS3Cache bool) *PDFDownloader {
	downloader := &PDFDownloader{
		downloadClient: resty.New().
			SetRetryCount(1).SetTimeout(20 * time.Second).
			SetRetryWaitTime(5 * time.Second).
			SetRetryMaxWaitTime(30 * time.Second).
			SetHeaders(headers),
		downloadFromS3Cache: downloadFromS3Cache,
		uploadToS3Cache:     uploadToS3Cache,
		s3CacheBucket:       s3CacheBucket,
	}

	if downloadFromS3Cache || uploadToS3Cache {
		if s3CacheBucket == "" {
			log.Fatalf("failed to provide S3 cache bucket")
		}
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("failed to load AWS config: %v", err)
		}
		downloader.s3Client = s3.NewFromConfig(cfg)
	}

	return downloader
}

func (downloader *PDFDownloader) downloadWithPlaywright(url string) (string, error) {
	pw, err := playwright.Run(&playwright.RunOptions{Browsers: []string{"firefox"}})
	if err != nil {
		return "", fmt.Errorf("error starting playwright: %w", err)
	}
	// Skipping error check since there's nothing we can do if this fails
	defer pw.Stop() //nolint:errcheck

	browser, err := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
	if err != nil {
		return "", fmt.Errorf("error launching browser: %w", err)
	}
	defer browser.Close()

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		AcceptDownloads:   playwright.Bool(true),
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("error creating browser context: %w", err)
	}
	defer context.Close()

	// When pdfs fail to download it is often just because they reach the timeout,
	// which slows down processing. Decreasing the timeout will hopefully speed this up.
	context.SetDefaultTimeout(15000)

	page, err := context.NewPage()
	if err != nil {
		return "", fmt.Errorf("error opening browser page: %w", err)
	}
	// context.Close() closes pages in the context

	download, err := page.ExpectDownload(func() error {
		// Page.Goto returns an error saying that the download is starting, so we ignore the error
		page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}) //nolint:errcheck

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("download error: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "tmp-download-*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()

	if err := download.SaveAs(tmpFile.Name()); err != nil {
		return "", fmt.Errorf("error saving downloaded paper: %w", err)
	}

	file, err := os.Open(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("error opening downloaded file for validation: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 4)
	n, err := file.Read(buf)
	if err != nil {
		return "", fmt.Errorf("error reading bytes from downloaded file: %w", err)
	}
	if n < 4 || !bytes.HasPrefix(buf, []byte("%PDF")) {
		return "", fmt.Errorf("download did not return valid pdf")
	}

	return tmpFile.Name(), nil
}

func (downloader *PDFDownloader) downloadWithHttp(url string) (string, error) {
	res, err := downloader.downloadClient.R().
		SetDoNotParseResponse(true).
		Get(url)
	if err != nil {
		return "", fmt.Errorf("download error: %w", err)
	}
	if !res.IsSuccess() {
		return "", fmt.Errorf("download returned error, received status_code=%d", res.StatusCode())
	}

	reader := bufio.NewReader(res.RawBody())
	prefix, err := reader.Peek(4)
	if err != nil {
		return "", fmt.Errorf("failed to read prefix: %w", err)
	}
	if !bytes.HasPrefix(prefix, []byte("%PDF")) {
		return "", fmt.Errorf("download did not return valid pdf")
	}

	tmpFile, err := os.CreateTemp("", "tmp-download-*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, reader); err != nil {
		return "", fmt.Errorf("failed to write data to file: %w", err)
	}

	return tmpFile.Name(), nil
}

func (downloader *PDFDownloader) downloadFromCache(doi string) (string, error) {
	key := fmt.Sprintf("pdfs/%s.pdf", doi)
	input := &s3.GetObjectInput{
		Bucket: aws.String(downloader.s3CacheBucket),
		Key:    aws.String(key),
	}
	resp, err := downloader.s3Client.GetObject(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("cache miss or error retrieving cached file: %w", err)
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "tmp-download-*.pdf")

	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write data to file: %w", err)
	}

	return tmpFile.Name(), nil
}

func (downloader *PDFDownloader) uploadToCache(doi string, pdfPath string) error {
	key := fmt.Sprintf("pdfs/%s.pdf", doi)

	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(downloader.s3CacheBucket),
		Key:    aws.String(key),
	}
	_, err := downloader.s3Client.HeadObject(context.Background(), headInput)
	if err == nil {
		return nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		if apiErr.ErrorCode() != "NotFound" {
			return fmt.Errorf("failed to check if object exists in S3: %w", err)
		}
	} else {
		return fmt.Errorf("failed to check if object exists in S3: %w", err)
	}

	file, err := os.Open(pdfPath)
	if err != nil {
		return fmt.Errorf("failed reading file to upload to S3 cache: %w", err)
	}
	defer file.Close()

	input := &s3.PutObjectInput{
		Bucket: aws.String(downloader.s3CacheBucket),
		Key:    aws.String(key),
		Body:   file,
	}
	_, err = downloader.s3Client.PutObject(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to upload to S3 cache: %w", err)
	}
	return nil
}

func (downloader *PDFDownloader) DownloadWork(work openalex.Work) (string, error) {
	oaURL := work.DownloadUrl
	doi := strings.TrimPrefix(work.DOI, "https://doi.org/")
	var errs []error

	var pdfPath string

	if downloader.downloadFromS3Cache {
		if pdfPath, err := downloader.downloadFromCache(doi); err == nil {
			return pdfPath, nil
		} else {
			fmt.Printf("s3 cache download error: %v", err)
		}
	}

	if attempt1, err1 := downloader.downloadWithHttp(oaURL); attempt1 != "" {
		pdfPath = attempt1
	} else {
		errs = append(errs, fmt.Errorf("http oa error: %w", err1))
	}

	if pdfPath == "" {
		if attempt2, err2 := downloader.downloadWithPlaywright(oaURL); attempt2 != "" {
			pdfPath = attempt2
		} else {
			errs = append(errs, fmt.Errorf("playwright oa error: %w", err2))
		}
	}

	if pdfPath == "" {
		return "", fmt.Errorf("unable to download pdf from %s: %w", oaURL, errors.Join(errs...))
	}

	if downloader.uploadToS3Cache {
		if err := downloader.uploadToCache(doi, pdfPath); err != nil {
			fmt.Printf("warning: failed to upload pdf to S3 cache: %v\n", err)
		}
	}

	return pdfPath, nil

}
