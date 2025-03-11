import axiosInstance from './axios.config';
import { API_ROUTES } from './constants';

export const autocompleteService = {
  autocompleteAuthors: async (query, institution_id) => {
    const response = await axiosInstance.get(API_ROUTES.AUTOCOMPLETE.AUTHOR, {
      params: { query, institution_id },
    });
    return response.data;
  },

  autocompleteInstitutions: async (query) => {
    const response = await axiosInstance.get(API_ROUTES.AUTOCOMPLETE.INSTITUTION, {
      params: { query },
    });
    return response.data;
  },

  autocompletePaperTitles: async (query) => {
    const response = await axiosInstance.get(API_ROUTES.AUTOCOMPLETE.PAPER_TITLE, {
      params: { query },
    });
    return response.data;
  },
};
