// src/SearchComponent.js
import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import apiService from './api/apiService';
import TodoListComponent from './TodoListComponent';
import { AuthorInstiutionSearchBar } from './SearchBar';
import Logo from "./assets/logo.png"
import "./SearchBar.css";
import { levenshteinDistance, makeVariations } from './api/nameUtils';

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
    setShowResultHeaders(true);
    handleDeepSearch(`${author.display_name} ${institution.display_name}`, nextToken, /* reset= */ false);
    searchOpenAlex(author, institution);
    setIsLoadingScopus(true);
    setAuthor(author);
    setInstitution(institution);
    setHasSearched(false);
    setQuery(`${author.display_name} ${institution ? institution.display_name : ''}`);
    setResults([]);

    for (const authorName of makeVariations(author.display_name)) {
      await new Promise(resolve => setTimeout(resolve, /* ms= */ 300));
      try {
        await apiService.search(authorName, institution ? institution.display_name : '').then(result => {
          console.log(result.profiles);
          setResults(prev => {
            let newResults = [...prev];
            for (const profile of result.profiles) {
              let seen = false;
              console.log("profile", profile);
              console.log("new results", newResults);
              for (const otherProfile of newResults) {
                if (otherProfile.id === profile.id) {
                  seen = true;
                }
              }
              if (!seen) {
                newResults.push(profile);
              }
            }
            newResults = newResults.map(x => [x, levenshteinDistance(x.display_name, author.display_name)])
            console.log(newResults);
            newResults.sort((a, b) => a[1] - b[1]);
            console.log(newResults);
            newResults = newResults.map(x => x[0]);
            return newResults;
          });
        });
        setResultHeader('Scopus Results');

      } catch (error) {
        console.log('Unable to fetch Scopus results.')
      }
    }
    setIsLoadingScopus(false);
    setLoadMoreCount(0);
    setHasSearched(true);
  };

  const searchOpenAlex = async (author, institution) => {
    // handleDeepSearch(`${author.display_name} ${institution.display_name}`, nextToken, /* reset= */ false);
    setIsOALoading(true);
    setAuthor(author);
    setInstitution(institution);
    setHasSearched(false);
    setQuery(`${author.display_name} ${institution ? institution.display_name : ''}`);
    setResults([]);

    for (const authorName of makeVariations(author.display_name)) {
      await new Promise(resolve => setTimeout(resolve, /* ms= */ 300));

      try {
        await apiService.searchOpenAlex(authorName, institution ? institution.display_name : '').then(result => {
          console.log(result.profiles);
          setOpenAlexResults(prev => {
            let newResults = [...prev];
            for (const profile of result.profiles) {
              let seen = false;
              console.log("profile", profile);
              console.log("new results", newResults);
              for (const otherProfile of newResults) {
                if (otherProfile.id === profile.id) {
                  seen = true;
                }
              }
              if (!seen) {
                newResults.push(profile);
              }
            }
            newResults = newResults.map(x => [x, levenshteinDistance(x.display_name, author.display_name)])
            console.log(newResults);
            newResults.sort((a, b) => a[1] - b[1]);
            console.log(newResults);
            newResults = newResults.map(x => x[0]);
            return newResults;
          });
        });
      } catch (error) {
        console.log('Unable to fetch from OpenAlex: ', error);
      }

    }
    setIsOALoading(false);
    setLoadMoreCount(0);
    setHasSearched(true);
  };

  console.log(results);

  // const handleSearchResult = (profiles) => {
  //   setTriedDp(false);
  //   setResults(profiles);
  // };


  // const loadMore = async () => {
  //   if (loadMoreCount === 0) {
  //     return await apiService.search(author.display_name, institution.id).then(result => {
  //       setResults(prev => [...prev, ...result.profiles]);
  //       setLoadMoreCount(prev => prev + 1);
  //     });
  //   }

  //   return await handleDeepSearch(query, nextToken);
  // }


  const handleDeepSearch = async (query, ntoken, reset = false) => {
    if (reset) {
      setScholarResults([]);
    }
    setIsDpLoading(true)
    setTriedDp(true);

    try {
      let result;

      console.log("N token is", ntoken);
      if (ntoken !== null) {
        result = await apiService.deepSearch(query, ntoken);
      } else {
        console.log("New deep search with query", query);
        result = await apiService.deepSearch(query);
        console.log("Got results", results);
      }

      setScholarResults((prevResults) => {
        let newResults = reset ? result.profiles : [...prevResults, ...result.profiles];
        newResults = newResults.map(x => [x, levenshteinDistance(x.display_name, query)]);
        console.log(newResults);
        newResults.sort((a, b) => a[1] - b[1]);
        console.log(newResults);
        newResults = newResults.map(x => x[0]);
        return newResults;
      });

      setNextToken(result.next_page_token)

      setIsDpLoading(false)
    } catch (error) {
      console.error('Error during deep search', error);
      setIsDpLoading(false)
    }
  };

  console.log(nextToken);

  return (
    <div className='basic-setup' style={{ color: "white" }}>
      <div style={{ position: 'absolute', top: '20px', left: '20px' }}>
        <Link
          to="/entity-lookup"
          className="author-institution-search-button"
          style={{
            padding: '10px 15px',
            fontSize: '14px',
            whiteSpace: 'nowrap',
            textDecoration: 'none',
            display: 'inline-block'
          }}
        >
          Go To Entity Lookup
        </Link>
      </div>
      <img src={Logo} style={{ width: "100px", marginTop: "12%", animation: "fade-in 0.5s" }} />
      <h1 style={{ fontWeight: "bold", marginTop: 50, animation: "fade-in 0.75s" }}>Welcome to Prism</h1>
      <div style={{ animation: "fade-in 1s" }}>
        <div className='d-flex justify-content-center align-items-center'>
          <div style={{ marginTop: 10, color: "#888888" }}>We help you comply with research security requirements by automating author assessments.</div>
        </div>
        <div className='d-flex justify-content-center align-items-center'>
          <div style={{ marginTop: 10, marginBottom: 50, color: "#888888" }}>Who would you like to conduct an assessment on?</div>
        </div>
      </div>
      <div className='d-flex justify-content-center align-items-center pt-5'>
        <div style={{ width: "80%", animation: "fade-in 1.25s" }}>
          <AuthorInstiutionSearchBar onSearch={search} />
        </div>
      </div>

      {showResultHeaders && (<div style={{ paddingTop: "30px", textAlign: "center", fontSize: "24px", fontWeight: "bold" }}>OpenAlex Results</div>)}
      {isLoadingScopus && <div className="spinner-border text-primary" style={{ width: '3rem', height: '3rem' }} role="status"></div>}
      {
        hasSearched &&
        <TodoListComponent
          results={openAlexResults}
          canLoadMore={false}
          loadMore={() => { }}
        />
      }

      {showResultHeaders && <div style={{ paddingTop: "30px", textAlign: "center", fontSize: "24px", fontWeight: "bold" }}>Google Scholar Results</div>}
      {isDpLoading && <div className="spinner-border text-primary" style={{ width: '3rem', height: '3rem' }} role="status"></div>}
      {
        hasSearched &&
        <TodoListComponent
          results={scholarResults}
          canLoadMore={false}
          loadMore={() => { }}
        />
      }

      {/* {showResultHeaders && <div style={{paddingTop: "30px", textAlign: "center", fontSize: "24px", fontWeight: "bold"}}>Scopus Results</div>} */}
      {showResultHeaders && (<div style={{ paddingTop: "30px", textAlign: "center", fontSize: "24px", fontWeight: "bold" }}>{resultHeader}</div>)}
      {isLoadingScopus && <div className="spinner-border text-primary" style={{ width: '3rem', height: '3rem' }} role="status"></div>}
      {
        hasSearched &&
        <TodoListComponent
          results={results}
          canLoadMore={false}
          loadMore={() => { }}
        />
      }
    </div>
  );

};

export default SearchComponent;
