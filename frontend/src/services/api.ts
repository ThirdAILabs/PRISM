// api.ts
import fetchClient from './fetchClient';

const API_BASE_URL = 'https://api.example.com';

// GET request example
export const getData = async <T = any>(): Promise<T> => {
    const url = `${API_BASE_URL}/data-endpoint`;
    return fetchClient<T>(url, {
        method: 'GET'
    });
};

// POST request example
export const postData = async <T = any>(payload: any): Promise<T> => {
    const url = `${API_BASE_URL}/data-endpoint`;
    return fetchClient<T>(url, {
        method: 'POST',
        body: JSON.stringify(payload)
    });
};

// PUT or DELETE functions can be added in a similar fashion
