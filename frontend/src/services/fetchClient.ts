// fetchClient.ts
export interface FetchConfig extends RequestInit {
    // You can extend this interface to include additional configuration options if needed
}

const fetchClient = async <T>(
    url: RequestInfo,
    config?: FetchConfig
): Promise<T> => {
    // Merge default configuration with any passed-in options
    const defaultConfig: RequestInit = {
        headers: {
            'Content-Type': 'application/json'
        }
    };

    // You may want to attach authentication tokens here, for example:
    const token = localStorage.getItem('token');
    if (token) {
        defaultConfig.headers = {
            ...defaultConfig.headers,
            'Authorization': `Bearer ${token}`
        };
    }

    const mergedConfig = { ...defaultConfig, ...config };

    const response = await fetch(url, mergedConfig);

    if (!response.ok) {
        // Here, you could process error responses further
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || 'An error occurred while fetching data.');
    }

    return response.json();
};

export default fetchClient;
