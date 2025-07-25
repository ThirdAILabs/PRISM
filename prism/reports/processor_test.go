package reports_test

import (
	"os"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/openalex"
	"prism/prism/reports"
	"prism/prism/reports/flaggers"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/reports/utils"
	"prism/prism/schema"
	"prism/prism/search"
	"prism/prism/triangulation"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	const licensePath = "../../.test_license/thirdai.license"
	if err := search.SetLicensePath(licensePath); err != nil {
		panic(err)
	}
}

func setupReportManager(t *testing.T) *reports.ReportManager {
	db := schema.SetupTestDB(t)

	// We're using old date ranges for these so they only process works around the
	// date of the flagged work, particularly for the acknowledgements flagger, this
	// ensures the tests don't take too long to run. However this means that the reports
	// are viewed as stale and will be reprocessed on the next test. Setting a large
	// threshold here fixes it. In ~100 years this threshold would again be too small,
	// but at that point this code will likely not be in use, or if it is, then it
	// will be someone else's problem (or more likely an AI).
	return reports.NewManager(db).SetAuthorReportUpdateInterval(100 * 365 * 25 * time.Hour)
}

func eqOrderInvariant(a, b []string) bool {
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

func yearStart(year int) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
}

func yearEnd(year int) time.Time {
	return time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
}

func getReportContent(t *testing.T, report reports.ReportUpdateTask, processor *reports.ReportProcessor, manager *reports.ReportManager) map[string][]api.Flag {
	user := uuid.New()
	reportId, err := manager.CreateAuthorReport(user, report.AuthorId, report.AuthorName, report.Source, "", "")
	if err != nil {
		t.Fatal(err)
	}

	nextReport, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	if nextReport == nil {
		t.Fatal("expected next report")
	}

	report.Id = nextReport.Id
	processor.ProcessAuthorReport(report)

	content, err := manager.GetAuthorReport(user, reportId)
	if err != nil {
		t.Fatal(err)
	}

	return content.Content
}

func TestProcessorCoauthorAffiliation(t *testing.T) {
	manager := setupReportManager(t)

	processor := reports.NewProcessor(
		[]reports.WorkFlagger{
			flaggers.NewOpenAlexCoauthorAffiliationIsEOC(eoc.LoadGeneralEOC(), eoc.LoadInstitutionEOC()),
		},
		nil,
		manager,
	)

	t.Run("Case1", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5084836278",
				AuthorName: "Charles M. Lieber",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2019),
				EndDate:    yearEnd(2019),
			},
			processor,
			manager,
		)

		if len(report[api.CoauthorAffiliationType]) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := report[api.CoauthorAffiliationType][0].(*api.CoauthorAffiliationFlag)

		if len(flag.Affiliations) != 1 || flag.Affiliations[0] != "Central South University" || len(flag.Coauthors) != 1 || flag.Coauthors[0] != "Jian Sun" {
			t.Fatal("incorrect flag")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5075113943",
				AuthorName: "Zijian Hong",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2024),
				EndDate:    yearEnd(2024),
			},
			processor,
			manager,
		)

		if len(report[api.CoauthorAffiliationType]) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		expectedAuthors := []string{
			"Yongjun Wu",
			"Congqin Zheng",
			"Yu Huang",
			"Tiantian Chen",
			"Fu Lv",
			"W. Li",
			"Xin Li",
			"Qian Li",
		}

		expectedInstitutions := []string{
			"Tsinghua University",
			"Zhejiang University",
		}

		foundFlag := false
		for _, flag := range report[api.CoauthorAffiliationType] {
			flag := flag.(*api.CoauthorAffiliationFlag)
			if flag.Work.WorkId == "https://openalex.org/W4402273377" {
				foundFlag = true
				if !eqOrderInvariant(flag.Affiliations, expectedInstitutions) || !eqOrderInvariant(flag.Coauthors, expectedAuthors) {
					t.Fatalf("incorrect flag returned, flagged affiliation %v, flagged co-authors %v", flag.Affiliations, flag.Coauthors)
				}
			}
		}
		if !foundFlag {
			t.Fatal("didn't find flag")
		}
	})
}

