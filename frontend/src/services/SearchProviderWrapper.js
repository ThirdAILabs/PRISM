import React from 'react';
import { Outlet } from 'react-router-dom';
import { SearchProvider } from '../store/searchContext';

const SearchProviderWrapper = () => {
  return (
    <SearchProvider>
      <Outlet />
    </SearchProvider>
  );
};

export default SearchProviderWrapper;
