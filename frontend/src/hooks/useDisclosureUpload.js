import { useState, useCallback } from "react";
import { useLocation, useParams } from 'react-router-dom';
import axios from 'axios';
import UserService from '../services/userService';
import { API_BASE_URL } from '../services/apiService';

const useDisclosureUpload = () => {
    const [uploadStatus, setUploadStatus] = useState("");
    const [isUploading, setIsUploading] = useState(false);
    const [uploadResult, setUploadResult] = useState(null);
    const { report_id } = useParams();
    const location = useLocation();

    const uploadPdf = useCallback(async (pdfFiles) => {
        if (!pdfFiles || pdfFiles.length === 0) {
            setUploadStatus("No files selected.");
            return null;
        }

        setIsUploading(true);
        // Create a new state update to trigger re-render
        setUploadResult(null);
        setUploadStatus("");

        const formData = new FormData();
        const filePromises = Array.from(pdfFiles).map(file =>
            new Promise((resolve, reject) => {
                const reader = new FileReader();
                reader.onload = () => {
                    const blob = new Blob([reader.result], { type: 'text/plain' });
                    formData.append('files', blob, file.name);
                    resolve();
                };
                reader.onerror = reject;
                reader.readAsArrayBuffer(file);
            })
        );

        try {
            await Promise.all(filePromises);
            const token = UserService.getToken();

            const response = await axios.post(
                `${API_BASE_URL}/api/v1/report/${report_id}/check-disclosure`,
                formData,
                {
                    headers: {
                        Authorization: `Bearer ${token}`,
                        'Content-Type': 'multipart/form-data'
                    }
                }
            );

            if (response.status === 200) {
                console.log("Upload successful, about to set uploadResult to:", response.data);

                // Ensure we're creating a new object reference
                const newResult = { ...response.data };

                // Set states in sequence to ensure updates
                setUploadStatus(`Successfully uploaded ${pdfFiles.length} file(s)!`);
                setUploadResult(newResult);

                console.log("States have been updated");
                return newResult;
            }
        } catch (error) {
            console.error("File upload failed:", error);
            setUploadStatus(
                error.response?.data?.message || "File upload failed. Please try again."
            );
            setUploadResult(null);
        } finally {
            setIsUploading(false);
        }

        return null;
    }, [location.state]);

    return {
        uploadPdf,
        uploadStatus,
        isUploading,
        uploadResult
    };
};

export default useDisclosureUpload;