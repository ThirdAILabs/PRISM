// src/SearchComponent.js
import React, { useState, useContext } from 'react';
import { Link } from 'react-router-dom';
import TodoListComponent from './TodoListComponent';
import { AuthorInstiutionSearchBar } from '../../common/searchBar/SearchBar';
import Logo from '../../../assets/images/prism-logo.png';
import '../../common/searchBar/SearchBar.css';
import '../../common/tools/button/button1.css';
import UserService from '../../../services/userService';
import { searchService } from '../../../api/search';
import { SearchContext } from '../../../store/searchContext';

const SearchComponent = () => {
  const { searchState, setSearchState } = useContext(SearchContext);
  const { author, institution, openAlexResults, hasSearched, canLoadMore } = searchState;

  const search = async (author, institution) => {
    setSearchState((prev) => ({
      ...prev,
      author: author,
      institution: institution,
      openAlexResults: [],
      hasSearched: false,
      loadMoreCount: 0,
      canLoadMore: true,
    }));
    searchOpenAlex(author, institution);
    // setIsLoadingScopus(true);
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
    <div className="basic-setup" style={{ color: 'black' }}>
      <div style={{ textAlign: 'center', marginTop: '3%', animation: 'fade-in 0.75s' }}>
        <img
          src={Logo}
          style={{
            width: '320px',
            marginTop: '1%',
            marginBottom: '1%',
            marginRight: '2%',
            animation: 'fade-in 0.5s',
          }}
        />
        <h1 style={{ fontWeight: 'bold', marginTop: 20, animation: 'fade-in 0.75s' }}>
          Individual Assessment
        </h1>
        <div style={{ animation: 'fade-in 1s' }}>
          <div className="d-flex justify-content-center align-items-center">
            <div style={{ marginTop: 10, color: '#888888' }}>
              We help you comply with research security requirements by automating author
              assessments.
            </div>
          </div>
          <div className="d-flex justify-content-center align-items-center">
            <div style={{ marginTop: 10, marginBottom: '1%', color: '#888888' }}>
              Who would you like to conduct an assessment on?
            </div>
          </div>
        </div>
        <div className="d-flex justify-content-center align-items-center pt-5">
          <div style={{ width: '80%', animation: 'fade-in 1.25s' }}>
            <AuthorInstiutionSearchBar
              onSearch={search}
              defaultAuthor={author}
              defaultInstitution={institution}
            />
          </div>
        </div>
      </div>
      <div className="d-flex justify-content-center align-items-center pt-5">
        <div style={{ width: '80%', animation: 'fade-in 1.25s' }}>
          {hasSearched && (
            <TodoListComponent
              results={openAlexResults}
              setResults={(newResults) =>
                setSearchState((prev) => ({ ...prev, openAlexResults: newResults }))
              }
              canLoadMore={canLoadMore}
              loadMore={async () => {
                if (!canLoadMore) {
                  return [];
                }
                const result = await searchService.searchGoogleScholarAuthors(
                  author.Name,
                  institution.Name,
                  null
                );
                setSearchState((prev) => ({ ...prev, canLoadMore: false }));
                return result.Authors;
              }}
            />
          )}
        </div>
      </div>
    </div>
  );
};

export default SearchComponent;
