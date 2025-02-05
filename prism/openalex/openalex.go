package openalex

import "prism/api"

type KnowledgeBase interface {
	AutocompleteAuthor(query string) ([]api.Author, error)

	AutocompleteInstitution(query string) ([]api.Institution, error)

	FindAuthors(author, institution string) ([]api.Author, error)

	StreamWorks(authorId string, startYear, endYear int) (chan api.WorkBatch, chan error)

	FindWorksByTitle(titles []string, startYear, endYear int) ([]api.Work, error)
}
