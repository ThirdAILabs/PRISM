package flaggers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"prism/prism/openalex"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
	"golang.org/x/sync/semaphore"
)

type AcknowledgementsExtractor interface {
	GetAcknowledgements(logger *slog.Logger, works []openalex.Work) chan CompletedTask[Acknowledgements]
}

type GrobidAcknowledgementsExtractor struct {
	cache       DataCache[Acknowledgements]
	maxThreads  int
	grobidSem   *semaphore.Weighted
	grobid      *resty.Client
	downloadDir string
}

func NewGrobidExtractor(cache DataCache[Acknowledgements], grobidEndpoint, downloadDir string) *GrobidAcknowledgementsExtractor {
	return &GrobidAcknowledgementsExtractor{
		cache:      cache,
		maxThreads: 40,
		grobidSem:  semaphore.NewWeighted(int64(10)),
		grobid: resty.New().
			SetBaseURL(grobidEndpoint).
			SetRetryCount(2).
			AddRetryCondition(func(response *resty.Response, err error) bool {
				if err != nil {
					return true // The err can be non nil for some network errors.
				}
				// Grobid returns 503 if there are too many requests:
				// https://grobid.readthedocs.io/en/latest/Grobid-service/
				// TODO: Should we retry on error code 500? grobid will return 500 for invalid pdfs,
				// so it's not clear if this is a retryable error.
				return response != nil && (response.StatusCode() == http.StatusServiceUnavailable)
			}).
			SetRetryWaitTime(2 * time.Second).
			SetRetryMaxWaitTime(10 * time.Second),
		downloadDir: downloadDir,
	}
}

type Entity struct {
	EntityText    string
	EntityType    string
	StartPosition int

	FundCodes []string
}

type Acknowledgement struct {
	RawText            string
	SearchableEntities []Entity
	MiscEntities       []Entity
}

type Acknowledgements struct {
	WorkId           string
	Acknowledgements []Acknowledgement
}

func (extractor *GrobidAcknowledgementsExtractor) GetAcknowledgements(logger *slog.Logger, works []openalex.Work) chan CompletedTask[Acknowledgements] {
	outputCh := make(chan CompletedTask[Acknowledgements], len(works))

	queue := make(chan openalex.Work, len(works))

	for _, work := range works {
		workId := parseOpenAlexId(work)
		if workId == "" {
			continue
		}

		if cachedAck := extractor.cache.Lookup(workId); cachedAck != nil {
			outputCh <- CompletedTask[Acknowledgements]{Result: *cachedAck, Error: nil}
		} else {
			queue <- work
		}

	}
	close(queue)

	worker := func(next openalex.Work) (Acknowledgements, error) {
		workId := parseOpenAlexId(next)

		acks, err := extractor.extractAcknowledgments(workId, next)
		if err != nil {
			return Acknowledgements{}, fmt.Errorf("error extracting acknowledgments for work %s (%s): %w", next.WorkId, next.DownloadUrl, err)
		}

		extractor.cache.Update(workId, acks)

		return acks, nil
	}

	nWorkers := min(len(queue), extractor.maxThreads)

	RunInPool(worker, queue, outputCh, nWorkers)

	return outputCh
}

func (extractor *GrobidAcknowledgementsExtractor) extractAcknowledgments(workId string, work openalex.Work) (Acknowledgements, error) {
	destPath := filepath.Join(extractor.downloadDir, uuid.NewString()+".pdf")
	pdf, err := downloadPdf(work.DownloadUrl, destPath)
	if err != nil {
		return Acknowledgements{}, err
	}

	if closer, ok := pdf.(io.Closer); ok {
		defer closer.Close()
	}

	defer func() {
		if err := os.Remove(destPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Error("error removing temp download file", "error", err)
		}
	}()

	if err := extractor.grobidSem.Acquire(context.Background(), 1); err != nil {
		slog.Error("error aquiring semaphore for grobid access", "error", err)
		return Acknowledgements{}, fmt.Errorf("error acquiring semaphore for grobid access: %w", err)
	}

	defer extractor.grobidSem.Release(1)

	acks, err := extractor.processPdfWithGrobid(pdf)
	if err != nil {
		return Acknowledgements{}, err
	}

	return Acknowledgements{WorkId: workId, Acknowledgements: acks}, nil
}

func downloadWithPlaywright(url, destPath string) (io.ReadCloser, error) {
	pw, err := playwright.Run(&playwright.RunOptions{Browsers: []string{"firefox"}})
	if err != nil {
		return nil, fmt.Errorf("error starting playwright: %w", err)
	}
	// Skipping error check since there's nothing we can do if this fails
	defer pw.Stop() //nolint:errcheck

	browser, err := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
	if err != nil {
		return nil, fmt.Errorf("error launching browser: %w", err)
	}
	defer browser.Close()

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		AcceptDownloads:   playwright.Bool(true),
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating browser context: %w", err)
	}
	defer context.Close()

	// When pdfs fail to download it is often just because they reach the timeout,
	// which slows down processing. Decreasing the timeout will hopefully speed this up.
	context.SetDefaultTimeout(15000)

	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("error opening browser page: %w", err)
	}
	// context.Close() closes pages in the context

	download, err := page.ExpectDownload(func() error {
		// Page.Goto returns an error saying that the download is starting, so we ignore the error
		page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}) //nolint:errcheck

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error downloading pdf '%s': %w", url, err)
	}

	if err := download.SaveAs(destPath); err != nil {
		return nil, fmt.Errorf("error saving downloaded paper: %w", err)
	}

	file, err := os.Open(destPath)
	if err != nil {
		return nil, fmt.Errorf("error opening downloaded paper: %w", err)
	}

	return file, nil
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

