package openalex

import "prism/api"

type KnowledgeBase interface {
	AutocompleteAuthor(query string) ([]api.Author, error)

	AutocompleteInstitution(query string) ([]api.Institution, error)

	FindAuthors(author, institution string) ([]api.Author, error)
}
