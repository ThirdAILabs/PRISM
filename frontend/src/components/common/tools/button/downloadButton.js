import React, { useState } from 'react';
import styled from 'styled-components';
import { FaFilePdf, FaFileCsv, FaFileExcel } from 'react-icons/fa';
import { reportService } from '../../../../api/reports';
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
    // left: 0;
    right: 0;
    margin-top: 8px;
    z-index: 800;
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

const Button = ({ reportId, isOpen, setIsOpen }) => {
    const handleDownload = (format) => {
        reportService.downloadReport(reportId, format);
        setIsOpen(false);
    };
    return (
        <StyledWrapper>
            <button className="cssbuttons-io-button"
                onClick={() => setIsOpen(!isOpen)}
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width={20} height={20}>
                    <path fill="none" d="M0 0h24v24H0z" />
                    <path fill="currentColor" d="M1 14.5a6.496 6.496 0 0 1 3.064-5.519 8.001 8.001 0 0 1 15.872 0 6.5 6.5 0 0 1-2.936 12L7 21c-3.356-.274-6-3.078-6-6.5zm15.848 4.487a4.5 4.5 0 0 0 2.03-8.309l-.807-.503-.12-.942a6.001 6.001 0 0 0-11.903 0l-.12.942-.805.503a4.5 4.5 0 0 0 2.029 8.309l.173.013h9.35l.173-.013zM13 12h3l-4 5-4-5h3V8h2v4z" />
                </svg>
                {/* <span>Download Report</span> */}
            </button>
            {isOpen && (
                <div className="dropdown-container">
                    <div
                        className="dropdown-item"
                        onClick={() => handleDownload('pdf')}
                    >
                        <FaFilePdf color="#ff0000" />
                        PDF
                    </div>
                    <div
                        className="dropdown-item"
                        onClick={() => handleDownload('csv')}
                    >
                        <FaFileCsv color="#217346" />
                        CSV
                    </div>
                    <div
                        className="dropdown-item"
                        onClick={() => handleDownload('xlsx')}
                    >
                        <FaFileExcel color="#217346" />
                        Excel
                    </div>
                </div>
            )}
        </StyledWrapper>
    );
}



export default Button;
