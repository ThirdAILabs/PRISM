// src/SearchComponent.js
import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import TodoListComponent from '../TodoListComponent';
import { AuthorInstiutionSearchBar } from '../common/searchBar/SearchBar';
import Logo from "../../assets/images/prism-logo.png";
import "../common/searchBar/SearchBar.css";
import "../common/tools/button/button1.css"
import UserService from '../../services/userService';
import { searchService } from '../../api/search';
import { reportService } from '../../api/reports';

const SearchComponent = () => {
  const [query, setQuery] = useState('');
  const [author, setAuthor] = useState();
  const [hasSearched, setHasSearched] = useState(false);
  const [resultHeader, setResultHeader] = useState(null);
  const [institution, setInstitution] = useState();
  const [results, setResults] = useState([]);
  const [openAlexResults, setOpenAlexResults] = useState([]);
  const [isOALoading, setIsOALoading] = useState([]);
  const [scholarResults, setScholarResults] = useState([]);
  const [nextToken, setNextToken] = useState(null);
  const [isDpLoading, setIsDpLoading] = useState(false);
  const [triedDp, setTriedDp] = useState(false);
  const [loadMoreCount, setLoadMoreCount] = useState(0);
  const [isLoadingScopus, setIsLoadingScopus] = useState(false);
  const [showResultHeaders, setShowResultHeaders] = useState(false);

  const search = async (author, institution) => {
    // setShowResultHeaders(true);
    handleDeepSearch(
      `${author.AuthorName} ${institution.InstitutionName}`,
      nextToken,
      /* reset= */ false
    );
    searchOpenAlex(author, institution);
    // setIsLoadingScopus(true);
    setAuthor(author);
    setInstitution(institution);
    setHasSearched(false);
    setQuery(`${author.AuthorName} ${institution ? institution.InstitutionName : ''}`);
    setResults([]);
    setLoadMoreCount(0);
    setHasSearched(true);
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

  const handleDeepSearch = async (query, ntoken, reset = false) => {
    if (reset) {
      setScholarResults([]);
    }
    setIsDpLoading(true);
    setTriedDp(true);

    try {
      let result;

      console.log('N token is', ntoken);
      if (ntoken !== null) {
        result = await searchService.searchGoogleScholarAuthors(query, ntoken);
        console.log('Old deep search with query', query);
      } else {
        console.log('New deep search with query', query);
        result = await searchService.searchGoogleScholarAuthors(query, ntoken);
        console.log('Got results', results);
      }

      setScholarResults(result.Authors);

      setNextToken(result.next_page_token);

      setIsDpLoading(false);
    } catch (error) {
      console.error('Error during deep search', error);
      setIsDpLoading(false);
    }
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

      <div style={{ textAlign: "center", marginTop: "3%", animation: "fade-in 0.75s" }}>
        <img src={Logo} style={{ width: "320px", marginTop: "1%", marginBottom: "1%", marginRight: "2%", animation: "fade-in 0.5s" }} />
        <h1 style={{ fontWeight: "bold", marginTop: 20, animation: "fade-in 0.75s" }}>Welcome</h1>
        <div style={{ animation: "fade-in 1s" }}>
          <div className='d-flex justify-content-center align-items-center'>
            <div style={{ marginTop: 10, color: "#888888" }}>We help you comply with research security requirements by automating author assessments.</div>
          </div>
          <div className='d-flex justify-content-center align-items-center'>
            <div style={{ marginTop: 10, marginBottom: "1%", color: "#888888" }}>Who would you like to conduct an assessment on?</div>
          </div>
        </div>
        <div className="d-flex justify-content-center align-items-center pt-5">
          <div style={{ width: '80%', animation: 'fade-in 1.25s' }}>
            <AuthorInstiutionSearchBar onSearch={search} />
          </div>
        </div>
      </div>
      {showResultHeaders && (
        <div
          style={{ paddingTop: '30px', textAlign: 'center', fontSize: '24px', fontWeight: 'bold' }}
        >
          OpenAlex Results
        </div>
      )}
      {isLoadingScopus && (
        <div
          className="spinner-border text-primary"
          style={{ width: '3rem', height: '3rem' }}
          role="status"
        ></div>
      )}
      {hasSearched && (
        <TodoListComponent results={openAlexResults} canLoadMore={false} loadMore={() => {}} />
      )}

      {showResultHeaders && (
        <div
          style={{ paddingTop: '30px', textAlign: 'center', fontSize: '24px', fontWeight: 'bold' }}
        >
          Google Scholar Results
        </div>
      )}
      {isDpLoading && (
        <div
          className="spinner-border text-primary"
          style={{ width: '3rem', height: '3rem' }}
          role="status"
        ></div>
      )}
      {hasSearched && (
        <TodoListComponent results={scholarResults} canLoadMore={false} loadMore={() => {}} />
      )}
    </div>
  );
};

export default SearchComponent;
