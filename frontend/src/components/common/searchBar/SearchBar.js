import React, { useState, useCallback } from 'react';
import { autocompleteService } from '../../../api/autocomplete';
import './SearchBar.css';
import '../tools/button/button1.css';
import useCallOnPause from '../../../hooks/useCallOnPause';
import AutocompleteSearchBar from '../../../utils/autocomplete';
// function AutocompleteSearchBar({ title, autocomplete, onSelect, type }) {
//   const [suggestions, setSuggestions] = useState([]);
//   const [query, setQuery] = useState('');

<<<<<<< HEAD
//   function handleInputChange(e) {
//     setQuery(e.target.value);
//     autocomplete(e.target.value).then(setSuggestions);
//   }
=======
function AutocompleteSearchBar({ title, autocomplete, onSelect, type, placeholder }) {
  const [suggestions, setSuggestions] = useState([]);
  const [query, setQuery] = useState('');
>>>>>>> origin/main

//   function handleSelectSuggestion(suggestion) {
//     return () => {
//       setSuggestions([]);
//       setQuery(type === 'author' ? suggestion.AuthorName : suggestion.InstitutionName);
//       onSelect(suggestion);
//     };
//   }

//   function handleSelectSuggestionInstitute(suggestion) {
//     return () => {
//       setSuggestions([]);
//       setQuery(suggestion.InstitutionName);
//       onSelect(suggestion);
//     };
//   }

//   return (
//     // Column
//     <div className="autocomplete-search-bar">
//       {/* Header 2 or 3, bold */}
//       <div className="autocomplete-search-bar-title">{title}</div>

//       {/* Search bar */}
//       <input type="text" className="search-bar" value={query} onChange={handleInputChange} />

<<<<<<< HEAD
//       {query && query.length && suggestions && suggestions.length > 0 && (
//         // Autocomplete suggestion container. Column.
//         <div className="suggestion-container">
//           {suggestions.map((suggestion, index) =>
//             // Clickable suggestion
//             type === 'author' ? (
//               <div className="suggestion" key={index} onClick={handleSelectSuggestion(suggestion)}>
//                 {suggestion.AuthorName}
//               </div>
//             ) : (
//               <div
//                 className="suggestion"
//                 key={index}
//                 onClick={handleSelectSuggestionInstitute(suggestion)}
//               >
//                 {suggestion.InstitutionName}
//               </div>
//             )
//           )}
//         </div>
//       )}
//     </div>
//   );
// }
=======
      {/* Search bar */}
      <input
        type="text"
        className="search-bar"
        placeholder={placeholder || ''}
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
>>>>>>> origin/main

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
          type={'author'}
          placeholder={'E.g. John Doe'}
        />
      </div>

      <div className="author-institution-search-bar-container">
        <AutocompleteSearchBar
          title="Institution"
          autocomplete={autocompleteInstitution}
          onSelect={setInstitution}
          type={'institute'}
          placeholder={'E.g. University of Prism'}
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
