import axios from 'axios';
import { prismBaseUrl } from './constants';
import UserService from '../services/userService';
const axiosInstance = axios.create({
    baseURL: prismBaseUrl,
    headers: {
        'Content-Type': 'application/json'
    }
});

axiosInstance.interceptors.request.use(
    (config) => {
        const token = UserService.getToken();
        console.log('token in axios.config.js', token);
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

export default axiosInstance;