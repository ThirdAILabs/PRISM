import React, { useState } from 'react';
import "./SearchBar.css";
import apiService from './api/apiService';

function AutocompleteSearchBar({title, autocomplete, onSelect}) {
    const [suggestions, setSuggestions] = useState([]);
    const [query, setQuery] = useState("");

    function handleInputChange(e) {
        setQuery(e.target.value);
        autocomplete(e.target.value).then(setSuggestions);
    }

    function handleSelectSuggestion(suggestion) {
        return () => {
            setSuggestions([]);
            setQuery(suggestion.display_name);
            onSelect(suggestion);
        }
    }

    return (
        // Column
        <div className='autocomplete-search-bar'>
            {/* Header 2 or 3, bold */}
            <div className='autocomplete-search-bar-title'>{title}</div>
            
            {/* Search bar */}
            <input type='text' className='search-bar' value={query} onChange={handleInputChange}/>
            
            {
                suggestions && suggestions.length > 0 &&
                // Autocomplete suggestion container. Column.
                <div className='suggestion-container'>
                    {suggestions.map((suggestion, index) => (
                        // Clickable suggestion
                        <div className='suggestion' key={index} onClick={handleSelectSuggestion(suggestion)}>{suggestion.display_name}</div>
                    ))}
                </div>
            }
        </div>
    );
}

export function AuthorInstiutionSearchBar({onSearch}) {
    const [author, setAuthor] = useState();
    const [institution, setInstitution] = useState();

    function autocompleteAuthor(query) {
        return apiService.autocomplete(query).then(res => Array.from(new Set(res.profiles.map(p => p.display_name))).map(n => ({display_name: n})));
    }

    function autocompleteInstitution(query) {
        return apiService.autocompleteInstitution(query).then(res => res.profiles);
    }

    function search() {
        onSearch(author, institution);
        // if (author && institution) {
        //     onSearch(author, institution);
        // } else {
        //     alert("Please select an author and institution.");
        // }
    }

    return <div className='author-institution-search-bar'>
        <div className='author-institution-search-bar-container'>
            <AutocompleteSearchBar title="Author" autocomplete={autocompleteAuthor} onSelect={setAuthor} />
        </div>
        
        <div className='author-institution-search-bar-container'>
            <AutocompleteSearchBar title="Institution" autocomplete={autocompleteInstitution} onSelect={setInstitution} />
        </div>

        <div className='author-institution-search-button-container'>
            <button className='author-institution-search-button' onClick={search}>Search</button>
        </div>
    </div>
}