package flaggers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"prism/openalex"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
	"go.etcd.io/bbolt"
)

var cacheBucket = []byte("ack-cache")

type AcknowledgementsExtractor struct {
	cache          *bbolt.DB
	maxWorkers     int
	grobidEndpoint string
	downloadDir    string
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
	OpenAlexId       string
	Acknowledgements []Acknowledgement
}

func (extractor *AcknowledgementsExtractor) GetAcknowledgements(works []openalex.Work) (chan Acknowledgements, chan error) {
	outputCh := make(chan Acknowledgements, len(works))
	errorCh := make(chan error, 10)

	queue := make(chan openalex.Work, len(works))

	for _, work := range works {
		id := parseOpenAlexId(work)
		if id == "" {
			continue
		}

		cachedAck, err := extractor.checkCache(id)
		if err != nil {
			slog.Error("error checking cache for acknowledgement", "id", id, "error", err)
			errorCh <- fmt.Errorf("error checking acknowledgment cache: %w", err)
		} else if cachedAck != nil {
			outputCh <- *cachedAck
			continue
		}

		queue <- work
	}
	close(queue)

	nWorkers := min(len(queue), extractor.maxWorkers)
	for i := 0; i < nWorkers; i++ {
		go extractor.worker(queue, outputCh, errorCh)
	}

	return outputCh, errorCh
}

func (extractor *AcknowledgementsExtractor) worker(queue chan openalex.Work, outputCh chan Acknowledgements, errorCh chan error) {
	for {
		next, done := <-queue
		if done {
			return
		}

		id := parseOpenAlexId(next)

		acks, err := extractor.extractAcknowledgments(id, next)
		if err != nil {
			slog.Error("error extracting acknowledgements for work", "id", id, "name", next.DisplayName, "error", err)
			errorCh <- fmt.Errorf("error extracting acknowledgments: %w", err)
		} else {
			outputCh <- acks
		}

		extractor.updateCache(id, acks)
	}
}

func parseOpenAlexId(work openalex.Work) string {
	idx := strings.LastIndex(work.WorkId, "/")
	if idx < 0 {
		return ""
	}
	return work.WorkId[idx+1:]
}

func (extractor *AcknowledgementsExtractor) checkCache(id string) (*Acknowledgements, error) {
	slog.Info("checking acknowledgement cache", "id", id)
	var data []byte
	err := extractor.cache.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(cacheBucket)

		data = bucket.Get([]byte(id))

		return nil
	})

	if err != nil {
		slog.Error("cache access failed", "id", id, "error", err)
		return nil, fmt.Errorf("cache access failed: %w", err)
	}

	if data == nil {
		slog.Info("no cached acknowledgements found", "id", id)
		return nil, nil
	}

	var acks Acknowledgements
	if err := json.Unmarshal(data, &acks); err != nil {
		slog.Info("error parsing cache data", "id", id, "error", err)
		return nil, fmt.Errorf("error parsing cache data: %w", err)
	}

	slog.Info("found cached acknowledgements", "id", id)

	return &acks, nil
}

func (extractor *AcknowledgementsExtractor) updateCache(id string, acks Acknowledgements) {
	slog.Info("updating acknowledgement cache", "id", id)

	data, err := json.Marshal(acks)
	if err != nil {
		slog.Error("error updating acknowledgements cache: error serializing data", "id", id, "error", err)
		return // No error since cache update isn't critical
	}

	if err := extractor.cache.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(cacheBucket).Put([]byte(id), data)
	}); err != nil {
		slog.Error("error updating acknowledgements cache: bolt db error", "id", id, "error", err)
	}

	slog.Info("successfully updated acknowledgements cache", "id", id)
}

func (extractor *AcknowledgementsExtractor) extractAcknowledgments(id string, work openalex.Work) (Acknowledgements, error) {
	slog.Info("extracting acknowledgments from", "id", work.WorkId, "name", work.DisplayName)

	destPath := filepath.Join(extractor.downloadDir, uuid.NewString()+".pdf")
	pdf, err := downloadPdf(work.DownloadUrl, destPath)
	if err != nil {
		slog.Error("error downloading pdf", "id", work.WorkId, "name", work.DisplayName, "error", err)
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
		slog.Error("error processing pdf with grobid", "id", work.WorkId, "name", work.DisplayName, "error", err)
		return Acknowledgements{}, err
	}

	return Acknowledgements{OpenAlexId: id, Acknowledgements: acks}, nil
}

func downloadWithPlaywright(url, destPath string) (io.ReadCloser, error) {
	pw, err := playwright.Run(&playwright.RunOptions{Browsers: []string{"firefox"}})
	if err != nil {
		return nil, fmt.Errorf("error starting playwright: %w", err)
	}
	defer pw.Stop()

	browser, err := pw.Firefox.Launch()
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

	download, err := page.ExpectDownload(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating download handler: %w", err)
	}

	if _, err := page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
		return nil, fmt.Errorf("error accessing download url: %w", err)
	}

	if err := download.SaveAs(destPath); err != nil {
		return nil, fmt.Errorf("error downloading paper: %w", err)
	}

	file, err := os.Open(destPath)
	if err != nil {
		return nil, fmt.Errorf("error opening downloaded file: %w", err)
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

	return res.Body, nil
}

func downloadPdf(url, destPath string) (io.ReadCloser, error) {
	slog.Info("attempting to download pdf", "url", url)

	attempt1, err1 := downloadWithHttp(url)
	if err1 != nil {
		slog.Error("basic download failed", "url", url, "error", err1)
	} else {
		return attempt1, nil
	}

	attempt2, err2 := downloadWithPlaywright(url, destPath)
	if err2 != nil {
		slog.Error("playwright download failed", "url", url, "error", err2)
	} else {
		return attempt2, nil
	}

	return nil, fmt.Errorf("unable to download pdf:\n%w\n%w", err1, err2)
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

func parseGrobidReponse(data io.Reader) ([]Acknowledgement, error) {
	doc, err := goquery.NewDocumentFromReader(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing grobid response: %w", err)
	}

	acks := make([]Acknowledgement, 0)

	processor := func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())

		searchable := make([]Entity, 0)
		misc := make([]Entity, 0)

		last := 0

		s.Find("rs").Each(func(i int, s *goquery.Selection) {
			entityText := s.Text()
			entityType := s.AttrOr("type", "misc")
			start := strings.Index(text[last:], entityText)
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

func (extractor *AcknowledgementsExtractor) processPdfWithGrobid(pdf io.Reader) ([]Acknowledgement, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("input", "filename.pdf")
	if err != nil {
		return nil, fmt.Errorf("error creating multipart request: %w", err)
	}

	if _, err := io.Copy(part, pdf); err != nil {
		return nil, fmt.Errorf("error copying data to multipart request: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error finalizing multipart request: %w", err)
	}

	// TODO: add backoff/retry
	res, err := http.Post(extractor.grobidEndpoint, writer.FormDataContentType(), body)
	if err != nil {
		return nil, fmt.Errorf("error making request to grobid: %w", err)
	}
	defer res.Body.Close()

	return parseGrobidReponse(res.Body)
}
