import React, { useState, useContext, useCallback } from 'react';
import { SearchContext } from '../../../store/searchContext';
import { searchService } from '../../../api/search';
import { autocompleteService } from '../../../api/autocomplete';
import TodoListComponent from './TodoListComponent';
import AutocompleteSearchBar from '../../../utils/autocomplete';
import useCallOnPause from '../../../hooks/useCallOnPause';

const PaperTitleSearchComponent = () => {
  const { searchState, setSearchState } = useContext(SearchContext);
  const { paperResults, hasSearchedPaper, paperTitleQuery } = searchState;

  const [paperTitle, setPaperTitle] = useState(paperTitleQuery || '');
  const [isSelected, setIsSelected] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  
  const debouncedSearch = useCallOnPause(300);

  const autocompletePaperTitle = useCallback(
    (query) => {
      return new Promise((resolve) => {
        debouncedSearch(async () => {
          try {
            const res = await autocompleteService.autocompletePaperTitles(query);
            resolve(res);
            return res;
          } catch (error) {
            console.error('Paper title autocomplete error:', error);
            resolve([]);
          }
        });
      });
    },
    [debouncedSearch]
  );

  const search = async (title) => {
    if (!isSelected) {
        alert('Please select a valid paper title from the suggestions');
        return;
      }
    setSearchState((prev) => ({
      ...prev,
      paperTitleQuery: title,
      paperResults: [],
      hasSearchedPaper: false,
    }));
    setIsLoading(true);

    try {
      const result = await searchService.searchByPaperTitle(title);
      console.log('Paper title search result:', result);
      setSearchState((prev) => ({
        ...prev,
        paperResults: result,
        hasSearchedPaper: true,
      }));
    } catch (error) {
      console.error('Error searching by paper title:', error);
    }
    setIsLoading(false);
  };

  return (
    <div style={{ textAlign: 'center', marginTop: '3%' }}>
      <div style={{ marginTop: '1rem' }}>
        <div className="author-institution-search-bar">
          <div className="autocomplete-search-bar">
            <AutocompleteSearchBar
              title="Paper Title"
              autocomplete={autocompletePaperTitle}
              onSelect={(selected) => {
                setPaperTitle(selected.Name);
                setSearchState((prev) => ({ ...prev, paperTitleQuery: selected.Name }));
                setIsSelected(true);
              }}
              placeholder="E.g. Deep Learning for NLP"
              initialValue={paperTitle}
            />
          </div>
          <div style={{ width: '40px' }} />
          <div style={{ width: '200px', marginTop: '40px' }}>
            <button className="button" onClick={() => search(paperTitle)}>
              {isLoading ? 'Searching...' : 'Search'}
            </button>
          </div>
        </div>
      </div>
      {isLoading && <div>Loading...</div>}
      {hasSearchedPaper && !isLoading && paperResults.length === 0 && (
        <div>No results found.</div>
      )}
      {hasSearchedPaper && paperResults.length > 0 && (
        <TodoListComponent results={paperResults} />
      )}
    </div>
  );
};

export default PaperTitleSearchComponent;
