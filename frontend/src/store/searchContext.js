import React, { createContext, useState } from 'react';

export const SearchContext = createContext();

export const SearchProvider = ({ children }) => {
  const [searchState, setSearchState] = useState({
    author: null,
    institution: null,
    openAlexResults: [],
    hasSearched: false,
    loadMoreCount: 0,
    canLoadMore: true,
    isOALoading: false,

    orcidResults: [],
    isOrcidLoading: false,
    hasSearchedOrcid: false,
    orcidQuery: '',
    // Paper Title search state
    paperResults: [],
    isPaperLoading: false,
    hasSearchedPaper: false,
    paperTitleQuery: '',
  });

  return (
    <SearchContext.Provider value={{ searchState, setSearchState }}>
      {children}
    </SearchContext.Provider>
  );
};
