import axios from 'axios';
import { prismBaseUrl } from './constants';
import UserService from '../services/userService';

const axiosInstance = axios.create({
  baseURL: prismBaseUrl,
  headers: {
    'Content-Type': 'application/json',
  },
});

axiosInstance.interceptors.request.use(
  (config) => {
    const token = UserService.getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => {
    const errorMessage =
      error.response?.data?.message || error.message || 'An unexpected error occurred';
    // window.location.href = `/error?message=${encodeURIComponent(errorMessage)}`;
    return Promise.reject(error);
  }
);

export default axiosInstance;
