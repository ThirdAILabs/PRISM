import React, { useState } from 'react';
import styled from 'styled-components';
import { FaFilePdf, FaFileCsv, FaFileExcel } from 'react-icons/fa';
import { reportService } from '../../../../api/reports';
import { Tooltip } from '@mui/material';
import '../../../../styles/components/_primaryButton.scss';
const StyledWrapper = styled.div`
  position: relative;
  .cssbuttons-io-button {
    position: relative;
    transition: all 0.3s ease-in-out;
    box-shadow: 0px 10px 20px rgba(0, 0, 0, 0.2);
    padding-block: 0.5rem;
    padding-inline: 0.75rem;
    background-color: rgb(0 107 179);
    border-radius: 9999px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    color: #ffff;
    gap: 10px;
    font-weight: bold;
    border: 3px solid #ffffff4d;
    outline: none;
    overflow: hidden;
    font-size: 15px;
  }

  .cssbuttons-io-button:hover {
    transform: scale(1.05);
    border-color: #fff9;
  }

  .cssbuttons-io-button:hover .icon {
    transform: translate(4px);
  }

  .dropdown-container {
    position: absolute;
    top: 100%;
    right: 0;
    margin-top: 8px;
    z-index: 1050;
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 8px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    min-width: 160px;
    animation: fadeIn 0.2s ease-in-out;
  }
  .dropdown-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 20px;
    color: #333;
    cursor: pointer;
    transition: background-color 0.2s;

    &:hover {
      background-color: #f5f5f5;
    }

    svg {
      font-size: 20px;
    }
  }
`;

const DownloadDropdown = ({ reportId, metadata, content, isOpen, setIsOpen }) => {
  const handleDownload = (format) => {
    reportService.downloadReport(reportId, format, metadata, content);
    setIsOpen(false);
  };

  return (
    <StyledWrapper>
      {isOpen && (
        <div className="dropdown-container" style={{ marginTop: '20px' }}>
          <div className="dropdown-item" onClick={() => handleDownload('pdf')}>
            <FaFilePdf color="#ff0000" />
            PDF
          </div>
          <div className="dropdown-item" onClick={() => handleDownload('csv')}>
            <FaFileCsv color="#217346" />
            CSV
          </div>
          <div className="dropdown-item" onClick={() => handleDownload('xlsx')}>
            <FaFileExcel color="#217346" />
            Excel
          </div>
        </div>
      )}
    </StyledWrapper>
  );
};

export default DownloadDropdown;
