package flaggers

import (
	"os"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/openalex"
	"prism/prism/reports/flaggers/eoc"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func eqOrderInvariant(a, b []string) bool {
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
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
	}

	t.Run("Case1", func(t *testing.T) {
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5084836278",
			AuthorName: "Charles M. Lieber",
			Source:     api.OpenAlexSource,
			StartYear:  2019,
			EndYear:    2019,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.CoauthorAffiliations) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := report.CoauthorAffiliations[0]

		if len(flag.Affiliations) != 1 || flag.Affiliations[0] != "Central South University" || len(flag.Coauthors) != 1 || flag.Coauthors[0] != "Jian Sun" {
			t.Fatal("incorrect flag")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5075113943",
			AuthorName: "Zijian Hong",
			Source:     api.OpenAlexSource,
			StartYear:  2024,
			EndYear:    2024,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.CoauthorAffiliations) < 1 {
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
		for _, flag := range report.CoauthorAffiliations {
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
	}

	t.Run("Case1", func(t *testing.T) {
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5084836278",
			AuthorName: "Charles M. Lieber",
			Source:     api.OpenAlexSource,
			StartYear:  2013,
			EndYear:    2013,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.AuthorAffiliations) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		found := false
		for _, flag := range report.AuthorAffiliations {
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
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5075113943",
			AuthorName: "Zijian Hong",
			Source:     api.OpenAlexSource,
			StartYear:  2024,
			EndYear:    2024,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.AuthorAffiliations) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		found := false
		for _, flag := range report.AuthorAffiliations {
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
	universityNDB := BuildUniversityNDB("../../../data/university_websites_with_entities_filtered.json", t.TempDir())
	defer universityNDB.Free()

	processor := ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		authorFacultyAtEOC: &AuthorIsFacultyAtEOCFlagger{
			universityNDB: universityNDB,
		},
	}

	t.Run("Case1", func(t *testing.T) {
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5019148940",
			AuthorName: "Natalie Artzi",
			Source:     api.OpenAlexSource,
			StartYear:  2024,
			EndYear:    2024,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.PotentialAuthorAffiliations) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := report.PotentialAuthorAffiliations[0]

		if flag.University != "Fudan University" || !strings.Contains(flag.UniversityUrl, "fudan.edu") {
			t.Fatal("incorrect flag")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5075113943",
			AuthorName: "Zijian Hong",
			Source:     api.OpenAlexSource,
			StartYear:  2024,
			EndYear:    2024,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.PotentialAuthorAffiliations) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		found := false
		for _, flag := range report.PotentialAuthorAffiliations {
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
	docNDB := BuildDocNDB("../../../data/doj_articles_with_content_and_entities_as_text.json", t.TempDir())
	defer docNDB.Free()

	auxNDB := BuildAuxNDB("../../../data/doj_relevant_webpages_cleaned_with_entities_as_content.json", t.TempDir())
	defer auxNDB.Free()

	processor := ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		authorAssociatedWithEOC: &AuthorIsAssociatedWithEOCFlagger{
			docNDB: docNDB,
			auxNDB: auxNDB,
		},
	}

	t.Run("PrimaryConnection", func(t *testing.T) {
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5084836278",
			AuthorName: "Charles M. Lieber",
			Source:     api.OpenAlexSource,
			StartYear:  2013,
			EndYear:    2013,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.MiscHighRiskAssociations) < 1 {
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

		for _, flag := range report.MiscHighRiskAssociations {
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
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5012289937",
			AuthorName: "Anqi Zhang",
			Source:     api.OpenAlexSource,
			StartYear:  2015,
			EndYear:    2020,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.MiscHighRiskAssociations) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		for _, flag := range report.MiscHighRiskAssociations {
			if len(flag.Connections) != 1 || flag.FrequentCoauthor == nil || *flag.FrequentCoauthor != "Charles M. Lieber" {
				t.Fatal("incorrect flag")
			}
		}
	})

	t.Run("TertiaryConnection", func(t *testing.T) {
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5016320004",
			AuthorName: "David Zhang",
			Source:     api.OpenAlexSource,
			StartYear:  2020,
			EndYear:    2020,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.MiscHighRiskAssociations) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		entitiesMentioned := map[string]bool{}
		for _, flag := range report.MiscHighRiskAssociations {
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

	ackFlagCache, err := NewCache[cachedAckFlag]("ack_flags", filepath.Join(testDir, "ack_flags.cache"))
	if err != nil {
		t.Fatal(err)
	}
	defer ackFlagCache.Close()

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
				flagCache:    ackFlagCache,
				authorCache:  authorCache,
				extractor:    NewGrobidExtractor(ackCache, grobidEndpoint, testDir),
				sussyBakas:   eoc.LoadSussyBakas(),
			},
		},
	}

	t.Run("Case1", func(t *testing.T) {
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5084836278",
			AuthorName: "Charles M. Lieber",
			Source:     api.OpenAlexSource,
			StartYear:  2011,
			EndYear:    2013,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.TalentContracts) != 3 || len(report.HighRiskFunders) != 1 {
			t.Fatal("expected 4 acknowledgement flags")
		}

		expectedTitles := []string{
			"Nanoelectronics-biology frontier: From nanoscopic probes for action potential recording in live cells to three-dimensional cyborg tissues",
			"Nanowire Biosensors for Label-Free, Real-Time, Ultrasensitive Protein Detection",
			"Design and Synthesis of Diverse Functional Kinked Nanowire Structures for Nanoelectronic Bioprobes",
			"Nanoelectronics Meets Biology: From New Nanoscale Devices for Live‐Cell Recording to 3D Innervated Tissues",
		}

		titles := make([]string, 0)
		for _, flag := range report.TalentContracts {
			titles = append(titles, flag.Work.DisplayName)
		}
		for _, flag := range report.HighRiskFunders {
			titles = append(titles, flag.Work.DisplayName)
		}

		if !eqOrderInvariant(expectedTitles, titles) {
			t.Fatal("incorrect flags")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		report, err := processor.ProcessReport(api.Report{
			Id:         uuid.New(),
			AuthorId:   "https://openalex.org/A5075113943",
			AuthorName: "Zijian Hong",
			Source:     api.OpenAlexSource,
			StartYear:  2023,
			EndYear:    2023,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(report.HighRiskFunders) < 1 {
			t.Fatal("expected >= 1 flags")
		}

		found := false
		for _, flag := range report.HighRiskFunders {
			if flag.Work.WorkId == "https://openalex.org/W4384197626" {
				found = true
				if !flag.FromAcknowledgements || !strings.Contains(flag.Funders[0], "Zhejiang University") {
					t.Fatal("incorrect acknowledgement found")
				}
			}
		}

		if !found {
			t.Fatal("missing expected flag")
		}
	})
}