func TestProcessorAuthorAffiliation(t *testing.T) {
	manager := setupReportManager(t)

	processor := reports.NewProcessor(
		[]reports.WorkFlagger{
			flaggers.NewOpenAlexAuthorAffiliationIsEOC(eoc.LoadGeneralEOC(), eoc.LoadInstitutionEOC()),
		},
		nil,
		manager,
	)

	t.Run("Case1", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5084836278",
				AuthorName: "Charles M. Lieber",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2013),
				EndDate:    yearEnd(2013),
			},
			processor,
			manager,
		)

		if len(report[api.AuthorAffiliationType]) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		found := false
		for _, flag := range report[api.AuthorAffiliationType] {
			flag := flag.(*api.AuthorAffiliationFlag)
			if len(flag.Affiliations) != 1 {
				t.Fatal("incorrect flag")
			}
			if flag.Affiliations[0] == "Peking University" {
				found = true
			}
		}

		if !found {
			t.Fatal("missing flag")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5075113943",
				AuthorName: "Zijian Hong",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2024),
				EndDate:    yearEnd(2024),
			},
			processor,
			manager,
		)

		if len(report[api.AuthorAffiliationType]) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		found := false
		for _, flag := range report[api.AuthorAffiliationType] {
			flag := flag.(*api.AuthorAffiliationFlag)

			if len(flag.Affiliations) != 1 {
				t.Fatal("incorrect flag")
			}
			if flag.Affiliations[0] == "Zhejiang University" {
				found = true
			}
		}

		if !found {
			t.Fatal("missing flag")
		}
	})
}

func TestProcessorUniversityFacultySeach(t *testing.T) {
	universityNDB := flaggers.BuildUniversityNDB("../../data/university_webpages.json", t.TempDir())
	defer universityNDB.Free()

	manager := setupReportManager(t)

	processor := reports.NewProcessor(
		nil,
		[]reports.AuthorFlagger{
			flaggers.NewAuthorIsFacultyAtEOCFlagger(universityNDB),
		},
		manager,
	)

	t.Run("Case1", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5019148940",
				AuthorName: "Natalie Artzi",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2024),
				EndDate:    yearEnd(2024),
			},
			processor,
			manager,
		)

		if len(report[api.PotentialAuthorAffiliationType]) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := report[api.PotentialAuthorAffiliationType][0].(*api.PotentialAuthorAffiliationFlag)

		if flag.University != "Fudan University" || !strings.Contains(flag.UniversityUrl, "fudan.edu") {
			t.Fatal("incorrect flag")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5075113943",
				AuthorName: "Zijian Hong",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2024),
				EndDate:    yearEnd(2024),
			},
			processor,
			manager,
		)

		if len(report[api.PotentialAuthorAffiliationType]) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		found := false
		for _, flag := range report[api.PotentialAuthorAffiliationType] {
			flag := flag.(*api.PotentialAuthorAffiliationFlag)
			if flag.University == "Zhejiang University" && strings.Contains(flag.UniversityUrl, "zju.edu") {
				found = true
			}
		}

		if !found {
			t.Fatal("missing flag")
		}
	})
}

