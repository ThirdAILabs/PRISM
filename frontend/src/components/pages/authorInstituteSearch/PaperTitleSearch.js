import React, { useState, useContext } from 'react';
import { SearchContext } from '../../../store/searchContext';
import { searchService } from '../../../api/search';
import TodoListComponent from './TodoListComponent';
import { SingleSearchBar } from '../../common/searchBar/SearchBar';

const PaperTitleSearchComponent = () => {
  const { searchState, setSearchState } = useContext(SearchContext);
  const { paperResults, hasSearchedPaper, paperTitleQuery } = searchState;
  const [isLoading, setIsLoading] = useState(false);

  const search = async (paperTitle) => {
    setSearchState((prev) => ({
      ...prev,
      paperTitleQuery: paperTitle,
      paperResults: [],
      hasSearchedPaper: false,
    }));
    setIsLoading(true);

    try {
      const result = await searchService.searchByPaperTitle(paperTitle);
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
        <SingleSearchBar
          title="Paper Title"
          onSearch={search}
          placeholder="E.g. Deep Learning for NLP"
          initialValue={paperTitleQuery}
        />
      </div>
      {isLoading && <div>Loading...</div>}
      {hasSearchedPaper && !isLoading && paperResults.length === 0 && <div>No results found.</div>}
      {hasSearchedPaper && paperResults.length > 0 && <TodoListComponent results={paperResults} />}
    </div>
  );
};

export default PaperTitleSearchComponent;
