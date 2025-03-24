package flaggers

type EntityOfConcern struct {
	Name        string `json:"name"`
	Country     string `json:"country"`
	FieldOfWork string `json:"field_of_work"`
}

type IndividualEOC struct {
	EntityOfConcern
	Affiliations []string        `json:"affiliations"`
	Relations    []RelatedEntity `json:"relations,omitempty"`
}

type InstitutionEOC struct {
	EntityOfConcern
	Relations []RelatedEntity `json:"relations,omitempty"`
}

type RelatedEntity struct {
	RelatedTo    EntityOfConcern `json:"related_to"`
	RelationType string          `json:"relation_type"`
}

type DojArticleRecord struct {
	Title        string           `json:"title"`
	Url          string           `json:"url"`
	Text         string           `json:"text"`
	Individuals  []IndividualEOC  `json:"individuals"`
	Institutions []InstitutionEOC `json:"institutions"`
}

func (r *DojArticleRecord) getEntitiesForIndexing() []string {
	entities := make([]string, 0, len(r.Individuals)+len(r.Institutions))
	for _, individual := range r.Individuals {
		entities = append(entities, individual.Name)
	}
	for _, institution := range r.Institutions {
		entities = append(entities, institution.Name)
	}
	return entities
}

type ReleveantWebpageRecord struct {
	Url          string           `json:"url"`
	Title        string           `json:"title"`
	Text         string           `json:"text"`
	Individuals  []IndividualEOC  `json:"individuals"`
	Institutions []InstitutionEOC `json:"institutions"`
	ReferredFrom string           `json:"referred_from"`
}

func (r *ReleveantWebpageRecord) getEntitiesForIndexing() []string {
	entities := make([]string, 0, len(r.Individuals)+len(r.Institutions))
	for _, individual := range r.Individuals {
		entities = append(entities, individual.Name)
	}
	for _, institution := range r.Institutions {
		entities = append(entities, institution.Name)
	}
	return entities
}

func (r *ReleveantWebpageRecord) getEntitiesForHops() []string {
	entity_set := make(map[string]bool)
	for _, individual := range r.Individuals {
		entity_set[individual.Name] = true
		for _, relation := range individual.Relations {
			entity_set[relation.RelatedTo.Name] = true
		}
	}
	for _, institution := range r.Institutions {
		entity_set[institution.Name] = true
		for _, relation := range institution.Relations {
			entity_set[relation.RelatedTo.Name] = true
		}
	}
	entities := make([]string, 0, len(entity_set))
	for entity := range entity_set {
		entities = append(entities, entity)
	}
	return entities
}
