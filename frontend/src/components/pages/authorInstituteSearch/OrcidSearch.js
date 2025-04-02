import React, { useState, useContext } from 'react';
import { SearchContext } from '../../../store/searchContext';
import { searchService } from '../../../api/search';
import TodoListComponent from './TodoListComponent';
import { SingleSearchBar } from '../../common/searchBar/SearchBar';
import NoResultsFound from '../../common/tools/NoResultsFound';

const orcidRegex = /^[0-9]{4}-[0-9]{4}-[0-9]{4}-[0-9X]{4}$/;

const OrcidSearchComponent = () => {
  const { searchState, setSearchState } = useContext(SearchContext);
  const { orcidResults, hasSearchedOrcid, orcidQuery } = searchState;
  const [isLoading, setIsLoading] = useState(false);

  const search = async (orcidId) => {
    if (!orcidRegex.test(orcidId)) {
      alert('Invalid ORCID format. Please use 0000-0000-0000-0000 (last digit can be 0-9 or X).');
      return;
    }
    setSearchState((prev) => ({
      ...prev,
      orcidQuery: orcidId,
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
          label="Enter ORCID ID"
          onSearch={search}
          placeholder="E.g. 0000-0002-1825-0097"
          initialValue={orcidQuery}
        />
      </div>
      {isLoading && (
        <div
          className="spinner-border text-primary"
          style={{ width: '3rem', height: '3rem', marginTop: '20px' }}
          role="status"
        ></div>
      )}
      {hasSearchedOrcid && !isLoading && orcidResults.length === 0 && (
        <NoResultsFound
          msg="Author Not Found"
          submsg="We couldn't find an author associated with this ORCID ID."
        />
      )}
      {hasSearchedOrcid && orcidResults.length > 0 && <TodoListComponent results={orcidResults} />}
    </div>
  );
};

export default OrcidSearchComponent;
