import axiosInstance from './axios.config';
import { API_ROUTES } from './constants';

export const licenseService = {
    listLicenses: async () => {
        const response = await axiosInstance.get(API_ROUTES.LICENSES.LIST);
        return response.data;
    },

    createLicense: async (licenseData) => {
        const response = await axiosInstance.post(API_ROUTES.LICENSES.CREATE, licenseData);
        return response.data;
    },

    deactivateLicense: async (licenseId) => {
        await axiosInstance.delete(API_ROUTES.LICENSES.DEACTIVATE(licenseId));
    }
};