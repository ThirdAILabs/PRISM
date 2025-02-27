// src/api/apiService.jsx
import axios from 'axios';

let API_BASE_URL = window.location.origin;

console.log(API_BASE_URL);

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

class MaxCount {
  constructor() {
    this.counts = [1, 4, 8, 16, 32, 64, 128, 200];
    this.calls = 0;
  }

  next() {
    if (this.calls < this.counts.length) {
      const count = this.counts[this.calls];
      this.calls += 1;
      return count;
    }
    return 200;
  }
}

const apiService = {
  autocomplete: async (query) => {
    try {
      console.log(`autocomplete: ${query}`);
      const response = await axios.get(`${API_BASE_URL}/autocomplete`, {
        params: { query },
      });
      return response.data;
    } catch (error) {
      console.error('Error calling autocomplete API', error);
      throw error;
    }
  },

  autocompleteInstitution: async (query) => {
    try {
      const response = await axios.get(`${API_BASE_URL}/autocomplete_institution`, {
        params: { query },
      });
      return response.data;
    } catch (error) {
      console.error('Error calling autocomplete institution API', error);
      throw error;
    }
  },

  search: async (author_name, institution_name) => {
    const [givenName, lastName] = author_name;
    let query = `authlast(${lastName})`;
    if (givenName) {
      query += ` and authfirst(${givenName})`;
    }
    if (institution_name) {
      query += ` and affil(${institution_name})`;
    }
    try {
      const response = await axios.get(`https://api.elsevier.com/content/search/author`, {
        headers: { 'Content-Type': 'application/json' },
        params: { query, apiKey: 'c20d01f998041241c36b6e9e6d9d6402' },
      });
      if (
        response.data['search-results']['entry'].length === 1 &&
        !response.data['search-results']['entry'][0]['dc:identifier']
      ) {
        return { profiles: [] };
      }
      const profiles = response.data['search-results']['entry'].map((entry) => ({
        id: entry['dc:identifier'].split(':')[1],
        display_name:
          entry['preferred-name']['given-name'] + ' ' + entry['preferred-name']['surname'],
        institutions: [entry['affiliation-current']['affiliation-name']],
        source: 'scopus',
        works_count: null,
      }));
      return { profiles };
    } catch (error) {
      console.error('Error calling search API', error);
      throw error;
    }
  },

  searchOpenAlex: async (author_name, institution_name) => {
    const [givenName, lastName] = author_name;

    let authorName = `${givenName} ${lastName}`;
    let institutionID;

    try {
      if (institution_name) {
        const response = await apiService.autocompleteInstitution(institution_name);
        institutionID = response.profiles[0].id.split('/').pop();
      }

      const response = await axios.get(`${API_BASE_URL}/search`, {
        params: { author_name: authorName, institution_id: institutionID },
      });

      console.log('response.data-> ', response.data);

      return response.data;
    } catch (error) {
      console.error('Error calling search API', error);
      throw error;
    }
  },

  deepSearch: async (query, nextPageToken = null) => {
    try {
      const params = { query };
      if (nextPageToken) params.next_page_token = nextPageToken;
      const response = await axios.get(`${API_BASE_URL}/deep_search`, { params });
      return response.data;
    } catch (error) {
      console.error('Error calling deep search API', error);
      throw error;
    }
  },

  formalRelations: async (author, institution) => {
    try {
      const params = { author, institution };
      const response = await axios.get(`${API_BASE_URL}/formal_relations`, { params });
      return response.data;
    } catch (error) {
      console.error('Error calling formal relations API', error);
      throw error;
    }
  },

  scopusPaperTitles: async (authorId, onLoadTitleBatch) => {
    console.log('Getting paper titles...');
    const query = `au-id(${authorId})`;
    const maxCountGenerator = new MaxCount();
    let maxCount = maxCountGenerator.next();
    let startIndex = 0;
    const response = await axios.get(`https://api.elsevier.com/content/search/scopus`, {
      headers: { 'Content-Type': 'application/json' },
      params: {
        query,
        apiKey: 'c20d01f998041241c36b6e9e6d9d6402',
        count: maxCount,
        start: startIndex,
      },
    });

    const numWorks = Number.parseInt(response.data['search-results']['opensearch:totalResults']);
    console.log(numWorks);
    const titleBatch = response.data['search-results']['entry'].map((entry) => entry['dc:title']);
    console.log('Got first batch...');
    onLoadTitleBatch(titleBatch);
    let titles = titleBatch;
    if (numWorks > maxCount) {
      let promises = [];
      startIndex += maxCount;
      maxCount = maxCountGenerator.next();
      while (startIndex < numWorks) {
        const promise = axios
          .get(`https://api.elsevier.com/content/search/scopus`, {
            headers: { 'Content-Type': 'application/json' },
            params: {
              query,
              apiKey: 'c20d01f998041241c36b6e9e6d9d6402',
              count: maxCount,
              start: startIndex,
            },
          })
          .then((response) => {
            const titleBatch = response.data['search-results']['entry'].map(
              (entry) => entry['dc:title']
            );
            onLoadTitleBatch(titleBatch);
            return titleBatch;
          });
        promises.push(promise);
        await sleep(600);
        startIndex += maxCount;
        maxCount = maxCountGenerator.next();
      }
      const titleBatches = await Promise.all(promises);
      for (const batch of titleBatches) {
        titles = titles.concat(batch);
      }
    }
    return titles;
  },
};

export default apiService;
export { API_BASE_URL };
