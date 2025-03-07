import React, { useState, useContext } from 'react';
import { SearchContext } from '../../../store/searchContext';
import { searchService } from '../../../api/search';
import TodoListComponent from './TodoListComponent';
import { SingleSearchBar } from '../../common/searchBar/SearchBar';

const OrcidSearchComponent = () => {
  const { searchState, setSearchState } = useContext(SearchContext);
  const { orcidResults, hasSearchedOrcid } = searchState;
  const [isLoading, setIsLoading] = useState(false);

  const search = async (orcidId) => {
    setSearchState((prev) => ({
      ...prev,
      orcidResults: [],
      hasSearchedOrcid: false,
    }));
    setIsLoading(true);

    try {
      const result = await searchService.searchByOrcid(orcidId);
      console.log('ORCID search result:', result);
      setSearchState((prev) => ({
        ...prev,
        orcidResults: result,
        hasSearchedOrcid: true,
      }));
    } catch (error) {
      console.error('Error searching by ORCID:', error);
    }
    setIsLoading(false);
  };

  return (
    <div style={{ textAlign: 'center', marginTop: '3%' }}>
      <div style={{ marginTop: '1rem' }}>
        <SingleSearchBar
          title="ORCID ID"
          onSearch={search}
          placeholder="E.g. 0000-0002-1825-0097"
        />
      </div>
      {isLoading && <div>Loading...</div>}
      {hasSearchedOrcid && <TodoListComponent results={orcidResults} />}
    </div>
  );
};

export default OrcidSearchComponent;