func TestProcessorAuthorAssociations(t *testing.T) {
	manager := setupReportManager(t)

	processor := reports.NewProcessor(
		[]reports.WorkFlagger{
			flaggers.NewAuthorIsAssociatedWithEOCFlagger(
				flaggers.BuildDocIndex("../../data/docs_and_press_releases.json"),
				flaggers.BuildAuxIndex("../../data/auxiliary_webpages.json"),
			),
		},
		nil,
		manager,
	)

	t.Run("PrimaryConnection", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5084836278",
				AuthorName: "Charles M. Lieber",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2013),
				EndDate:    yearEnd(2013),
			},
			processor,
			manager,
		)

		if len(report[api.MiscHighRiskAssociationType]) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		expectedTitles := []string{
			"Former Harvard University Professor Sentenced for Lying About His Affiliation with Wuhan University of Technology; China’s Thousand Talents Program; and Filing False Tax Returns",
			"Harvard University Professor Convicted of Making False Statements and Tax Offenses",
			"Harvard University Professor Indicted on False Statement Charges",
			"Harvard University Professor Charged with Tax Offenses",
			"Harvard University Professor and Two Chinese Nationals Charged in Three Separate China Related Cases",
		}

		titles := make([]string, 0)

		for _, flag := range report[api.MiscHighRiskAssociationType] {
			flag := flag.(*api.MiscHighRiskAssociationFlag)
			if len(flag.Connections) != 0 || flag.EntityMentioned != "Charles M. Lieber" {
				t.Fatal("incorrect flag")
			}
			titles = append(titles, flag.DocTitle)
		}

		if !eqOrderInvariant(expectedTitles, titles) {
			t.Fatal("incorrect docs found")
		}
	})

	t.Run("SecondaryConnection", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5012289937",
				AuthorName: "Anqi Zhang",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2015),
				EndDate:    yearEnd(2020),
			},
			processor,
			manager,
		)

		if len(report[api.MiscHighRiskAssociationType]) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		for _, flag := range report[api.MiscHighRiskAssociationType] {
			flag := flag.(*api.MiscHighRiskAssociationFlag)
			if len(flag.Connections) != 1 || flag.FrequentCoauthor == nil || *flag.FrequentCoauthor != "Charles M. Lieber" {
				t.Fatal("incorrect flag")
			}
		}
	})

	t.Run("TertiaryConnection", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5016320004",
				AuthorName: "David Zhang",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2020),
				EndDate:    yearEnd(2020),
			},
			processor,
			manager,
		)

		if len(report[api.MiscHighRiskAssociationType]) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		entitiesMentioned := map[string]bool{}
		for _, flag := range report[api.MiscHighRiskAssociationType] {
			flag := flag.(*api.MiscHighRiskAssociationFlag)
			if len(flag.Connections) != 2 ||
				flag.Connections[0].DocTitle != "NuProbe About Us" ||
				flag.Connections[0].DocUrl != "https://nuprobe.com/about-us/" ||
				flag.Connections[1].DocTitle != "NuProbe Announces $11 Million Series A Funding Round" ||
				flag.Connections[1].DocUrl != "https://nuprobe.com/2018/04/nuprobe-announces-11-million-series-a-funding-round-2/" {
				t.Fatalf("incorrect flag; %+v", flag)
			}
			entitiesMentioned[flag.EntityMentioned] = true
		}
		if !entitiesMentioned["WuXi AppTec"] || !entitiesMentioned["Sequoia Capital China"] {
			t.Fatal("incorrect entities mentioned")
		}
	})
}

func TestProcessorAcknowledgements(t *testing.T) {
	grobidEndpoint := os.Getenv("GROBID_ENDPOINT")
	if grobidEndpoint == "" {
		t.Skip("No grobid endpoint detected, skipping acknowledgements tests. Set GROBID_ENDPOINT env variable to run these tests")
	}

	testDir := t.TempDir()

	authorCache, err := utils.NewCache[openalex.Author]("authors", filepath.Join(testDir, "authors.cache"))
	if err != nil {
		t.Fatal(err)
	}
	defer authorCache.Close()

	ackCache, err := utils.NewCache[flaggers.Acknowledgements]("acks", filepath.Join(testDir, "acks.cache"))
	if err != nil {
		t.Fatal(err)
	}
	defer ackCache.Close()

	entityStore := flaggers.BuildWatchlistEntityIndex(eoc.LoadSourceToAlias())

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&triangulation.Author{},
		&triangulation.FundCode{},
	); err != nil {
		t.Fatal(err)
	}

	triangulationDB := triangulation.CreateTriangulationDB(db)

	manager := setupReportManager(t)

	processor := reports.NewProcessor(
		[]reports.WorkFlagger{
			flaggers.NewOpenAlexAcknowledgementIsEOC(
				entityStore, authorCache, flaggers.NewGrobidExtractor(ackCache, grobidEndpoint, 40, 10, "thirdai-prism"), eoc.LoadSussyBakas(), triangulationDB,
			),
		},
		nil,
		manager,
	)

	t.Run("Case1", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5100327325",
				AuthorName: "Xin Zhang",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2022),
				EndDate:    yearEnd(2022),
			},
			processor,
			manager,
		)

		if len(report[api.HighRiskFunderType]) < 1 || len(report[api.TalentContractType]) < 1 {
			t.Fatal("expected high risk funders and talent contract flags")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5075113943",
				AuthorName: "Zijian Hong",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2023),
				EndDate:    yearEnd(2023),
			},
			processor,
			manager,
		)

		if len(report[api.HighRiskFunderType]) < 1 {
			t.Fatal("expected >= 1 flags")
		}

		found := false
		for _, flag := range report[api.HighRiskFunderType] {
			flag := flag.(*api.HighRiskFunderFlag)
			if flag.Work.WorkId == "https://openalex.org/W4384197626" {
				found = true
				if len(flag.RawAcknowledgements) == 0 || !strings.Contains(flag.Funders[0], "Zhejiang University") {
					t.Fatal("incorrect acknowledgement found")
				}
			}
		}

		if !found {
			t.Fatal("missing expected flag")
		}
	})
}
