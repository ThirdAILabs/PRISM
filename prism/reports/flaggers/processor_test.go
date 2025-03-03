package flaggers

import (
	"os"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/licensing"
	"prism/prism/openalex"
	"prism/prism/reports"
	"prism/prism/reports/flaggers/eoc"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

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

func getLicensing(t *testing.T) *licensing.LicenseVerifier {
	licensing, err := licensing.NewLicenseVerifier("AC013F-FD0B48-00B160-64836E-76E88D-V3")
	if err != nil {
		t.Fatal(err)
	}
	return licensing
}

func TestProcessorCoauthorAffiliationCase1(t *testing.T) {
	processor := ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		workFlaggers: []WorkFlagger{
			&OpenAlexCoauthorAffiliationIsEOC{
				concerningEntities:     eoc.LoadGeneralEOC(),
				concerningInstitutions: eoc.LoadInstitutionEOC(),
			},
		},
		licensing: getLicensing(t),
	}

	t.Run("Case1", func(t *testing.T) {
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5084836278",
			AuthorName: "Charles M. Lieber",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2019),
			EndDate:    yearEnd(2019),
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report[api.CoauthorAffiliationType]) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := report[api.CoauthorAffiliationType][0].(*api.CoauthorAffiliationFlag)

		if len(flag.Affiliations) != 1 || flag.Affiliations[0] != "Central South University" || len(flag.Coauthors) != 1 || flag.Coauthors[0] != "Jian Sun" {
			t.Fatal("incorrect flag")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5075113943",
			AuthorName: "Zijian Hong",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2024),
			EndDate:    yearEnd(2024),
		})
		if err != nil {
			t.Fatal(err)
		}

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
	processor := ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		workFlaggers: []WorkFlagger{
			&OpenAlexAuthorAffiliationIsEOC{
				concerningEntities:     eoc.LoadGeneralEOC(),
				concerningInstitutions: eoc.LoadInstitutionEOC(),
			},
		},
		licensing: getLicensing(t),
	}

	t.Run("Case1", func(t *testing.T) {
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5084836278",
			AuthorName: "Charles M. Lieber",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2013),
			EndDate:    yearEnd(2013),
		})
		if err != nil {
			t.Fatal(err)
		}

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
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5075113943",
			AuthorName: "Zijian Hong",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2024),
			EndDate:    yearEnd(2024),
		})
		if err != nil {
			t.Fatal(err)
		}

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

	processor := ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		authorFacultyAtEOC: &AuthorIsFacultyAtEOCFlagger{
			universityNDB: universityNDB,
		},
		licensing: getLicensing(t),
	}

	t.Run("Case1", func(t *testing.T) {
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5019148940",
			AuthorName: "Natalie Artzi",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2024),
			EndDate:    yearEnd(2024),
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report[api.PotentialAuthorAffiliationType]) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := report[api.PotentialAuthorAffiliationType][0].(*api.PotentialAuthorAffiliationFlag)

		if flag.University != "Fudan University" || !strings.Contains(flag.UniversityUrl, "fudan.edu") {
			t.Fatal("incorrect flag")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5075113943",
			AuthorName: "Zijian Hong",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2024),
			EndDate:    yearEnd(2024),
		})
		if err != nil {
			t.Fatal(err)
		}

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

	processor := ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		authorAssociatedWithEOC: &AuthorIsAssociatedWithEOCFlagger{
			docNDB: docNDB,
			auxNDB: auxNDB,
		},
		licensing: getLicensing(t),
	}

	t.Run("PrimaryConnection", func(t *testing.T) {
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5084836278",
			AuthorName: "Charles M. Lieber",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2013),
			EndDate:    yearEnd(2013),
		})
		if err != nil {
			t.Fatal(err)
		}

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
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5012289937",
			AuthorName: "Anqi Zhang",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2015),
			EndDate:    yearEnd(2020),
		})
		if err != nil {
			t.Fatal(err)
		}

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
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5016320004",
			AuthorName: "David Zhang",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2020),
			EndDate:    yearEnd(2020),
		})
		if err != nil {
			t.Fatal(err)
		}

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

	processor := ReportProcessor{
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
		licensing: getLicensing(t),
	}

	t.Run("Case1", func(t *testing.T) {
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5084836278",
			AuthorName: "Charles M. Lieber",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2011),
			EndDate:    yearEnd(2013),
		})
		if err != nil {
			t.Fatal(err)
		}

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
		report, err := processor.ProcessReport(reports.ReportUpdateTask{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5075113943",
			AuthorName: "Zijian Hong",
			Source:     api.OpenAlexSource,
			StartDate:  yearStart(2023),
			EndDate:    yearEnd(2023),
		})
		if err != nil {
			t.Fatal(err)
		}

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
