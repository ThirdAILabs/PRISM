import React, { useState } from 'react';
import '../../common/searchBar/SearchBar.css';
import '../../common/tools/button/button1.css';
import Logo from '../../../assets/images/prism-logo.png';
import '../../common/searchBar/SearchBar.css';
import '../../common/tools/button/button1.css';
import { searchService } from '../../../api/search';
import './page.css';

const makeLinksClickable = (text) => {
  const urlRegex = /(https?:\/\/[^\s]+)/g;
  if (!text) return text;

  const parts = text.split(urlRegex);
  return parts.map((part, i) => {
    if (part.match(urlRegex)) {
      return (
        <a
          key={i}
          href={part}
          target="_blank"
          rel="noopener noreferrer"
          style={{ color: '#2c5282', textDecoration: 'underline' }}
        >
          {part}
        </a>
      );
    }
    return part;
  });
};

function EntityLookup() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setHasSearched(true);
    try {
      const entities = await searchService.matchEntities(query);
      console.log(entities);
      setResults(entities);
    } catch (error) {
      console.error('Error fetching data:', error);
      alert('Error fetching data: ' + error.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="entity-lookup-container">
      <div className="logo-section">
        <img src={Logo} alt="PRISM Logo" />
      </div>

      <div className="header-section">
        <h1 className="header-title">Entity Lookup</h1>
        <p className="header-subtitle">We help you comply with research security requirements.</p>
        <p className="header-subtitle" style = {{ marginTop: '10', marginBottom: '2%', color: '#888888' }}>
          Search for an entity to see if it is on any list of concerning entities.
        </p>
      </div>

      <div className="search-section">
        <form onSubmit={handleSubmit} className="author-institution-search-bar">
          <div className="autocomplete-search-bar">
            <div className="autocomplete-search-bar-title">Entity</div>
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="E.g. PQR Company"
              className="search-bar"
            />
          </div>
          <div style={{ width: '40px' }} />
          <div className="author-institution-search-button-container">
            <button type="submit" disabled={isLoading || query.length == 0} className="button">
              {isLoading ? 'Searching...' : 'Search'}
            </button>
          </div>
        </form>
      </div>

      {isLoading && (
        <div className="d-flex justify-content-center mt-4">
          <div className="spinner-border text-primary" role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
        </div>
      )}

      <div className="results-section">
        {results.length > 0
          ? results.map((entity, index) => (
              <div key={index} className="detail-item">
                <b>Names</b>
                <ul className="bulleted-list">
                  {entity.Names.split('\n').map((name, index2) => (
                    <li key={`${index}-${index2}`}>{name}</li>
                  ))}
                </ul>

                {entity.Address && (
                  <>
                    <b>Address</b>
                    <p>{entity.Address}</p>
                  </>
                )}
                {entity.Country && (
                  <>
                    <b>Country</b>
                    <p>{entity.Country}</p>
                  </>
                )}
                {entity.Type && (
                  <>
                    <b>Type</b>
                    <p>{entity.Type}</p>
                  </>
                )}
                {entity.Resource && (
                  <>
                    <b>Resource</b>
                    <p>{makeLinksClickable(entity.Resource)}</p>
                  </>
                )}
              </div>
            ))
          : !isLoading &&
            hasSearched && (
              <div className="no-results">
                <p>No results found</p>
              </div>
            )}
      </div>
    </div>
  );
}

export default EntityLookup;
