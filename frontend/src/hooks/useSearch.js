import { useState, useEffect } from 'react';
import { searchOpenAlexAuthors, searchGoogleScholarAuthors } from '../api/search';

const useSearch = () => {
    const [query, setQuery] = useState('');
    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const handleSearch = async () => {
        setLoading(true);
        setError(null);
        try {
            const openAlexResults = await searchOpenAlexAuthors(query);
            const googleScholarResults = await searchGoogleScholarAuthors(query);
            setResults([...openAlexResults, ...googleScholarResults]);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (query) {
            handleSearch();
        } else {
            setResults([]);
        }
    }, [query]);

    return { query, setQuery, results, loading, error };
};

export default useSearch;