// src/api/apiService.jsx
import axios from 'axios';

let API_BASE_URL;
if (process.env.REACT_APP_BACKEND_ORIGIN === 'match') {
  API_BASE_URL = window.location.origin.replace(":" + window.location.port, ":" + process.env.REACT_APP_BACKEND_PORT);
} else {
  API_BASE_URL = process.env.REACT_APP_BACKEND_ORIGIN;
}
console.log(API_BASE_URL);

const sleep = ms => new Promise(r => setTimeout(r, ms));

class MaxCount {
  constructor() {
    this.counts = [1, 4, 8, 16, 32, 64, 128, 200];
    this.calls = 0;
  }

  next() {
    if (this.calls < this.counts.length) {
      const count = this.counts[this.calls]
      this.calls += 1;
      return count;
    }
    return 200;
  }
};

const apiService = {
  autocomplete: async (query) => {
    try {
      console.log(`autocomplete: ${query}`)
      const response = await axios.get(`${API_BASE_URL}/autocomplete`, {
        params: { query }
      });
      return response.data;
    } catch (error) {
      console.error('Error calling autocomplete API', error);
      throw error;
    }
  }, 

  // autocomplete: async (query) => {
  //   try {
  //     console.log(`autocomplete: ${query}`);
  //     const words = query.trim().split(/\s+/);
      
  //     // If query has exactly two words, search both combinations
  //     if (words.length === 2) {
  //       const [first, second] = words;
  //       const normalQuery = `${first} ${second}`;
  //       const reversedQuery = `${second} ${first}`;
        
  //       // Make parallel requests for both combinations
  //       const [normalResponse, reversedResponse] = await Promise.all([
  //         axios.get(`${API_BASE_URL}/autocomplete`, {
  //           params: { query: normalQuery }
  //         }),
  //         axios.get(`${API_BASE_URL}/autocomplete`, {
  //           params: { query: reversedQuery }
  //         })
  //       ]);
        
  //       // Combine and deduplicate results
  //       const combinedProfiles = [
  //         ...normalResponse.data.profiles,
  //         ...reversedResponse.data.profiles
  //       ];
        
  //       return {
  //         profiles: combinedProfiles
  //       };
  //     }
      
  //     // For single word or 3+ words, use original behavior
  //     const response = await axios.get(`${API_BASE_URL}/autocomplete`, {
  //       params: { query }
  //     });
  //     return response.data;
      
  //   } catch (error) {
  //     console.error('Error calling autocomplete API', error);
  //     throw error;
  //   }
  // },
  
  autocompleteInstitution: async (query) => {
    try {
      const response = await axios.get(`${API_BASE_URL}/autocomplete_institution`, {
        params: { query }
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
        params: { query, apiKey: "c20d01f998041241c36b6e9e6d9d6402" }
      });
      if (response.data['search-results']['entry'].length === 1 && !response.data['search-results']['entry'][0]['dc:identifier']) {
        return { profiles: [] };
      }
      const profiles = response.data['search-results']['entry'].map(entry => ({
        id: entry['dc:identifier'].split(':')[1],
        display_name: entry['preferred-name']['given-name'] + ' ' + entry['preferred-name']['surname'],
        institutions: [entry['affiliation-current']['affiliation-name']],
        source: "scopus",
        works_count: null,
      }));
      return { profiles };
    } catch (error) {
      console.error('Error calling search API', error);
      throw error;
    }
  },

  // TODO: Scopus name reversing
  // search: async (author_name, institution_name) => {
  //   const [givenName, lastName] = author_name;
  
  //   // Function to fetch profiles for a given query
  //   const fetchProfiles = async (query) => {
  //     try {
  //       const response = await axios.get(`https://api.elsevier.com/content/search/author`, {
  //         headers: { 'Content-Type': 'application/json' },
  //         params: { query, apiKey: "c20d01f998041241c36b6e9e6d9d6402" }
  //       });
  
  //       if (response.data['search-results']['entry'].length === 1 && 
  //           !response.data['search-results']['entry'][0]['dc:identifier']) {
  //         return [];
  //       }
  
  //       return response.data['search-results']['entry'].map(entry => ({
  //         id: entry['dc:identifier'].split(':')[1],
  //         display_name: entry['preferred-name']['given-name'] + ' ' + entry['preferred-name']['surname'],
  //         institutions: [entry['affiliation-current']['affiliation-name']],
  //         source: "scopus",
  //         works_count: null,
  //       }));
  //     } catch (error) {
  //       console.error('Error calling search API', error);
  //       throw error;
  //     }
  //   };
  
  //   try {
  //     let query1 = `authlast(${lastName})`; 
  //     let query2 = '';
  //       if (givenName) {
  //         query1 += ` and authfirst(${givenName})`;
  //         query2 = `authfirst(${givenName}) and `;
  //       }
  //     query2 += `authlast(${lastName})`;
      
  //     if (institution_name) {
  //       query1 += ` and affil(${institution_name})`;
  //       query2 += ` and affil(${institution_name})`;
  //     }
  
  //     // Fetch results from both queries
  //     const [profiles1, profiles2] = await Promise.all([
  //       fetchProfiles(query1),
  //       fetchProfiles(query2)
  //     ]);
  
  //     // Combine results and remove duplicates based on id
  //     const allProfiles = [...profiles1, ...profiles2];
  //     const uniqueProfiles = allProfiles.filter((profile, index, self) =>
  //       index === self.findIndex((p) => p.id === profile.id)
  //     );
  
  //     return { profiles: uniqueProfiles };
  //   } catch (error) {
  //     console.error('Error in search', error);
  //     throw error;
  //   }
  // },

  searchOpenAlex: async (author_name, institution_name) => {
    const [givenName, lastName] = author_name;
    
    let authorName = `${givenName} ${lastName}`;
    let institutionID;

    try {
      if (institution_name) {
        const response = await apiService.autocompleteInstitution(institution_name);
        institutionID = response.profiles[0].id.split('/').pop();
      }

      const response = await axios.get(`${API_BASE_URL}/search`, { params: { 'author_name': authorName, 'institution_id': institutionID } });
      
      console.log('response.data-> ', response.data)
      
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
    console.log("Getting paper titles...")
    const query = `au-id(${authorId})`;
    const maxCountGenerator = new MaxCount();
    let maxCount = maxCountGenerator.next();
    let startIndex = 0;
    const response = await axios.get(`https://api.elsevier.com/content/search/scopus`, {
      headers: { 'Content-Type': 'application/json' },
      params: { query, apiKey: "c20d01f998041241c36b6e9e6d9d6402", count: maxCount, start: startIndex }
    });
    
    const numWorks =  Number.parseInt(response.data['search-results']['opensearch:totalResults']);
    console.log(numWorks)
    const titleBatch = response.data['search-results']['entry'].map(entry => entry['dc:title']);
    console.log("Got first batch...")
    onLoadTitleBatch(titleBatch);
    let titles = titleBatch;
    if (numWorks > maxCount) {
      let promises = [];
      startIndex += maxCount;
      maxCount = maxCountGenerator.next();
      while (startIndex < numWorks) {
        const promise = axios.get(`https://api.elsevier.com/content/search/scopus`, {
          headers: { 'Content-Type': 'application/json' },
          params: { query, apiKey: "c20d01f998041241c36b6e9e6d9d6402", count: maxCount, start: startIndex }
        }).then(response => {
          const titleBatch = response.data['search-results']['entry'].map(entry => entry['dc:title']);
          onLoadTitleBatch(titleBatch);
          return titleBatch;
        })
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
  }
};

export default apiService;
export { API_BASE_URL };
