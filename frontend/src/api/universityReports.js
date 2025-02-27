import axiosInstance from './axios.config';
import { API_ROUTES } from './constants';

export const universityReportService = {
  listReports: async () => {
    const response = await axiosInstance.get(API_ROUTES.UNIVERSITY_REPORTS.LIST);
    return response.data;
  },

  createReport: async (reportData) => {
    const response = await axiosInstance.post(API_ROUTES.UNIVERSITY_REPORTS.CREATE, {
      UniversityId: reportData.UniversityId,
      UniversityName: reportData.UniversityName,
    });
    return response.data;
  },

  getReport: async (reportId) => {
    const response = await axiosInstance.get(API_ROUTES.UNIVERSITY_REPORTS.GET(reportId));
    return response.data;
  },
};

export default universityReportService;
