package flaggers

import "prism/api"

const (
	AuthorIsFacultyAtEOCType = "uni_faculty_eoc"
	AuthorIsAssociatedWithEOCType = "doj_press_release_eoc"
)

type Flag struct {
	FlaggerType   string
	Title         string
	Message       string
	UniversityUrl string
	Affiliations  []string
	Metadata      map[string]any
}

type Work struct {
	Authors []api.Author
}
