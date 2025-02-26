// src/SearchComponent.js
import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import TodoListComponent from '../TodoListComponent';
import { AuthorInstiutionSearchBar } from '../common/searchBar/SearchBar';
import Logo from '../../assets/images/prism-logo.png';
import '../common/searchBar/SearchBar.css';
import '../common/tools/button/button1.css';
import UserService from '../../services/userService';
import { searchService } from '../../api/search';
import { reportService } from '../../api/reports';

const SearchComponent = () => {
  const [query, setQuery] = useState('');
  const [author, setAuthor] = useState();
  const [hasSearched, setHasSearched] = useState(false);
  const [institution, setInstitution] = useState();
  const [results, setResults] = useState([]);
  const [openAlexResults, setOpenAlexResults] = useState([]);
  const [isOALoading, setIsOALoading] = useState([]);
  const [loadMoreCount, setLoadMoreCount] = useState(0);
  const [canLoadMore, setCanLoadMore] = useState(true);

  const search = async (author, institution) => {
    searchOpenAlex(author, institution);
    // setIsLoadingScopus(true);
    setAuthor(author);
    setInstitution(institution);
    setHasSearched(false);
    setQuery(`${author.AuthorName} ${institution ? institution.InstitutionName : ''}`);
    setResults([]);
    setLoadMoreCount(0);
    setHasSearched(true);
    setCanLoadMore(true);
  };

  const searchOpenAlex = async (author, institution) => {
    setIsOALoading(true);
    setAuthor(author);
    setInstitution(institution);
    setHasSearched(false);
    setQuery(`${author.AuthorName} ${institution ? institution.InstitutionName : ''}`);
    setResults([]);

    const result = await searchService.searchOpenalexAuthors(
      author.AuthorName,
      institution.InstitutionId
    );
    console.log('result in openAlex', result);
    setOpenAlexResults(result);
    setIsOALoading(false);
    setLoadMoreCount(0);
    setHasSearched(true);
  };

  return (
    <div className="basic-setup" style={{ color: 'black' }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          width: '100%',
          padding: '20px',
          position: 'relative',
        }}
      >
        <Link
          to="/entity-lookup"
          className="button"
          style={{
            padding: '10px 15px',
            fontSize: '14px',
            whiteSpace: 'nowrap',
            textDecoration: 'none',
            display: 'inline-block',
            width: '25%',
          }}
        >
          Go To Entity Lookup
        </Link>
        <button
          className="button"
          style={{
            padding: '10px 15px',
            fontSize: '14px',
            whiteSpace: 'nowrap',
            textDecoration: 'none',
            display: 'inline-block',
            width: '10%',
          }}
          onClick={UserService.doLogout}
        >
          Logout
        </button>
      </div>

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
        <h1 style={{ fontWeight: 'bold', marginTop: 20, animation: 'fade-in 0.75s' }}>Welcome</h1>
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
            <AuthorInstiutionSearchBar onSearch={search} />
          </div>
        </div>
      </div>
      {hasSearched && (
        <TodoListComponent
          results={openAlexResults}
          setResults={setOpenAlexResults}
          canLoadMore={canLoadMore}
          loadMore={async () => {
            const result = await searchService.searchGoogleScholarAuthors(
              author.AuthorName,
              institution.InstitutionName,
              null
            );
            setCanLoadMore(false);
            return result.Authors;
          }}
        />
      )}
    </div>
  );
};

export default SearchComponent;
