import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../../../services/apiService';
import "../../common/searchBar/SearchBar.css";

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
      const response = await axios.get(`${backendUrl}/match?query=${encodeURIComponent(query)}`);
      const entities = response.data.result.split('[ENTITY START]')
        .filter(entity => entity.trim())
        .map(entity => entity.split('[ENTITY END]')[0].trim());
      setResults(entities);
    } catch (error) {
      console.error('Error fetching data:', error);
      alert('Error fetching data: ' + error.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="basic-setup" style={{ color: "white" }}>
      <div style={{ position: 'absolute', top: '20px', left: '20px' }}>
        <Link
          to="/"
          className="author-institution-search-button"
          style={{
            padding: '10px 15px',
            fontSize: '14px',
            whiteSpace: 'nowrap',
            textDecoration: 'none',
            display: 'inline-block'
          }}
        >
          Go To Individual Assessment
        </Link>
      </div>
      <h1 style={{ marginTop: '50px', fontWeight: 'bold' }}>Entity Lookup</h1>
      <div className='d-flex justify-content-center align-items-center pt-5'>
        <div style={{ width: "80%" }}>
          <form onSubmit={handleSubmit} className="author-institution-search-bar">
            <div className='autocomplete-search-bar'>
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Enter query"
                className="search-bar"
              />
            </div>
            <div style={{ width: "20px" }} />
            <div style={{ width: "200px" }}>
              <button type="submit" disabled={isLoading} className="author-institution-search-button">
                {isLoading ? 'Searching...' : 'Search'}
              </button>
            </div>
          </form>
        </div>
      </div>
      {isLoading && <div className="spinner-border text-primary" style={{ width: '3rem', height: '3rem' }} role="status"></div>}
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