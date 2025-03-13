package flaggers

import (
	"os"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/openalex"
	"prism/prism/reports"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/schema"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupReportManager(t *testing.T) *reports.ReportManager {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&schema.AuthorReport{}, &schema.AuthorFlag{}, &schema.UserAuthorReport{},
		&schema.UniversityReport{}, &schema.UserUniversityReport{}); err != nil {
		t.Fatal(err)
	}

	// We're using old date ranges for these so they only process works around the
	// date of the flagged work, particularly for the acknowledgements flagger, this
	// ensures the tests don't take too long to run. However this means that the reports
	// are viewed as stale and will be reprocessed on the next test. Setting a large
	// threshold here fixes it. In ~100 years this threshold would again be too small,
	// but at that point this code will likely not be in use, or if it is, then it
	// will be someone else's problem (or more likely an AI).
	return reports.NewManager(db, 100*365*24*time.Hour)
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

func getReportContent(t *testing.T, report reports.ReportUpdateTask, processor *ReportProcessor, manager *reports.ReportManager) map[string][]api.Flag {
	user := uuid.New()
	reportId, err := manager.CreateAuthorReport(user, report.AuthorId, report.AuthorName, report.Source)
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
	processor := &ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		workFlaggers: []WorkFlagger{
			&OpenAlexCoauthorAffiliationIsEOC{
				concerningEntities:     eoc.LoadGeneralEOC(),
				concerningInstitutions: eoc.LoadInstitutionEOC(),
			},
		},
		manager: manager,
	}

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
			"Wei Li",
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
					t.Fatal("incorrect flag returned")
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

	processor := &ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		workFlaggers: []WorkFlagger{
			&OpenAlexAuthorAffiliationIsEOC{
				concerningEntities:     eoc.LoadGeneralEOC(),
				concerningInstitutions: eoc.LoadInstitutionEOC(),
			},
		},
		manager: manager,
	}

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
	universityNDB := BuildUniversityNDB("../../../data/university_webpages.json", t.TempDir())
	defer universityNDB.Free()

	manager := setupReportManager(t)
	processor := &ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		authorFacultyAtEOC: &AuthorIsFacultyAtEOCFlagger{
			universityNDB: universityNDB,
		},
		manager: manager,
	}

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
	docNDB := BuildDocNDB("../../../data/docs_and_press_releases.json", t.TempDir())
	defer docNDB.Free()

	auxNDB := BuildAuxNDB("../../../data/auxiliary_webpages.json", t.TempDir())
	defer auxNDB.Free()

	manager := setupReportManager(t)
	processor := &ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		authorAssociatedWithEOC: &AuthorIsAssociatedWithEOCFlagger{
			docNDB: docNDB,
			auxNDB: auxNDB,
		},
		manager: manager,
	}

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
				t.Fatal("incorrect flag")
			}
			entitiesMentioned[flag.EntityMentioned] = true
		}
		if !entitiesMentioned["WuXi AppTec"] || !entitiesMentioned["Sequoia Capital China"] {
			t.Fatal("incorrect entities mentioned")
		}
		t.Fatal("stop here")
	})
}

func TestProcessorAcknowledgements(t *testing.T) {
	grobidEndpoint := os.Getenv("GROBID_ENDPOINT")
	if grobidEndpoint == "" {
		t.Skip("No grobid endpoint detected, skipping acknowledgements tests. Set GROBID_ENDPOINT env variable to run these tests")
	}

	testDir := t.TempDir()

	authorCache, err := NewCache[openalex.Author]("authors", filepath.Join(testDir, "authors.cache"))
	if err != nil {
		t.Fatal(err)
	}
	defer authorCache.Close()

	ackCache, err := NewCache[Acknowledgements]("acks", filepath.Join(testDir, "acks.cache"))
	if err != nil {
		t.Fatal(err)
	}
	defer ackCache.Close()

	entityStore, err := NewEntityStore(filepath.Join(testDir, "entity_lookup.ndb"), eoc.LoadSourceToAlias())
	if err != nil {
		t.Fatal(err)
	}
	defer entityStore.Free()

	manager := setupReportManager(t)
	processor := &ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		workFlaggers: []WorkFlagger{
			&OpenAlexAcknowledgementIsEOC{
				openalex:     openalex.NewRemoteKnowledgeBase(),
				entityLookup: entityStore,
				authorCache:  authorCache,
				extractor:    NewGrobidExtractor(ackCache, grobidEndpoint, testDir),
				sussyBakas:   eoc.LoadSussyBakas(),
			},
		},
		manager: manager,
	}

	t.Run("Case1", func(t *testing.T) {
		report := getReportContent(
			t, reports.ReportUpdateTask{
				Id:         uuid.New(),
				AuthorId:   "https://openalex.org/A5084836278",
				AuthorName: "Charles M. Lieber",
				Source:     api.OpenAlexSource,
				StartDate:  yearStart(2011),
				EndDate:    yearEnd(2013),
			},
			processor,
			manager,
		)

		if len(report[api.TalentContractType]) != 3 || len(report[api.HighRiskFunderType]) != 1 {
			t.Fatal("expected 4 acknowledgement flags")
		}

		expectedTitles := []string{
			"Nanoelectronics-biology frontier: From nanoscopic probes for action potential recording in live cells to three-dimensional cyborg tissues",
			"Nanowire Biosensors for Label-Free, Real-Time, Ultrasensitive Protein Detection",
			"Design and Synthesis of Diverse Functional Kinked Nanowire Structures for Nanoelectronic Bioprobes",
			"Nanoelectronics Meets Biology: From New Nanoscale Devices for Live‐Cell Recording to 3D Innervated Tissues",
		}

		titles := make([]string, 0)
		for _, flag := range report[api.TalentContractType] {
			flag := flag.(*api.TalentContractFlag)
			titles = append(titles, flag.Work.DisplayName)
		}
		for _, flag := range report[api.HighRiskFunderType] {
			flag := flag.(*api.HighRiskFunderFlag)
			titles = append(titles, flag.Work.DisplayName)
		}

		if !eqOrderInvariant(expectedTitles, titles) {
			t.Fatal("incorrect flags")
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
				if len(flag.RawAcknowledements) == 0 || !strings.Contains(flag.Funders[0], "Zhejiang University") {
					t.Fatal("incorrect acknowledgement found")
				}
			}
		}

		if !found {
			t.Fatal("missing expected flag")
		}
	})
}
