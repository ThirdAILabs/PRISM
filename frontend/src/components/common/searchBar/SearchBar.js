import React, { useState, useEffect, useCallback, useContext } from 'react';
import { SearchContext } from '../../../store/searchContext';
import { autocompleteService } from '../../../api/autocomplete';
import './SearchBar.css';
import '../tools/button/button1.css';
import useCallOnPause from '../../../hooks/useCallOnPause';
import AutocompleteSearchBar from '../../../utils/autocomplete';

export function AuthorInstiutionSearchBar({ onSearch, defaultAuthor, defaultInstitution }) {
  const [author, setAuthor] = useState(defaultAuthor || null);
  const [institution, setInstitution] = useState(defaultInstitution || null);
  const { searchState, setSearchState } = useContext(SearchContext);
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
          setSearchState={setSearchState}
          showHint={false}
          placeholder={'E.g. John Doe'}
          initialValue={defaultAuthor ? defaultAuthor.Name : ''}
        />
      </div>

      <div className="author-institution-search-bar-container">
        <AutocompleteSearchBar
          title="Institution"
          autocomplete={autocompleteInstitution}
          onSelect={setInstitution}
          setSearchState={setSearchState}
          showHint={true}
          placeholder={'E.g. University of XYZ'}
          initialValue={defaultInstitution ? defaultInstitution.Name : ''}
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

export function SingleSearchBar({
  title = '',
  onSearch,
  placeholder = 'Enter Value',
  initialValue = '',
}) {
  const [value, setValue] = useState(initialValue);

  const handleSubmit = (e) => {
    e.preventDefault();
    if (value.trim()) {
      onSearch(value);
    }
  };

  return (
    <div style={{ textAlign: 'center', marginTop: '3%' }}>
      <div style={{ marginTop: '1rem' }}>
        <div className="single-search-container">
          {/* Same large title style as Author/Institution */}
          {title && <label className="single-search-bar-label">{title}</label>}

          {/* Row containing input and button side-by-side */}
          <div className="single-search-row">
            <div className="single-search-bar-container">
              <form onSubmit={handleSubmit}>
                <input
                  type="text"
                  className="search-bar"
                  placeholder={placeholder}
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                />
              </form>
            </div>

            <div className="single-search-button-container">
              <button className="button" onClick={handleSubmit}>
                Search
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
