import React, { createContext, useState } from 'react';

export const SearchContext = createContext();

export const SearchProvider = ({ children }) => {
  const [searchState, setSearchState] = useState({
    query: '',
    author: null,
    institution: null,
    openAlexResults: [],
    hasSearched: false,
    loadMoreCount: 0,
    canLoadMore: true,
    isOALoading: false,
  });

  return (
    <SearchContext.Provider value={{ searchState, setSearchState }}>
      {children}
    </SearchContext.Provider>
  );
};
