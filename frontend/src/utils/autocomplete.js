import React, { useState } from 'react';
import '../components/common/searchBar/SearchBar.css';
import '../components/common/tools/button/button1.css';

function AutocompleteSearchBar({
  title,
  autocomplete,
  onSelect,
  type,
  placeholder,
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
      setQuery(type === 'author' ? suggestion.AuthorName : suggestion.InstitutionName);
      onSelect(suggestion);
    };
  }

  return (
    // Column
    <div className="autocomplete-search-bar">
      {/* Header 2 or 3, bold */}
      <div className="autocomplete-search-bar-title">{title}</div>

      {/* Search bar */}
      <input
        type="text"
        placeholder={placeholder}
        className="search-bar"
        value={query}
        onChange={handleInputChange}
      />

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
              <div className="suggestion" key={index} onClick={handleSelectSuggestion(suggestion)}>
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
