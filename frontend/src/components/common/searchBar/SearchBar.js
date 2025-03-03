import React, { useState, useEffect, useCallback } from 'react';
import { autocompleteService } from '../../../api/autocomplete';
import './SearchBar.css';
import '../tools/button/button1.css';
import useCallOnPause from '../../../hooks/useCallOnPause';
import AutocompleteSearchBar from '../../../utils/autocomplete';

export function AuthorInstiutionSearchBar({ onSearch, defaultAuthor, defaultInstitution }) {
  const [author, setAuthor] = useState(defaultAuthor || null);
  const [institution, setInstitution] = useState(defaultInstitution || null);
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
          initialValue={defaultAuthor ? defaultAuthor.AuthorName : ''}
        />
      </div>

      <div className="author-institution-search-bar-container">
        <AutocompleteSearchBar
          title="Institution"
          autocomplete={autocompleteInstitution}
          onSelect={setInstitution}
          type={'institute'}
          placeholder={'E.g. University of Prism'}
          initialValue={defaultInstitution ? defaultInstitution.InstitutionName : ''}
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
