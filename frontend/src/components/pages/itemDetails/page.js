import React, { useEffect, useRef, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import {
  TALENT_CONTRACTS,
  ASSOCIATIONS_WITH_DENIED_ENTITIES,
  HIGH_RISK_FUNDERS,
  AUTHOR_AFFILIATIONS,
  POTENTIAL_AUTHOR_AFFILIATIONS,
  MISC_HIGH_RISK_AFFILIATIONS,
  COAUTHOR_AFFILIATIONS,
  FlagInformation,
} from '../../../constants/constants.js';
import ConcernVisualizer, { BaseFontSize, getFontSize } from '../../ConcernVisualization.js';
// import Graph from '../../common/graph/graph.js';
import Graph from '../../common/graph/graph2.js';
import { reportService } from '../../../api/reports.js';

import MuiAlert from '@mui/material/Alert';
import { Snackbar, Divider } from '@mui/material';
import useGoBack from '../../../hooks/useGoBack.js';

import '../../../styles/components/_primaryButton.scss';
import FlagPanel from './flagPanel.js';
import NodePanel from './nodePanel.js';

import { getTrailingWhiteSpace } from '../../../utils/helper.js';
import '../../../styles/components/_primaryButton.scss';
import '../../../styles/components/_authorInfoCard.scss';
import AuthorInfoCard from '../authorInstituteSearch/AuthorInfoCard.js';
import '../../../styles/pages/_authorReport.scss';
import ScoreCard from './ScoreCard.js';
import Lottie from 'lottie-react';
import loadingAnimation from '../../../assets/animations/Loader.json';

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
  const { report_id } = useParams();

  const [reportContent, setReportContent] = useState({});
  const [authorInfo, setAuthorInfo] = useState(null);
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
  const verifyWithDisclosure = {
    fileInputRef: fileInputRef,
    handleFileUploadClick: handleFileUploadClick,
    handleFileSelect: handleFileSelect,
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

        setAuthorInfo({
          AuthorId: report.AuthorId,
          AuthorName: report.AuthorName,
          Institutions: report.Affiliations.split(','),
          Interests: report.ResearchInterests.split(','),
          Source: report.Source,
        });
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
  const handleTabChange = (newValue) => {
    setActiveTab(newValue);
  };
  const handleStartDateChange = (e) => setStartDate(e.target.value);
  const handleEndDateChange = (e) => setEndDate(e.target.value);

  function parseLocalDate(dateStr) {
    const [year, month, day] = dateStr.split('-');
    return new Date(Number(year), Number(month) - 1, Number(day));
  }

  const handleDateFilter = () => {
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

    setReportContent(filteredContent);
    setReportMetadata({
      ...reportMetadata,
      TimeRange: `${displayStart} to ${displayEnd}`,
    });

    const maxLength = Math.max(...FLAG_ORDER.map((flag) => reportContent[flag]?.length || 0));
    const newFontSize = `${getFontSize(maxLength)}px`;

    setValueFontSize(newFontSize);
  };
  const handleClearFilter = () => {
    setStartDate('');
    setEndDate('');
    setReportContent(initialReportContent);
    setReportMetadata({
      ...reportMetadata,
      TimeRange: '',
    });
  };
  const filterProps = {
    startDate: startDate,
    endDate: endDate,
    todayStr: todayStr,
    handleStartDateChange: handleStartDateChange,
    handleEndDateChange: handleEndDateChange,
    handleDateFilter: handleDateFilter,
    handleClearFilter: handleClearFilter,
  };

  const [review, setReview] = useState();

  const goBack = useGoBack('/');

  const downloadProps = {
    reportId: report_id,
    metadata: reportMetadata,
    content: reportContent,
    disabled: loading,
  };

  const handleBackButtonClick = () => {
    if (activeTab === 1) {
      setActiveTab(0);
      setReview(null);
    } else {
      goBack();
    }
  };
  console.log('nodedataclickReview', review);
  const [graphNodeInfo, setGraphNodeInfo] = useState(null);

  return (
    <div className="basic-setup">
      <div className="grid grid-cols-2 gap-4">
        <div className="flex flex-row">
          <div className="detail-header">
            <div
              style={{
                flex: '1',
                display: 'flex',
                justifyContent: 'flex-start',
                marginBottom: '-15px',
              }}
            >
              <button
                onClick={handleBackButtonClick}
                className="btn text-dark mb-3"
                style={{ display: 'flex', marginTop: '-10px' }}
              >
                <svg
                  width="24"
                  height="32"
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
              </button>
              <h5>Individual Assessment Result</h5>
            </div>
          </div>
          <Divider
            sx={{
              backgroundColor: 'black',
              height: '1px',
              width: '100%',
              opacity: 0.1,
            }}
          />
        </div>
        {activeTab === 0 && (
          <div className="d-flex justify-content-end mt-2 gap-2 px-2">
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
          </div>
        )}
      </div>
      {activeTab === 0 && (
        <>
          {authorInfo ? (
            <>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <div
                  className="author-item"
                  style={{
                    marginTop: '20px',
                    marginBottom: '20px',
                    marginLeft: '3%',
                    width: '45%',
                  }}
                >
                  <AuthorInfoCard
                    result={authorInfo}
                    verifyWithDisclosure={verifyWithDisclosure}
                    downloadProps={downloadProps}
                    filterProps={filterProps}
                    loading={loading}
                  />
                </div>
                <div
                  className="author-item"
                  style={{
                    marginTop: '20px',
                    marginBottom: '20px',
                    marginRight: '3%',
                    width: '45%',
                  }}
                >
                  <ScoreCard
                    score={Object.keys(reportContent || {})
                      .map((name) => (reportContent[name] || []).length)
                      .reduce((prev, curr) => prev + curr, 0)}
                    setActiveTab={setActiveTab}
                    loading={loading}
                  />
                </div>
              </div>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  flexWrap: 'wrap',
                  marginTop: '20px',
                  marginInline: '3%',
                }}
              >
                {FLAG_ORDER.map((flag, index) => {
                  const flagCount = reportContent[flag] ? reportContent[flag].length : 0;
                  const isSelected = review === flag;
                  return (
                    <div
                      style={{
                        border: '1px solid rgb(230, 230, 230)',
                        borderRadius: '8px',
                        padding: '0px',
                        width: '13.5%',
                      }}
                    >
                      <ConcernVisualizer
                        title={FlagInformation[flag].title}
                        hoverText={FlagInformation[flag].desc}
                        value={flagCount}
                        speedometerHoverText={`${flagCount} Issues`}
                        onReview={() => setReview(flag)}
                        key={index}
                        selected={isSelected}
                        valueFontSize={valueFontSize}
                      />
                    </div>
                  );
                })}
              </div>
            </>
          ) : (
            <div
              style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
              }}
            >
              <Lottie
                animationData={loadingAnimation}
                loop={true}
                autoplay={true}
                style={{ width: '64%' }}
              />
            </div>
          )}

          {review && (
            <>
              <div className="overlay">
                <div className="flag-panel-container">
                  <FlagPanel
                    reportContent={reportContent}
                    review={review}
                    setReview={setReview}
                    authorName={authorName}
                    isDisclosureChecked={isDisclosureChecked}
                  />
                </div>
              </div>
            </>
          )}
        </>
      )}

      {activeTab === 1 && (
        <>
          <Graph
            authorName={authorName}
            reportContent={reportContent}
            review={review}
            setReview={setReview}
            setGraphNodeInfo={setGraphNodeInfo}
          />
          {review && (
            <NodePanel
              reportContent={reportContent}
              review={review}
              setReview={setReview}
              authorName={authorName}
              isDisclosureChecked={isDisclosureChecked}
              graphNodeInfo={graphNodeInfo}
            />
          )}
        </>
      )}
    </div>
  );
};

export default ItemDetails;
