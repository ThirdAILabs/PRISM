package flaggers

import (
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
		flags, err := processor.ProcessReport(api.Report{
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

		if len(flags) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := flags[0].(*EOCCoauthorAffiliationsFlag)

		if len(flag.Institutions) != 1 || flag.Institutions[0] != "Central South University" || len(flag.Authors) != 1 || flag.Authors[0] != "Jian Sun" {
			t.Fatal("incorrect flag")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		flags, err := processor.ProcessReport(api.Report{
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

		if len(flags) < 1 {
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
		for _, flag := range flags {
			flag := flag.(*EOCCoauthorAffiliationsFlag)
			if flag.Work.WorkId == "https://openalex.org/W4402273377" {
				foundFlag = true
				if !eqOrderInvariant(flag.Institutions, expectedInstitutions) || !eqOrderInvariant(flag.Authors, expectedAuthors) {
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
		flags, err := processor.ProcessReport(api.Report{
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

		if len(flags) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		found := false
		for _, flag := range flags {
			flag := flag.(*EOCAuthorAffiliationsFlag)
			if len(flag.Institutions) != 1 {
				t.Fatal("incorrect flag")
			}
			if flag.Institutions[0] == "Peking University" {
				found = true
			}
		}

		if !found {
			t.Fatal("missing flag")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		flags, err := processor.ProcessReport(api.Report{
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

		if len(flags) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		found := false
		for _, flag := range flags {
			flag := flag.(*EOCAuthorAffiliationsFlag)
			if len(flag.Institutions) != 1 {
				t.Fatal("incorrect flag")
			}
			if flag.Institutions[0] == "Zhejiang University" {
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
		flags, err := processor.ProcessReport(api.Report{
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

		if len(flags) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := flags[0].(*AuthorIsFacultyAtEOCFlag)

		if flag.University != "Fudan University" || !strings.Contains(flag.UniversityUrl, "fudan.edu") {
			t.Fatal("incorrect flag")
		}
	})

	t.Run("Case2", func(t *testing.T) {
		flags, err := processor.ProcessReport(api.Report{
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

		if len(flags) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		found := false
		for _, flag := range flags {
			flag := flag.(*AuthorIsFacultyAtEOCFlag)
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
		flags, err := processor.ProcessReport(api.Report{
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

		if len(flags) < 1 {
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

		for _, flag := range flags {
			flag := flag.(*AuthorIsAssociatedWithEOCFlag)
			if flag.ConnectionLevel != "primary" || flag.EntityMentioned != "Charles M. Lieber" {
				t.Fatal("incorrect flag")
			}
			titles = append(titles, flag.DocTitle)
		}

		if !eqOrderInvariant(expectedTitles, titles) {
			t.Fatal("incorrect docs found")
		}
	})

	t.Run("SecondaryConnection", func(t *testing.T) {
		flags, err := processor.ProcessReport(api.Report{
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

		if len(flags) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		for _, flag := range flags {
			flag := flag.(*AuthorIsAssociatedWithEOCFlag)
			if flag.ConnectionLevel != "secondary" || flag.FrequentCoauthor == nil || *flag.FrequentCoauthor != "Charles M. Lieber" {
				t.Fatal("incorrect flag")
			}
		}
	})

	t.Run("TertiaryConnection", func(t *testing.T) {
		flags, err := processor.ProcessReport(api.Report{
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

		if len(flags) < 1 {
			t.Fatal("expected >= 1 flag")
		}

		entitiesMentioned := map[string]bool{}
		for _, flag := range flags {
			flag := flag.(*AuthorIsAssociatedWithEOCFlag)

			if flag.ConnectionLevel != "tertiary" ||
				len(flag.Nodes) != 2 ||
				flag.Nodes[0].DocTitle != "NuProbe About Us" ||
				flag.Nodes[0].DocUrl != "https://nuprobe.com/about-us/" ||
				flag.Nodes[1].DocTitle != "NuProbe Announces $11 Million Series A Funding Round" ||
				flag.Nodes[1].DocUrl != "https://nuprobe.com/2018/04/nuprobe-announces-11-million-series-a-funding-round-2/" {
				t.Fatal("incorrect flag")
			}
			entitiesMentioned[flag.EntityMentioned] = true
		}
		if !entitiesMentioned["WuXi AppTec"] || !entitiesMentioned["Sequoia Capital China"] {
			t.Fatal("incorrect entities mentioned")
		}
	})
}
