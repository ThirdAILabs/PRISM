import React, { useState, useContext } from 'react';
import { SearchContext } from '../../../store/searchContext';
import { searchService } from '../../../api/search';
import TodoListComponent from './TodoListComponent';
import { SingleSearchBar } from '../../common/searchBar/SearchBar';

const PaperTitleSearchComponent = () => {
  const { searchState, setSearchState } = useContext(SearchContext);
  const { paperResults, hasSearchedPaper } = searchState;
  const [isLoading, setIsLoading] = useState(false);

  const search = async (paperTitle) => {
    setSearchState((prev) => ({
      ...prev,
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
        />
      </div>
      {isLoading && <div>Loading...</div>}
      {hasSearchedPaper && <TodoListComponent results={paperResults} />}
    </div>
  );
};

export default PaperTitleSearchComponent;
