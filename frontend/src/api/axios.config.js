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
    if (error.response?.status === 401) {
      UserService.doLogout();
      return Promise.reject(error);
    }
    const errorMessage =
      error.response?.data?.message || error.message || 'An unexpected error occurred';

    // Use a fallback value if status is null or undefined
    const errorStatus = error.response?.status || 500;

    // window.location.href = `/error?message=${encodeURIComponent(
    //   errorMessage
    // )}&status=${encodeURIComponent(errorStatus)}`;

    return Promise.reject(error);
  }
);

export default axiosInstance;
