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

  recreateReport: async (reportId) => {
    const response = await axiosInstance.post(API_ROUTES.REPORTS.RECREATE(reportId));
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
    files.forEach((file) => {
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
  downloadReport: async (reportId, format) => {
    try {
      const response = await axiosInstance.get(
        `${API_ROUTES.REPORTS.DOWNLOAD(reportId)}?format=${format}`,
        {
          responseType: 'blob',
        }
      );

      // Create blob and download
      const blob = response.data;
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;

      // Get filename from Content-Disposition or use default
      const contentDisposition = response.headers['content-disposition'];
      const filename = contentDisposition
        ? contentDisposition.split('filename=')[1].replace(/"/g, '')
        : `report.${format}`;

      link.setAttribute('download', filename);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Download failed:', error);
      throw error;
    }
  },
};
