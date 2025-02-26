import axiosInstance from './axios.config';
import { API_ROUTES } from './constants';

export const searchService = {
  searchOpenalexAuthors: async (authorName, institutionId, institutionName) => {
    const response = await axiosInstance.get(API_ROUTES.SEARCH.REGULAR, {
      params: {
        author_name: authorName,
        institution_id: institutionId,
        institution_name: institutionName,
      },
    });
    return response.data;
  },

  searchGoogleScholarAuthors: async (query, cursor) => {
    const response = await axiosInstance.get(API_ROUTES.SEARCH.ADVANCED, {
      params: { query, cursor },
    });
    return response.data;
  },

  matchEntities: async (query) => {
    const response = await axiosInstance.get(API_ROUTES.SEARCH.MATCH_ENTITIES, {
      params: { query },
    });
    return response.data;
  },
};
