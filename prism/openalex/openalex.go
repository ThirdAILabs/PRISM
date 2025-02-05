package openalex

import "prism/api"

type KnowledgeBase interface {
	AutocompleteAuthor(query string) ([]api.Author, error)

	AutocompleteInstitution(query string) ([]api.Institution, error)

	FindAuthors(author, institution string) ([]api.Author, error)

	FindWorks(authorId string, startYear, endYear int) (chan []api.Work, chan error)

	FindWorksByTitle(titles []string, startYear, endYear int) ([]api.Work, error)
}
