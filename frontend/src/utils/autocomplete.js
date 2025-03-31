import React, { useState } from 'react';
import '../components/common/searchBar/SearchBar.css';
import '../styles/components/_primaryButton.scss';
import TextField from '../components/common/tools/TextField';

function AutocompleteSearchBar({
  label,
  autocomplete,
  onSelect,
  showHint,
  setSearchState,
  initialValue = '',
}) {
  const [suggestions, setSuggestions] = useState([]);
  const [query, setQuery] = useState(initialValue);

  function handleInputChange(e) {
    const newValue = e.target.value;
    setQuery(newValue);
    autocomplete(newValue).then(setSuggestions);
  }

  function handleSelectSuggestion(suggestion) {
    return () => {
      setSuggestions([]);
      setQuery(suggestion.Name);
      if (setSearchState) {
        setSearchState((prev) => ({
          ...prev,
          openAlexResults: [],
          orcidResults: [],
          paperResults: [],
          hasSearched: false,
          hasSearchedOrcid: false,
          hasSearchedPaper: false,
          isOALoading: false,
          isOrcidLoading: false,
          isPaperLoading: false,
        }));
      }

      onSelect(suggestion);
    };
  }

  return (
    // Column
    <div className="autocomplete-search-bar">

      {/* Search bar */}
      <TextField
        label={label}
        variant="outlined"
        fullWidth
        value={query}
        autoComplete="off"
        onChange={handleInputChange}
      />

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
              <p style={{ marginRight: '10px' }}>{suggestion.Name}</p>{' '}
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

export default AutocompleteSearchBar;
