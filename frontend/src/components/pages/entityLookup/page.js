import React, { useState } from 'react';
import '../../common/searchBar/SearchBar.css';
import '../../common/tools/button/button1.css';
import Logo from '../../../assets/images/prism-logo.png';
import '../../common/searchBar/SearchBar.css';
import '../../common/tools/button/button1.css';
import { searchService } from '../../../api/search';

function EntityLookup() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
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
    <div className="basic-setup" style={{ color: 'white' }}>
      <div style={{ textAlign: 'center', marginTop: '5.5%', animation: 'fade-in 0.75s' }}>
        <img
          src={Logo}
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
                  <button type="submit" disabled={isLoading} className="button">
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
      <div className="results" style={{ marginTop: '30px' }}>
        {results.map((entity, index) => (
          <div key={index} className="detail-item">
            <b>Names:</b>
            <ul className="bulleted-list">
              {entity.Names.split('\n').map((name, index2) => {
                return <li key={`${index}-${index2}`}>{name}</li>;
              })}
            </ul>

            {entity.Address && entity.Address.length > 0 && (
              <>
                <b>Address:</b>
                <p>{entity.Address}</p>
              </>
            )}
            {entity.Country && entity.Country.length > 0 && (
              <>
                <b>Country:</b>
                <p>{entity.Country}</p>
              </>
            )}
            {entity.Type && entity.Type.length > 0 && (
              <>
                <b>Type:</b>
                <p>{entity.Type}</p>
              </>
            )}
            {entity.Resource && entity.Resource.length > 0 && (
              <>
                <b>Resource:</b>
                <p>{entity.Resource}</p>
              </>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

export default EntityLookup;
