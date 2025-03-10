import axiosInstance from './axios.config';
import { API_ROUTES } from './constants';

export const searchService = {
  searchOpenalexAuthors: async (authorName, institutionId, institutionName) => {
    const response = await axiosInstance.get(API_ROUTES.SEARCH.AUTHOR, {
      params: {
        author_name: authorName,
        institution_id: institutionId,
        institution_name: institutionName,
      },
    });
    return response.data;
  },

  searchGoogleScholarAuthors: async (authorName, institutionName, cursor) => {
    const response = await axiosInstance.get(API_ROUTES.SEARCH.AUTHOR_ADVANCED, {
      params: { author_name: authorName, institution_name: institutionName, cursor: cursor },
    });
    return response.data;
  },

  matchEntities: async (query) => {
    const response = await axiosInstance.get(API_ROUTES.SEARCH.MATCH_ENTITIES, {
      params: { query },
    });
    return response.data;
  },

  searchByOrcid: async (orcidId) => {
    const response = await axiosInstance.get(API_ROUTES.SEARCH.AUTHOR, {
      params: { orcid: orcidId },
    });
    return response.data;
  },

  searchByPaperTitle: async (title) => {
    const response = await axiosInstance.get(API_ROUTES.SEARCH.AUTHOR, {
      params: { paper_title: title },
    });
    return response.data;
  },
};
