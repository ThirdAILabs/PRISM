package flaggers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"prism/prism/openalex"
	"prism/prism/pdf"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/semaphore"
)

type AcknowledgementsExtractor interface {
	GetAcknowledgements(logger *slog.Logger, works []openalex.Work) chan CompletedTask[Acknowledgements]
}

type GrobidAcknowledgementsExtractor struct {
	cache        DataCache[Acknowledgements]
	maxThreads   int
	grobidSem    *semaphore.Weighted
	grobidClient *resty.Client
	downloader   *pdf.PDFDownloader
}

func NewGrobidExtractor(cache DataCache[Acknowledgements], grobidEndpoint string, maxDownloadThreads, maxGrobidThreads int, pdfS3CacheBucket string, downloadPDFFromS3Cache, uploadPDFToS3Cache bool) *GrobidAcknowledgementsExtractor {
	return &GrobidAcknowledgementsExtractor{
		cache:      cache,
		maxThreads: max(maxDownloadThreads, maxGrobidThreads),
		grobidSem:  semaphore.NewWeighted(int64(maxGrobidThreads)),
		grobidClient: resty.New().
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
		downloader: pdf.NewPDFDownloader(pdfS3CacheBucket, downloadPDFFromS3Cache, uploadPDFToS3Cache),
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
			return Acknowledgements{}, fmt.Errorf("error extracting acknowledgments for work %s: %w", next.WorkId, err)
		}

		extractor.cache.Update(workId, acks)

		return acks, nil
	}

	nWorkers := min(len(queue), extractor.maxThreads)

	RunInPool(worker, queue, outputCh, nWorkers)

	return outputCh
}

func (extractor *GrobidAcknowledgementsExtractor) extractAcknowledgments(workId string, work openalex.Work) (Acknowledgements, error) {
	pdfPath, err := extractor.downloader.DownloadWork(work)
	if err != nil {
		return Acknowledgements{}, err
	}

	if err := extractor.grobidSem.Acquire(context.Background(), 1); err != nil {
		// I don't think this can fail if we use context.Background, so this error check
		// is just in case.
		slog.Error("error aquiring semaphore for grobid access", "error", err)
		return Acknowledgements{}, fmt.Errorf("error acquiring semaphore for grobid access: %w", err)
	}

	defer extractor.grobidSem.Release(1)

	file, err := os.Open(pdfPath)
	if err != nil {
		return Acknowledgements{}, fmt.Errorf("failed reading file to send to grobid: %w", err)
	}
	defer file.Close()

	acks, err := extractor.processPdfWithGrobid(file)
	if err != nil {
		return Acknowledgements{}, err
	}

	return Acknowledgements{WorkId: workId, Acknowledgements: acks}, nil
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
	res, err := extractor.grobidClient.R().
		SetMultipartField("input", "filename.pdf", "application/pdf", pdf).
		Post("/api/processHeaderFundingDocument")
	if err != nil {
		return nil, fmt.Errorf("error making request to grobid: %w", err)
	}

	if !res.IsSuccess() {
		return nil, fmt.Errorf("grobid returned status=%d, error=%v", res.StatusCode(), res.String())
	}

	body := res.Body()

	return parseGrobidReponse(bytes.NewReader(body))
}
