import axiosInstance from './axios.config';
import { API_ROUTES } from './constants';

export const autocompleteService = {
  autocompleteAuthors: async (query) => {
    const response = await axiosInstance.get(API_ROUTES.AUTOCOMPLETE.AUTHOR, {
      params: { query },
    });
    return response.data;
  },

  autocompleteInstitutions: async (query) => {
    const response = await axiosInstance.get(API_ROUTES.AUTOCOMPLETE.INSTITUTION, {
      params: { query },
    });
    return response.data;
  },
};
