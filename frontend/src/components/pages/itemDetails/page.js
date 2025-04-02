// src/ItemDetails.js
import React, { useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  TALENT_CONTRACTS,
  ASSOCIATIONS_WITH_DENIED_ENTITIES,
  HIGH_RISK_FUNDERS,
  AUTHOR_AFFILIATIONS,
  POTENTIAL_AUTHOR_AFFILIATIONS,
  MISC_HIGH_RISK_AFFILIATIONS,
  COAUTHOR_AFFILIATIONS,
  TitlesAndDescriptions,
} from '../../../constants/constants.js';
import ConcernVisualizer, { BaseFontSize, getFontSize } from '../../ConcernVisualization.js';
import Graph from '../../common/graph/graph.js';
import Tabs from '../../common/tools/Tabs.js';
import DownloadButton from '../../common/tools/button/downloadButton.js';
import { reportService } from '../../../api/reports.js';
import MuiAlert from '@mui/material/Alert';
import { Snackbar, Tooltip } from '@mui/material';
import useGoBack from '../../../hooks/useGoBack.js';
import useOutsideClick from '../../../hooks/useOutsideClick.js';
import { getTrailingWhiteSpace } from '../../../utils/helper.js';
import '../../../styles/components/_primaryButton.scss';
import FlagPanel from './flagPanel.js';

const FLAG_ORDER = [
  TALENT_CONTRACTS,
  ASSOCIATIONS_WITH_DENIED_ENTITIES,
  HIGH_RISK_FUNDERS,
  AUTHOR_AFFILIATIONS,
  POTENTIAL_AUTHOR_AFFILIATIONS,
  MISC_HIGH_RISK_AFFILIATIONS,
  COAUTHOR_AFFILIATIONS,
];

