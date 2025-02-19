import React, { useState } from 'react';

const DisclosureUploadButton = ({ hookValues }) => {
  const { uploadPdf, uploadStatus, isUploading } = hookValues;
  const [selectedFiles, setSelectedFiles] = useState(null);
  
  const handleFileChange = (event) => {
    setSelectedFiles(event.target.files);
    event.target.value = '';
  };
  
  const handleUpload = () => {
    if (selectedFiles) uploadPdf(selectedFiles);
  };

  const getFileCountText = () => {
    if (!selectedFiles) return 'Upload Disclosures';
    const count = selectedFiles.length;
    return `${count} file${count !== 1 ? 's' : ''} selected`;
  };

  return (
    <div className="inline-flex items-center gap-3">
      <label 
        htmlFor="file-upload"
        className={`
          inline-flex items-center gap-2 px-4 py-2
          text-sm font-medium rounded-lg cursor-pointer
          shadow-sm
          transition-all duration-200 ease-in-out
          hover:shadow-md
          ${selectedFiles
            ? 'bg-gradient-to-r from-green-50 to-emerald-50 text-emerald-700 border-emerald-200'
            : 'bg-gradient-to-r from-gray-50 to-white text-gray-700 hover:from-blue-50 hover:to-indigo-50 hover:text-blue-600 hover:border-blue-200'}
        `}
      >
        {/* <Upload size={16} className="stroke-2" />
        <span className="font-semibold">
          {getFileCountText()}
        </span> */}
        <input
          id="file-upload"
          type="file"
          accept="application/pdf"
          onChange={handleFileChange}
          className="hidden"
          multiple
        />
      </label>

      {selectedFiles && (
        <button
          onClick={handleUpload}
          disabled={isUploading}
          className={`
            px-4 py-2 text-sm font-semibold rounded-lg
            shadow-sm border
            transition-all duration-200 ease-in-out
            hover:shadow-md
            ${isUploading
              ? 'bg-blue-50 text-blue-400 border-blue-200'
              : 'bg-gradient-to-r from-blue-500 to-indigo-500 hover:from-blue-600 hover:to-indigo-600 text-white border-transparent'}
          `}
        >
          {isUploading ? 'Uploading...' : 'Submit'}
        </button>
      )}

      {uploadStatus && (
        <span className="text-sm font-medium text-emerald-600">
          {uploadStatus}
        </span>
      )}
    </div>
  );
};

export default DisclosureUploadButton;