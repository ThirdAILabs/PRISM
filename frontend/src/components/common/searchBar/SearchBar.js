import React, { useState } from 'react';
import { autocompleteService } from '../../../api/autocomplete';
import "./SearchBar.css";


function AutocompleteSearchBar({ title, autocomplete, onSelect, type }) {
    const [suggestions, setSuggestions] = useState([]);
    const [query, setQuery] = useState("");

    function handleInputChange(e) {
        setQuery(e.target.value);
        autocomplete(e.target.value).then(setSuggestions);
    }

    function handleSelectSuggestion(suggestion) {
        return () => {
            setSuggestions([]);
            setQuery(suggestion.AuthorName);
            onSelect(suggestion);
        }
    }

    return (
        // Column
        <div className='autocomplete-search-bar'>
            {/* Header 2 or 3, bold */}
            <div className='autocomplete-search-bar-title'>{title}</div>

            {/* Search bar */}
            <input type='text' className='search-bar' value={query} onChange={handleInputChange} />

            {
                query && query.length && suggestions && suggestions.length > 0 &&
                // Autocomplete suggestion container. Column.
                <div className='suggestion-container'>
                    {suggestions.map((suggestion, index) => (
                        // Clickable suggestion
                        type === "author" ?
                            <div className='suggestion' key={index} onClick={handleSelectSuggestion(suggestion)}>{suggestion.AuthorName}</div>
                            :
                            <div className='suggestion' key={index} onClick={handleSelectSuggestion(suggestion)}>{suggestion.InstitutionName}</div>
                    ))}
                </div>
            }
        </div>
    );
}

export function AuthorInstiutionSearchBar({ onSearch }) {
    const [author, setAuthor] = useState();
    const [institution, setInstitution] = useState();

    async function autocompleteAuthor(query) {
        const res = await autocompleteService.autocompleteAuthors(query);
        return res;
    }

    async function autocompleteInstitution(query) {
        const res = await autocompleteService.autocompleteInstitutions(query);
        return res;
    }

    function search() {
        if (author && institution) {
            onSearch(author, institution);
        } else {
            alert("Please select an author and institution.");
        }
    }

    return <div className='author-institution-search-bar'>
        <div className='author-institution-search-bar-container'>
            <AutocompleteSearchBar title="Author" autocomplete={autocompleteAuthor} onSelect={setAuthor} type={"author"} />
        </div>

        <div className='author-institution-search-bar-container'>
            <AutocompleteSearchBar title="Institution" autocomplete={autocompleteInstitution} onSelect={setInstitution} type={"institute"} />
        </div>

        <div className='author-institution-search-button-container'>
            <button className='author-institution-search-button' onClick={search}>Search</button>
        </div>
    </div>
}