const Alert = React.forwardRef(function Alert(props, ref) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

const todayStr = new Date().toISOString().split('T')[0];

const ItemDetails = () => {
  const navigate = useNavigate();
  const { report_id } = useParams();

  const [yearDropdownOpen, setYearDropdownOpen] = useState(false);
  const [downloadDropdownOpen, setDownloadDropdownOpen] = useState(false);
  const [reportContent, setReportContent] = useState({});
  const [authorName, setAuthorName] = useState('');
  const [initialReportContent, setInitialReportContent] = useState({});
  const [isDisclosureChecked, setDisclosureChecked] = useState(false);
  const [loading, setLoading] = useState(true);
  const [valueFontSize, setValueFontSize] = useState(`${BaseFontSize}px`);
  const [reportMetadata, setReportMetadata] = useState({});

  const [notification, setNotification] = useState({
    open: false,
    severity: '',
    message: '',
  });

  const fileInputRef = useRef(null);

  const handleFileUploadClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const handleFileSelect = async (event) => {
    const files = Array.from(event.target.files);
    if (files.length === 0) {
      setNotification({
        open: true,
        severity: 'error',
        message: 'No files selected',
      });
      return;
    }
    await handleSubmit(files);
  };

  const handleSubmit = async (files) => {
    if (!files || files.length === 0) {
      setNotification({
        open: true,
        severity: 'error',
        message: 'No files selected',
      });
      return;
    }

    try {
      const result = await reportService.checkDisclosure(report_id, files);
      const { Content, ...metadata } = result;
      setReportContent(Content);
      setInitialReportContent(Content);
      setDisclosureChecked(true);

      setReportMetadata({
        ...metadata,
        ContainsDisclosure: true,
        ContainsReportContent: true,
      });

      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
      setNotification({
        open: true,
        severity: 'success',
        message: 'Disclosure check succeeded!',
      });

      const maxLength = Math.max(...FLAG_ORDER.map((flag) => result.Content[flag]?.length || 0));
      const newFontSize = `${getFontSize(maxLength)}px`;

      setValueFontSize(newFontSize);
    } catch (error) {
      setNotification({
        open: true,
        severity: 'error',
        message: error.response?.data?.message || 'Failed to check disclosure',
      });
    }
  };

  const handleCloseNotification = (event, reason) => {
    if (reason === 'clickaway') {
      return;
    }
    setNotification({ ...notification, open: false });
  };

  useEffect(() => {
    let isMounted = true;
    let timeoutId = null;

    const poll = async () => {
      try {
        const report = await reportService.getReport(report_id);
        const { Content, ...metadata } = report;

        if (!isMounted) return;

        setAuthorName(report.AuthorName);
        setReportContent(report.Content);
        setInitialReportContent(report.Content);
        setReportMetadata({
          ...metadata,
          ContainsReportContent: true,
        });

        const maxLength = Math.max(...FLAG_ORDER.map((flag) => report.Content[flag]?.length || 0));
        const newFontSize = `${getFontSize(maxLength)}px`;
        setValueFontSize(newFontSize);

        const inProgress = report.Status === 'queued' || report.Status === 'in-progress';

        if (inProgress) {
          timeoutId = setTimeout(poll, 2000);
        } else {
          setLoading(false);
        }
      } catch (error) {
        console.error('Polling error:', error);
        setLoading(false);
      }
    };

    poll();

    return () => {
      isMounted = false;
      if (timeoutId) clearTimeout(timeoutId);
    };
  }, []);

  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [activeTab, setActiveTab] = useState(0);

  const [filterMessage, setFilterMessage] = useState(
    getTrailingWhiteSpace(12) + 'Filter by Timeline'
  );
  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };
  const handleStartDateChange = (e) => setStartDate(e.target.value);
  const handleEndDateChange = (e) => setEndDate(e.target.value);

  function parseLocalDate(dateStr) {
    const [year, month, day] = dateStr.split('-');
    return new Date(Number(year), Number(month) - 1, Number(day));
  }

  const handleDateFilter = () => {
    setYearDropdownOpen(false);
    if (!startDate && !endDate) {
      setReportContent(initialReportContent);
      setFilterMessage('');
      return;
    }

    let start = startDate ? parseLocalDate(startDate) : null;
    let end = endDate ? parseLocalDate(endDate) : null;

    if (start && !end) {
      end = new Date();
    }

    if (!start && end) {
      start = new Date('1900-01-01');
    }

    if (start > end) {
      alert('Start date cannot be after End date.');
      return;
    }

    const filteredContent = {};
    FLAG_ORDER.forEach((flag) => {
      if (initialReportContent[flag]) {
        filteredContent[flag] = initialReportContent[flag].filter((item) => {
          if (!item?.Work?.PublicationDate) return true;
          const pubDate = new Date(item.Work.PublicationDate);

          if (pubDate < start) return false;
          if (pubDate > end) return false;

          return true;
        });
      } else {
        filteredContent[flag] = [];
      }
    });

    const displayStart = startDate
      ? parseLocalDate(startDate).toLocaleDateString('en-US', {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
        })
      : 'earliest';
    const displayEnd = endDate
      ? parseLocalDate(endDate).toLocaleDateString('en-US', {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
        })
      : 'today';

    setFilterMessage(`${displayStart} - ${displayEnd}`);

    setStartDate('');
    setEndDate('');

    setReportContent(filteredContent);
    setReportMetadata({
      ...reportMetadata,
      TimeRange: `${displayStart} to ${displayEnd}`,
    });

    const maxLength = Math.max(...FLAG_ORDER.map((flag) => reportContent[flag]?.length || 0));
    const newFontSize = `${getFontSize(maxLength)}px`;

    setValueFontSize(newFontSize);
  };

  const [review, setReview] = useState();

  const [showDisclosed, setShowDisclosed] = useState(false);
  const [showUndisclosed, setShowUndisclosed] = useState(false);
  const disclosedItems = (reportContent[review] || []).filter((item) => item.Disclosed);
  const undisclosedItems = (reportContent[review] || []).filter((item) => !item.Disclosed);

  const items = review ? reportContent[review] || [] : [];
  const hasDates = items.some(
    (item) => item?.Work?.PublicationDate && !isNaN(new Date(item.Work.PublicationDate).getTime())
  );
  const goBack = useGoBack('/');

  const dropdownFilterRef = useOutsideClick(() => {
    setYearDropdownOpen(false);
  });

  const dropdownDownloadRef = useOutsideClick(() => {
    setDownloadDropdownOpen(false);
  });

  return (
    <div className="basic-setup">
      <div className="grid grid-cols-2 gap-4">
        <div className="flex flex-row">
          <div
            className="detail-header"
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              margin: '0 auto',
              padding: '10px 0',
              height: '75px',
            }}
          >
            {/* Left section - Back button */}
            <div style={{ flex: '1', display: 'flex', justifyContent: 'flex-start' }}>
              <button
                onClick={() => goBack()}
                className="btn text-dark mb-3"
                style={{ minWidth: '80px', display: 'flex', alignItems: 'center' }}
              >
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                  style={{ marginRight: '8px' }}
                >
                  <path
                    d="M10 19L3 12L10 5"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                  <path
                    d="M3 12H21"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
                Back
              </button>
            </div>

            {/* Center section - Author information */}
            <div style={{ flex: '1', textAlign: 'center' }}>
              <h5 className="m-0">{authorName}</h5>
            </div>

            {/* Right section - Filter dropdown */}
            <div style={{ flex: '1', display: 'flex', justifyContent: 'flex-end' }}>
              <div className="dropdown" ref={dropdownFilterRef}>
                <style>
                  {`
                    .form-control::placeholder {
                      color: #888;
                    }
                  `}
                </style>
                <Tooltip title={loading ? 'Please wait while the report is being generated.' : ''}>
                  <span
                    style={{
                      cursor: loading ? 'not-allowed' : 'pointer',
                    }}
                  >
                    <button
                      className="btn dropdown-toggle"
                      onClick={() => setYearDropdownOpen(!yearDropdownOpen)}
                      style={{
                        backgroundColor: 'rgb(160, 160, 160)',
                        border: 'none',
                        marginRight: '10px',
                        color: 'white',
                        width: '225px',
                        fontWeight: 'bold',
                        fontSize: '14px',
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        cursor: loading ? 'not-allowed' : 'pointer',
                      }}
                      disabled={loading}
                    >
                      {filterMessage}
                    </button>
                  </span>
                </Tooltip>
                {yearDropdownOpen && (
                  <div
                    className="dropdown-menu show p-2"
                    style={{
                      width: '225px',
                      backgroundColor: 'rgb(160, 160, 160)',
                      border: 'none',
                      // right: 0,
                      marginTop: '5px',
                      marginRight: '10px',
                      color: 'white',
                      fontWeight: 'bold',
                      fontSize: '14px',
                      justifyContent: 'center',
                      alignItems: 'center',
                      display: 'flex',
                      flexDirection: 'column',
                    }}
                  >
                    <div
                      className="form-group"
                      style={{ marginBottom: '10px', width: '100%', padding: '7px' }}
                    >
                      <label>Start Date</label>
                      <input
                        type="date"
                        className="form-control"
                        value={startDate}
                        max={todayStr}
                        onChange={handleStartDateChange}
                        style={{
                          backgroundColor: 'rgb(220, 220, 220)',
                          border: 'none',
                          outline: 'none',
                          color: 'black',
                          marginTop: '5px',
                          width: '100%',
                        }}
                      />
                    </div>
                    <div
                      className="form-group"
                      style={{ marginBottom: '10px', width: '100%', padding: '0 7px' }}
                    >
                      <label>End Date</label>
                      <input
                        type="date"
                        value={endDate}
                        max={todayStr}
                        onChange={handleEndDateChange}
                        className="form-control"
                        style={{
                          backgroundColor: 'rgb(220, 220, 220)',
                          border: 'none',
                          outline: 'none',
                          color: 'black',
                          marginTop: '5px',
                          width: '100%',
                        }}
                      />
                    </div>
                    <button
                      className="form-control"
                      type="submit"
                      onClick={handleDateFilter}
                      disabled={!(startDate || endDate)}
                      style={{
                        backgroundColor: 'rgb(200, 200, 200)',
                        border: 'none',
                        color: 'white',
                        width: '100px',
                        fontWeight: 'bold',
                        fontSize: '14px',
                        cursor: startDate || endDate ? 'pointer' : 'not-allowed',
                        transition: 'background-color 0.3s',
                        marginTop: '10px',
                      }}
                    >
                      Submit
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>
          <Tabs activeTab={activeTab} handleTabChange={handleTabChange} disabled={loading} />
        </div>
        {activeTab === 0 && (
          <div className="d-flex justify-content-end mt-2 gap-2 px-2">
            <Tooltip title={loading ? 'Please wait while the report is being generated.' : ''}>
              <button
                className="button"
                onClick={handleFileUploadClick}
                disabled={loading}
                style={{
                  cursor: loading ? 'not-allowed' : 'pointer',
                }}
              >
                Verify with Disclosures
              </button>
            </Tooltip>
            <input
              type="file"
              ref={fileInputRef}
              style={{ display: 'none' }}
              multiple
              accept=".txt,.pdf"
              onChange={handleFileSelect}
            />
            <Snackbar
              open={notification.open}
              autoHideDuration={2000}
              onClose={handleCloseNotification}
              anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
            >
              <Alert onClose={handleCloseNotification} severity={notification.severity}>
                {notification.message}
              </Alert>
            </Snackbar>
            <div ref={dropdownDownloadRef}>
              <DownloadButton
                reportId={report_id}
                metadata={reportMetadata}
                content={reportContent}
                isOpen={downloadDropdownOpen}
                setIsOpen={() => setDownloadDropdownOpen(!downloadDropdownOpen)}
                disabled={loading}
              />
            </div>
          </div>
        )}
      </div>
      {activeTab === 0 && (
        <>
          {loading && (
            <div class="d-flex justify-content-start">
              <div class="spinner-border text-secondary ms-5 mb-3" role="status" />
            </div>
          )}
          <div
            className="d-flex w-100 flex-column align-items-center"
            style={{ color: 'rgb(78, 78, 78)', marginTop: '0px' }}
          >
            <div style={{ fontSize: 'large', fontWeight: 'bold' }}>Total Score</div>
            <div style={{ fontSize: '60px', fontWeight: 'bold' }}>
              {Object.keys(reportContent || {})
                .map((name) => (reportContent[name] || []).length)
                .reduce((prev, curr) => prev + curr, 0)}
            </div>
          </div>

          <div
            style={{
              display: 'flex',
              justifyContent: 'space-around',
              flexWrap: 'wrap',
              marginTop: '20px',
            }}
          >
            {FLAG_ORDER.map((flag, index) => {
              const flagCount = reportContent[flag] ? reportContent[flag].length : 0;
              const isSelected = review === flag;
              return (
                <ConcernVisualizer
                  title={TitlesAndDescriptions[flag].title}
                  hoverText={TitlesAndDescriptions[flag].desc}
                  value={flagCount}
                  speedometerHoverText={`${flagCount} Issues`}
                  onReview={() => setReview(flag)}
                  key={index}
                  selected={isSelected}
                  valueFontSize={valueFontSize}
                />
              );
            })}
          </div>
          {review && <FlagPanel reportContent={reportContent} review={review} setReview={setReview} authorName={authorName} isDisclosureChecked={isDisclosureChecked} disclosedItems={disclosedItems} showDisclosed={showDisclosed} setShowDisclosed={setShowDisclosed} undisclosedItems={undisclosedItems} showUndisclosed = {showUndisclosed} setShowUndisclosed={setShowUndisclosed} />}
        </>
      )}

      {activeTab === 1 && <Graph authorName={authorName} reportContent={reportContent} />}
    </div>
  );
};

export default ItemDetails;
