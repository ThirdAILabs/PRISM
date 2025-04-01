import React, { useState, useContext, useCallback } from 'react';
import { SearchContext } from '../../../store/searchContext';
import { searchService } from '../../../api/search';
import { autocompleteService } from '../../../api/autocomplete';
import TodoListComponent from './TodoListComponent';
import AutocompleteSearchBar from '../../../utils/autocomplete';
import useCallOnPause from '../../../hooks/useCallOnPause';
import NoResultsFound from '../../common/tools/NoResultsFound';

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
        <div className="paper-search-bar">
          <div className="paper-title-search-bar">
            <AutocompleteSearchBar
              label={'Enter Paper Title'}
              autocomplete={autocompletePaperTitle}
              onSelect={(selected) => {
                setPaperTitle(selected.Name);
                setSearchState((prev) => ({ ...prev, paperTitleQuery: selected.Name }));
                setIsSelected(true);
              }}
              initialValue={paperTitle}
            />
          </div>
          <div style={{ width: '40px' }} />
          <div style={{ width: '200px', marginTop: '-2px' }}>
            <button className="button button-3d" onClick={() => search(paperTitle)} disabled={!isSelected}>
              {isLoading ? 'Searching...' : 'Search'}
            </button>
          </div>
        </div>
      </div>
      {isLoading && (
        <div
          className="spinner-border text-primary"
          style={{ width: '3rem', height: '3rem', marginTop: '20px' }}
          role="status"
        ></div>
      )}
      {hasSearchedPaper && !isLoading && paperResults.length === 0 && (
        <NoResultsFound
          msg="Authors Not Found"
          submsg="We couldn't find any authors associated with this paper."
        />
      )}
      {hasSearchedPaper && paperResults.length > 0 && <TodoListComponent results={paperResults} />}
    </div>
  );
};

export default PaperTitleSearchComponent;
