import React, { useState } from 'react';
import '../components/common/searchBar/SearchBar.css';
import '../components/common/tools/button/button1.css';

function AutocompleteSearchBar({ title, autocomplete, onSelect, type }) {
  const [suggestions, setSuggestions] = useState([]);
  const [query, setQuery] = useState('');

  function handleInputChange(e) {
    setQuery(e.target.value);
    autocomplete(e.target.value).then(setSuggestions);
  }

  function handleSelectSuggestion(suggestion) {
    return () => {
      setSuggestions([]);
      setQuery(type === 'author' ? suggestion.AuthorName : suggestion.InstitutionName);
      onSelect(suggestion);
    };
  }

  function handleSelectSuggestionInstitute(suggestion) {
    return () => {
      setSuggestions([]);
      setQuery(suggestion.InstitutionName);
      onSelect(suggestion);
    };
  }

  return (
    // Column
    <div className="autocomplete-search-bar">
      {/* Header 2 or 3, bold */}
      <div className="autocomplete-search-bar-title">{title}</div>

      {/* Search bar */}
      <input type="text"
        placeholder={title === 'Author' ? 'E.g. John Doe' : 'E.g. University of XYZ'}
        className="search-bar"
        value={query} onChange={handleInputChange} />

      {query && query.length && suggestions && suggestions.length > 0 && (
        // Autocomplete suggestion container. Column.
        <div className="suggestion-container">
          {suggestions.map((suggestion, index) =>
            // Clickable suggestion
            type === 'author' ? (
              <div className="suggestion" key={index} onClick={handleSelectSuggestion(suggestion)}>
                {suggestion.AuthorName}
              </div>
            ) : (
              <div
                className="suggestion"
                key={index}
                onClick={handleSelectSuggestionInstitute(suggestion)}
              >
                {suggestion.InstitutionName}
              </div>
            )
          )}
        </div>
      )}
    </div>
  );
}

export default AutocompleteSearchBar;
