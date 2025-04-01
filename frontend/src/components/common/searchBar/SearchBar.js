import React, { useState, useEffect, useCallback, useContext } from 'react';
import { SearchContext } from '../../../store/searchContext';
import { autocompleteService } from '../../../api/autocomplete';
import './SearchBar.css';
import '../../../styles/components/_primaryButton.scss';
import useCallOnPause from '../../../hooks/useCallOnPause';
import AutocompleteSearchBar from '../../../utils/autocomplete';
// import { TextField } from '@mui/material';
import TextField from '../tools/TextField';
import '../../../styles/components/_primaryButton2.scss';
import { Tooltip } from '@mui/material';

export function AuthorInstiutionSearchBar({ onSearch, defaultAuthor, defaultInstitution }) {
  const [author, setAuthor] = useState(defaultAuthor || null);
  const [institution, setInstitution] = useState(defaultInstitution || null);
  const { searchState, setSearchState } = useContext(SearchContext);
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
          label="Enter Author Name"
          autocomplete={autocompleteAuthor}
          onSelect={setAuthor}
          setSearchState={setSearchState}
          showHint={false}
          initialValue={defaultAuthor ? defaultAuthor.Name : ''}
        />
      </div>

      <div className="author-institution-search-bar-container">
        <AutocompleteSearchBar
          label="Enter Institution Name"
          autocomplete={autocompleteInstitution}
          onSelect={setInstitution}
          setSearchState={setSearchState}
          showHint={true}
          initialValue={defaultInstitution ? defaultInstitution.Name : ''}
        />
      </div>

      <div className="author-institution-search-button-container">
        <button class="button button-3d" onClick={search} disabled={!author || !institution}>
          Search
        </button>
      </div>
    </div>
  );
}

export function SingleSearchBar({
  label,
  onSearch,
  initialValue = '',
}) {
  const orcidRegex = /^[0-9]{4}-[0-9]{4}-[0-9]{4}-[0-9X]{4}$/;
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
          {/* {title && <label className="single-search-bar-label">{title}</label>} */}

          {/* Row containing input and button side-by-side */}
          <div className="single-search-row">
            <div className="single-search-bar-container">
              <form onSubmit={handleSubmit}>

                <TextField
                  label={label}
                  variant="outlined"
                  fullWidth
                  autoComplete="off"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                />
              </form>
            </div>

            <div className="single-search-button-container" style={{ marginTop: '-2px' }}>
              <Tooltip title="Please enter a valid ORCID ID in the format 0000-0000-0000-0000 (last digit can be 0-9 or X).">
                <button className="button button-3d" onClick={handleSubmit} disabled={!orcidRegex.test(value)}>
                  Search
                </button>
              </Tooltip>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
