import React, { useState } from 'react';
import '../../common/searchBar/SearchBar.css';
import '../../common/tools/button/button1.css';
import './entityLookup.css';
import Logo from '../../../assets/images/prism-logo.png';
import { searchService } from '../../../api/search';

const makeLinksClickable = (text) => {
  const urlRegex = /((?:http|https):\/\/[^\s]+)/g;
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
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="basic-setup" style={{ color: 'white' }}>
      <div style={{ textAlign: 'center', marginTop: '5.5%', animation: 'fade-in 0.75s' }}>
        <img
          src={Logo}
          alt="Logo"
          style={{
            width: '320px',
            marginTop: '5%',
            marginBottom: '1%',
            marginRight: '2%',
            animation: 'fade-in 0.5s',
          }}
        />
        <div style={{ animation: 'fade-in 1s' }}>
          <div className="d-flex justify-content-center align-items-center">
            <div style={{ color: '#888888' }}>
              <h1
                style={{
                  marginTop: 20,
                  fontWeight: 'bold',
                  color: 'black',
                  animation: 'fade-in 0.75s',
                }}
              >
                Entity Lookup
              </h1>
              We help you comply with research security requirements.
            </div>
          </div>
          <div className="d-flex justify-content-center align-items-center">
            <div style={{ marginTop: 10, marginBottom: '0%', color: '#888888' }}>
              Search for an entity to see if it is on any list of concerning entities.
            </div>
          </div>
        </div>
        <div style={{ animation: 'fade-in 1s' }}>
          <div className="d-flex justify-content-center align-items-center pt-5">
            <div style={{ width: '80%' }}>
              <form onSubmit={handleSubmit} className="author-institution-search-bar">
                <div className="entity-lookup-search-bar-container">
                  <div className="autocomplete-search-bar-title">Entity</div>
                  <input
                    type="text"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder="E.g. PQR Company"
                    className="search-bar"
                  />
                </div>
                <div className="author-institution-search-button-container">
                  <button
                    type="submit"
                    disabled={isLoading || query.length === 0}
                    className="button"
                  >
                    {isLoading ? 'Searching...' : 'Search'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>
      {isLoading && (
        <div
          className="spinner-border text-primary"
          style={{ width: '3rem', height: '3rem', marginTop: '20px' }}
          role="status"
        ></div>
      )}

      {results.length > 0 ? (
        <div className="entity-lookup-results">
          {results.map((entity, index) => (
            <div key={index} className="entity-lookup-items">
              <b>Names:</b>
              <ul className="bulleted-list">
                {entity.Names.split('\n').map((name, index2) => (
                  <li key={`${index}-${index2}`}>{name}</li>
                ))}
              </ul>

              {entity.Address && (
                <>
                  <b>Address:</b>
                  <p>{entity.Address}</p>
                </>
              )}
              {entity.Country && (
                <>
                  <b>Country:</b>
                  <p>{entity.Country}</p>
                </>
              )}
              {entity.Type && (
                <>
                  <b>Type:</b>
                  <p>{entity.Type}</p>
                </>
              )}
              {entity.Resource && (
                <>
                  <b>Resource:</b>
                  <p>{makeLinksClickable(entity.Resource)}</p>
                </>
              )}
            </div>
          ))}
        </div>
      ) : (
        !isLoading &&
        hasSearched && (
          <div className="no-results">
            <div className="no-results-icon">🔍</div>
            <h3>We couldn't find any results</h3>
            <p>Try adjusting your search to find what you're looking for.</p>
          </div>
        )
      )}
    </div>
  );
}

export default EntityLookup;
