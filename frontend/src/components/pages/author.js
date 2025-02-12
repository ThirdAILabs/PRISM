import React, { useState, useEffect } from 'react';
import { autocompleteService } from '../../api/autocomplete';

const AuthorSearch = () => {
    const [query, setQuery] = useState('');
    const [suggestions, setSuggestions] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    useEffect(() => {
        if (query.length > 2) {
            setLoading(true);
            autocompleteService.autocompleteAuthors(query)
                .then(response => {
                    setSuggestions(response);
                    setLoading(false);
                })
                .catch(err => {
                    setError(err.message);
                    setLoading(false);
                });
        } else {
            setSuggestions([]);
        }
    }, [query]);

    const handleChange = (e) => {
        setQuery(e.target.value);
    };

    return (
        <div>
            <input
                type="text"
                value={query}
                onChange={handleChange}
                placeholder="Search for authors..."
            />
            {loading && <p>Loading...</p>}
            {error && <p>Error: {error}</p>}
            <ul>
                {suggestions.map(suggestion => (
                    <li key={suggestion.AuthorId} className='text-black'>{suggestion.DisplayName}</li>
                ))}
            </ul>
        </div>
    );
};

export default AuthorSearch;