import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_BASE_URL } from '../../../services/apiService';
import "../../common/searchBar/SearchBar.css";
import "../../common/tools/button/button1.css"
import Logo from "../../../assets/images/prism-logo.png";
import { useUser } from '../../../store/userContext';
import { searchService } from '../../../api/search';

function EntityLookup() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [backendUrl, setBackendUrl] = useState('');

  useEffect(() => {
    setBackendUrl(API_BASE_URL);
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      const response = await searchService.matchEntities(query);
      console.log(response);
      const entities = response.Entities.filter((entity) => entity.trim()).map((entity) =>
        entity.replace('[ENTITY START]', '').replace('[ENTITY END]', '').trim()
      );
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
      <div style={{ position: 'absolute', top: '20px', left: '20px' }}>
        <Link
          to="/"
          className="button"
          style={{
            padding: '10px 15px',
            fontSize: '14px',
            whiteSpace: 'nowrap',
            textDecoration: 'none',
            display: 'inline-block',
          }}
        >
          Go To Individual Assessment
        </Link>
      </div>

      <div style={{ textAlign: "center", marginTop: "5.5%", animation: "fade-in 0.75s" }}>
        <img src={Logo} style={{ width: "320px", marginTop: "5%", marginBottom: "1%", marginRight: "2%", animation: "fade-in 0.5s" }} />
        <div style={{ animation: "fade-in 1s" }}>
          <div className='d-flex justify-content-center align-items-center'>
            <div style={{ marginTop: 10, color: "#888888" }}>We help you comply with research security requirements.</div>
          </div>
          <div className='d-flex justify-content-center align-items-center'>
            <div style={{ marginTop: 10, marginBottom: "2%", color: "#888888" }}>Search for an entity to see if it is on any list if concerning entities.</div>
          </div>
        </div>
        <div style={{ animation: "fade-in 1s" }}>
          <div className='d-flex justify-content-center align-items-center pt-5'>
            <div style={{ width: "80%" }}>
              <form onSubmit={handleSubmit} className="author-institution-search-bar">
                <div className="autocomplete-search-bar">
                  <input
                    type="text"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder="Enter query"
                    className="search-bar"
                  />
                </div>
                <div style={{ width: '20px' }} />
                <div style={{ width: '200px' }}>
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
          style={{ width: '3rem', height: '3rem' }}
          role="status"
        ></div>
      )}
      <div className="results" style={{ marginTop: '30px' }}>
        {results.map((entity, index) => (
          <div key={index} className="detail-item">
            <pre>{entity}</pre>
          </div>
        ))}
      </div>
    </div>
  );
}

export default EntityLookup;
