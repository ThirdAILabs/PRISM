import axiosInstance from './axios.config';
import { API_ROUTES } from './constants';

export const reportService = {
    listReports: async () => {
        const response = await axiosInstance.get(API_ROUTES.REPORTS.LIST);
        return response.data;
    },

    createReport: async (reportData) => {
        const response = await axiosInstance.post(API_ROUTES.REPORTS.CREATE, reportData);
        return response.data;
    },

    getReport: async (reportId) => {
        const response = await axiosInstance.get(API_ROUTES.REPORTS.GET(reportId));
        return response.data;
    },

    activateLicense: async (licenseKey) => {
        await axiosInstance.post(API_ROUTES.REPORTS.ACTIVATE_LICENSE, { License: licenseKey });
    },

    checkDisclosure: async (reportId, files) => {
        const formData = new FormData();
        files.forEach(file => {
            formData.append('files', file);
        });

        const response = await axiosInstance.post(
            API_ROUTES.REPORTS.CHECK_DISCLOSURE(reportId),
            formData,
            {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
            }
        );
        return response.data;
    },
};