import React, { useContext, useState } from 'react';
import { SearchContext } from '../../../store/searchContext';
import { searchService } from '../../../api/search';
import TodoListComponent from './TodoListComponent';
import { AuthorInstiutionSearchBar } from '../../common/searchBar/SearchBar';

const AuthorInstitutionSearchComponent = () => {
  const { searchState, setSearchState } = useContext(SearchContext);
  const { author, institution, openAlexResults, hasSearched, canLoadMore } = searchState;

  const [isLoadingMore, setIsLoadingMore] = useState(false);

  const search = async (author, institution) => {
    setSearchState((prev) => ({
      ...prev,
      author,
      institution,
      openAlexResults: [],
      hasSearched: false,
      loadMoreCount: 0,
      canLoadMore: true,
    }));
    searchOpenAlex(author, institution);
  };

  const searchOpenAlex = async (author, institution) => {
    setSearchState((prev) => ({
      ...prev,
      isOALoading: true,
    }));

    const result = await searchService.searchOpenalexAuthors(
      author.Name,
      institution.Id,
      institution.Name
    );
    console.log('result in openAlex', result);
    setSearchState((prev) => ({
      ...prev,
      openAlexResults: result,
      isOALoading: false,
      loadMoreCount: 0,
      hasSearched: true,
    }));
  };

  return (
    <div>
      <AuthorInstiutionSearchBar
        onSearch={search}
        defaultAuthor={author}
        defaultInstitution={institution}
      />
      {hasSearched && !searchState.isOALoading && openAlexResults.length === 0 && (
        <div>No results found.</div>
      )}
      {hasSearched && openAlexResults.length > 0 && (
        <TodoListComponent
          results={openAlexResults}
          setResults={(newResults) =>
            setSearchState((prev) => ({ ...prev, openAlexResults: newResults }))
          }
          canLoadMore={canLoadMore}
          isLoadingMore={isLoadingMore}
          loadMore={async () => {
            if (!canLoadMore) {
              return [];
            }
            setIsLoadingMore(true);
            const result = await searchService.searchGoogleScholarAuthors(
              author.Name,
              institution.Name,
              null
            );
            setIsLoadingMore(false);
            setSearchState((prev) => ({ ...prev, canLoadMore: false }));
            return result.Authors;
          }}
        />
      )}
    </div>
  );
};

export default AuthorInstitutionSearchComponent;