var downloadClient = resty.New().SetRetryCount(1).SetTimeout(20 * time.Second).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(30 * time.Second)

func DownloadWithHttp(url string) (io.Reader, error) {
	res, err := downloadClient.R().SetHeaders(headers).Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading pdf from %s: %w", url, err)
	}

	if !res.IsSuccess() {
		return nil, fmt.Errorf("error downloading pdf from %s: recieved status_code=%d, body=%s", url, res.StatusCode(), res.String())
	}

	data := res.Body()

	// Sometimes the above download can succeed, but instead of returning a pdf, it
	// will return the html for the page containing the pdf. This leads to an error
	// when we make a call to grobid. This is a simple trick to check if it is a valid pdf.
	// https://stackoverflow.com/questions/6186980/determine-if-a-byte-is-a-pdf-file
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		return nil, fmt.Errorf("download did not return valid pdf")
	}

	return bytes.NewReader(data), nil
}

func downloadPdf(url, destPath string) (io.Reader, error) {
	attempt1, err1 := DownloadWithHttp(url)
	if err1 != nil {
	} else {
		return attempt1, nil
	}

	attempt2, err2 := downloadWithPlaywright(url, destPath)
	if err2 != nil {
	} else {
		return attempt2, nil
	}

	return nil, fmt.Errorf("unable to download pdf, http error: %w, playwright error: %w", err1, err2)
}

const (
	funderType      = "funder"
	funderNameType  = "funderName"
	grantNameType   = "grantName"
	grantNumberType = "grantNumber"
)

var searchableEntityTypes = map[string]bool{
	funderType:     true,
	funderNameType: true,
	grantNameType:  true,
	"affiliation":  true,
	"institution":  true,
	"programName":  true,
	"projectName":  true,
}

func mergeFundersAndFundCodes(entities []Entity) []Entity {
	merged := make([]Entity, 0)

	for _, entity := range entities {
		entityMerged := false

		if entity.EntityType == grantNameType && len(merged) > 0 {
			last := merged[len(merged)-1]

			if last.EntityType == funderType || last.EntityType == funderNameType {
				last.EntityText += " " + entity.EntityText
				entityMerged = true
			}
		} else if entity.EntityType == grantNumberType && len(merged) > 0 {
			last := merged[len(merged)-1]

			if last.EntityType == funderType || last.EntityType == funderNameType || last.EntityType == grantNameType {
				merged[len(merged)-1].FundCodes = append(merged[len(merged)-1].FundCodes, entity.EntityText)
				entityMerged = true
			}
		}

		if !entityMerged {
			merged = append(merged, entity)
		}
	}

	return merged
}

// Grobid extracts header from acknowledgements (e.g. "Acknowledgments" or "Funding")
// in the extracted text. For now, we remove that header using a regex on the text.
// A better approach would be to fix the root cause in the Grobid response itself,
// but that was giving unexpected results. We will revisit that when time permits.
func cleanAckHeader(raw string) string {
	re := regexp.MustCompile(`(?i)^\s*(acknowledgements|acknowledgments|funding)[:\s-]*`)
	return re.ReplaceAllString(raw, "")
}

func parseGrobidReponse(data io.Reader) ([]Acknowledgement, error) {
	doc, err := goquery.NewDocumentFromReader(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing grobid response: %w", err)
	}

	acks := make([]Acknowledgement, 0)

	processor := func(i int, s *goquery.Selection) {
		text := cleanAckHeader(strings.TrimSpace(s.Text()))

		last := 0

		allEntities := make([]Entity, 0)

		s.Find("rs").Each(func(i int, s *goquery.Selection) {
			entityText := s.Text()
			entityType := s.AttrOr("type", "misc")
			start := strings.Index(text[last:], entityText) + last
			last = start + len(entityText)

			allEntities = append(allEntities, Entity{EntityText: entityText, EntityType: entityType, StartPosition: start})
		})

		searchable := make([]Entity, 0)
		misc := make([]Entity, 0)

		for _, entity := range mergeFundersAndFundCodes(allEntities) {
			if searchableEntityTypes[entity.EntityType] {
				searchable = append(searchable, entity)
			} else {
				misc = append(misc, entity)
			}
		}

		acks = append(acks, Acknowledgement{RawText: text, SearchableEntities: searchable, MiscEntities: misc})
	}

	doc.Find("div[type=acknowledgement]").Each(processor)
	doc.Find("div[type=funding]").Each(processor)

	return acks, nil
}

func (extractor *GrobidAcknowledgementsExtractor) processPdfWithGrobid(pdf io.Reader) ([]Acknowledgement, error) {
	res, err := extractor.grobid.R().
		SetMultipartField("input", "filename.pdf", "application/pdf", pdf).
		Post("/api/processHeaderFundingDocument")
	if err != nil {
		return nil, fmt.Errorf("error making request to grobid: %w", err)
	}

	if !res.IsSuccess() {
		return nil, fmt.Errorf("grobid '%s' returned status=%d, error=%v", res.Request.URL, res.StatusCode(), res.String())
	}

	body := res.Body()

	return parseGrobidReponse(bytes.NewReader(body))
}
