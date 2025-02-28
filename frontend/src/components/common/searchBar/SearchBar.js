import React, { useState, useCallback } from 'react';
import { autocompleteService } from '../../../api/autocomplete';
import './SearchBar.css';
import '../tools/button/button1.css';
import useCallOnPause from '../../../hooks/useCallOnPause';

function AutocompleteSearchBar({ title, autocomplete, onSelect, showHint }) {
  const [suggestions, setSuggestions] = useState([]);
  const [query, setQuery] = useState('');

  function handleInputChange(e) {
    setQuery(e.target.value);
    autocomplete(e.target.value).then(setSuggestions);
  }

  function handleSelectSuggestion(suggestion) {
    return () => {
      setSuggestions([]);
      setQuery(suggestion.Name);
      onSelect(suggestion);
    };
  }

  return (
    // Column
    <div className="autocomplete-search-bar">
      {/* Header 2 or 3, bold */}
      <div className="autocomplete-search-bar-title">{title}</div>

      {/* Search bar */}
      <input type="text" className="search-bar" value={query} onChange={handleInputChange} />

      {query && query.length && suggestions && suggestions.length > 0 && (
        // Autocomplete suggestion container. Column.
        <div className="suggestion-container">
          {suggestions.map((suggestion, index) => (
            // Clickable suggestion
            <div
              className="suggestion"
              key={index}
              onClick={handleSelectSuggestion(suggestion)}
              style={{ display: 'flex', alignItems: 'end' }}
            >
              <p style={{ marginRight: '20px' }}>{suggestion.Name}</p>{' '}
              {showHint && (
                <p style={{ marginBottom: '16.5px', fontSize: 'small', fontStyle: 'italic' }}>
                  {suggestion.Hint}
                </p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export function AuthorInstiutionSearchBar({ onSearch }) {
  const [author, setAuthor] = useState();
  const [institution, setInstitution] = useState();
  const [results, setResults] = useState([]);
  const debouncedSearch = useCallOnPause(300); // 300ms delay

  const autocompleteInstitution = useCallback(
    (query) => {
      return new Promise((resolve) => {
        debouncedSearch(async () => {
          try {
            const res = await autocompleteService.autocompleteInstitutions(query);
            resolve(res);
            return res;
          } catch (error) {
            console.error('Autocomplete error:', error);
            resolve([]);
          }
        });
      });
    },
    [debouncedSearch]
  );

  const autocompleteAuthor = useCallback(
    (query) => {
      return new Promise((resolve) => {
        debouncedSearch(async () => {
          try {
            const res = await autocompleteService.autocompleteAuthors(query);
            resolve(res);
            return res;
          } catch (error) {
            console.error('Autocomplete error:', error);
            resolve([]);
          }
        });
      });
    },
    [debouncedSearch]
  );

  function search() {
    if (author && institution) {
      onSearch(author, institution);
    } else {
      alert('Please select an author and institution.');
    }
  }

  return (
    <div className="author-institution-search-bar">
      <div className="author-institution-search-bar-container">
        <AutocompleteSearchBar
          title="Author"
          autocomplete={autocompleteAuthor}
          onSelect={setAuthor}
          showHint={false}
        />
      </div>

      <div className="author-institution-search-bar-container">
        <AutocompleteSearchBar
          title="Institution"
          autocomplete={autocompleteInstitution}
          onSelect={setInstitution}
          showHint={true}
        />
      </div>

      <div className="author-institution-search-button-container">
        <button className="button" onClick={search}>
          Search
        </button>
      </div>
    </div>
  );
}
