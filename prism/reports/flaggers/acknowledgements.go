package flaggers

import (
	"bytes"
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
)

type AcknowledgementsExtractor interface {
	GetAcknowledgements(logger *slog.Logger, works []openalex.Work) chan CompletedTask[Acknowledgements]
}

type GrobidAcknowledgementsExtractor struct {
	cache       DataCache[Acknowledgements]
	maxWorkers  int
	grobid      *resty.Client
	downloadDir string
}

func NewGrobidExtractor(cache DataCache[Acknowledgements], grobidEndpoint, downloadDir string) *GrobidAcknowledgementsExtractor {
	return &GrobidAcknowledgementsExtractor{
		cache:      cache,
		maxWorkers: 10,
		grobid: resty.New().
			SetBaseURL(grobidEndpoint).
			AddRetryCondition(func(response *resty.Response, err error) bool {
				if err != nil {
					return true // The err can be non nil for some network errors.
				}
				// There's no reason to retry other 400 requests since the outcome should not change
				return response != nil && (response.StatusCode() > 499 || response.StatusCode() == http.StatusTooManyRequests)
			}).
			SetRetryWaitTime(500 * time.Millisecond).
			SetRetryMaxWaitTime(5 * time.Second),
		downloadDir: downloadDir,
	}
}

type Entity struct {
	EntityText    string
	EntityType    string
	StartPosition int
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
			return Acknowledgements{}, fmt.Errorf("error extracting acknowledgments for work %s: %w", next.WorkId, err)
		}

		extractor.cache.Update(workId, acks)

		return acks, nil
	}

	nWorkers := min(len(queue), extractor.maxWorkers)

	RunInPool(worker, queue, outputCh, nWorkers)

	return outputCh
}

func (extractor *GrobidAcknowledgementsExtractor) extractAcknowledgments(workId string, work openalex.Work) (Acknowledgements, error) {
	destPath := filepath.Join(extractor.downloadDir, uuid.NewString()+".pdf")
	pdf, err := downloadPdf(work.DownloadUrl, destPath)
	if err != nil {
		return Acknowledgements{}, err
	}
	defer pdf.Close()

	defer func() {
		if err := os.Remove(destPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Error("error removing temp download file", "error", err)
		}
	}()

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

func downloadWithHttp(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", strings.Replace(url, " ", "%20", -1), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating http request: %w", err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error downloading pdf: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading pdf: recieved status_code=%d", res.StatusCode)
	}

	return res.Body, nil
}

func downloadPdf(url, destPath string) (io.ReadCloser, error) {
	attempt1, err1 := downloadWithHttp(url)
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

var searchAbleEntityTypes = map[string]bool{
	"affiliation": true,
	"funderName":  true,
	"grantName":   true,
	"institution": true,
	"programName": true,
	"projectName": true,
	"funder":      true,
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
		text := strings.TrimSpace(s.Text())
		text = cleanAckHeader(text)

		searchable := make([]Entity, 0)
		misc := make([]Entity, 0)

		last := 0

		s.Find("rs").Each(func(i int, s *goquery.Selection) {
			entityText := s.Text()
			entityType := s.AttrOr("type", "misc")
			start := strings.Index(text[last:], entityText) + last
			last = start + len(entityText)

			if searchAbleEntityTypes[entityType] {
				searchable = append(searchable, Entity{EntityText: entityText, EntityType: entityType, StartPosition: start})
			} else {
				misc = append(misc, Entity{EntityText: entityText, EntityType: entityType, StartPosition: start})
			}
		})

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
