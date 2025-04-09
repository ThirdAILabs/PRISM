import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import useGoBack from '../../hooks/useGoBack.js';
import {
  TALENT_CONTRACTS,
  ASSOCIATIONS_WITH_DENIED_ENTITIES,
  HIGH_RISK_FUNDERS,
  AUTHOR_AFFILIATIONS,
  POTENTIAL_AUTHOR_AFFILIATIONS,
  MISC_HIGH_RISK_AFFILIATIONS,
  COAUTHOR_AFFILIATIONS,
} from '../../constants/constants.js';
import ConcernVisualizer, { BaseFontSize, getFontSize } from '../ConcernVisualization.js';

import { universityReportService } from '../../api/universityReports.js';
import AuthorCard from '../common/cards/AuthorCard.js';

import styled from 'styled-components';
import '../../styles/pages/_universityReport.scss';

const FLAG_ORDER = [
  TALENT_CONTRACTS,
  ASSOCIATIONS_WITH_DENIED_ENTITIES,
  HIGH_RISK_FUNDERS,
  AUTHOR_AFFILIATIONS,
  POTENTIAL_AUTHOR_AFFILIATIONS,
  MISC_HIGH_RISK_AFFILIATIONS,
  COAUTHOR_AFFILIATIONS,
];

const todayStr = new Date().toISOString().split('T')[0];

const TitlesAndDescriptions = {
  [TALENT_CONTRACTS]: {
    title: 'Talent Contracts',
    desc: 'Researchers in this list appear in papers funded by Talent Contracts.',
  },
  [ASSOCIATIONS_WITH_DENIED_ENTITIES]: {
    title: 'Funding from Denied Entities',
    desc: 'Researchers in this list appear in papers funded by Denied Entities.',
  },
  [HIGH_RISK_FUNDERS]: {
    title: 'High Risk Funding Sources',
    desc: 'Researchers in this list appear in papers funded by High Risk Funding Sources.',
  },
  [AUTHOR_AFFILIATIONS]: {
    title: 'Affiliations with High Risk Foreign Institutes',
    desc: 'Researchers in this list have affiliations with High Risk Foreign Institutes.',
  },
  [POTENTIAL_AUTHOR_AFFILIATIONS]: {
    title: 'Appointments at High Risk Foreign Institutes*',
    desc: 'Researchers in this list have appointments with High Risk Foreign Institutes.\n\n*Collated information from the web, might contain false positives.',
  },
  [MISC_HIGH_RISK_AFFILIATIONS]: {
    title: 'Miscellaneous High Risk Connections*',
    desc: 'Researchers in this list or their associates appear in press releases.\n\n*Collated information from the web, might contain false positives.',
  },
  [COAUTHOR_AFFILIATIONS]: {
    title: "Co-authors' affiliations with High Risk Foreign Institutes",
    desc: 'Researchers in this list have co-authors who are affiliated with High Risk Foreign Institutes.',
  },
};

const UniversityReport = () => {
  const navigate = useNavigate();
  const { report_id } = useParams();

  const [reportContent, setReportContent] = useState({});
  const [instituteName, setInstituteName] = useState('');
  const [toatlResearchers, setTotalResearchers] = useState(0);
  const [researchersAssessed, setResearchersAssessed] = useState(0);
  const [selectedFlag, setSelectedFlag] = useState(null);
  const [selectedFlagData, setSelectedFlagData] = useState(null);
  const [showModal, setShowModal] = useState(false);
  const [loading, setLoading] = useState(true);
  const [valueFontSize, setValueFontSize] = useState(`${BaseFontSize}px`);

  const handleReview = (flag) => {
    setSelectedFlag(flag);
    setSelectedFlagData(reportContent?.Flags[flag] || []);
    setShowModal(true);
  };

  useEffect(() => {
    let isMounted = true;

    const poll = async () => {
      const report = await universityReportService.getReport(report_id);

      if (report.Status === 'in-progress') {
        setInstituteName(report.UniversityName);
        setTotalResearchers(report.Content.TotalAuthors);
        setResearchersAssessed(report.Content.AuthorsReviewed);

        if (!isMounted) {
          return;
        }

        // Set font size based on the maximum number of flag count
        const newFontSize = `${getFontSize(
          Math.max(...FLAG_ORDER.map((flag) => report.Content.Flags[flag]?.length || 0))
        )}px`;
        setValueFontSize(newFontSize);
      } else if (report.Status === 'complete') {
        setInstituteName(report.UniversityName);
        setReportContent(report.Content);
        setTotalResearchers(report.Content.TotalAuthors);
        setResearchersAssessed(report.Content.AuthorsReviewed);

        // Set font size based on the maximum number of flag count
        const newFontSize = `${getFontSize(
          Math.max(...FLAG_ORDER.map((flag) => report.Content.Flags[flag]?.length || 0))
        )}px`;
        setValueFontSize(newFontSize);

        if (report.Content.TotalAuthors === report.Content.AuthorsReviewed) {
          setLoading(false);
        } else {
          setTimeout(poll, 10000);
        }
      } else {
        setTimeout(poll, 10000);
      }
    };

    poll();

    return () => {
      isMounted = false;
    };
  }, []);

  const goBack = useGoBack('/university');

  return (
    <div className="basic-setup" style={{ minHeight: '100vh', paddingBottom: '50px', position: 'relative', overflow: 'hidden' }}>
      <div
        className="detail-header"
        style={{
          width: '100%',
          position: 'relative',
          display: 'flex',
          alignItems: 'center',
          height: '75px',
        }}
      >
        <button
          onClick={() => goBack()}
          className="btn text-dark mb-3"
          style={{
            minWidth: '80px',
            position: 'absolute',
            left: '10px',
            top: '20px',
          }}
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
        <h5
          style={{
            margin: '0 auto',
            width: '100%',
            textAlign: 'center',
          }}
        >
          {instituteName}
        </h5>
      </div>

      <>
        {loading && (
          <div className="d-flex justify-content-start">
            <div className="spinner-border text-secondary ms-5 mt-3 mb-3" role="status" />
          </div>
        )}
        {
          <div
            className="d-flex w-100 flex-column align-items-center"
            style={{ color: 'rgb(78, 78, 78)', marginTop: '20px' }}
          >
            <div style={{ fontSize: 'large', fontWeight: 'bold' }}>Total Researchers</div>
            <div style={{ fontSize: '60px', fontWeight: 'bold' }}>{toatlResearchers}</div>
            <div style={{ fontSize: 'medium', fontWeight: 'bold' }}>Researchers Assessed</div>
            <div style={{ fontSize: '50px', fontWeight: 'bold' }}>{researchersAssessed}</div>
          </div>
        }

        <div
          style={{
            display: 'flex',
            justifyContent: 'space-around',
            flexWrap: 'wrap',
            marginTop: '20px',
          }}
        >
          {reportContent?.Flags
            ? FLAG_ORDER.map((flag, index) => {
                const flagData = reportContent.Flags[flag] || [];

                return (
                  <ConcernVisualizer
                    title={TitlesAndDescriptions[flag].title}
                    hoverText={TitlesAndDescriptions[flag].desc}
                    value={flagData.length || 0}
                    speedometerHoverText={`${flagData.length} Authors`}
                    onReview={() => handleReview(flag)}
                    selected={flag === selectedFlag}
                    key={index}
                    valueFontSize={valueFontSize}
                  />
                );
              })
            : FLAG_ORDER.map((flag, index) => {
                return (
                  <ConcernVisualizer
                    title={TitlesAndDescriptions[flag].title}
                    hoverText={TitlesAndDescriptions[flag].desc}
                    value={0}
                    speedometerHoverText={`0 Authors`}
                    onReview={() => handleReview(flag)}
                    selected={flag === selectedFlag}
                    key={index}
                    valueFontSize={valueFontSize}
                  />
                );
              })}
        </div>

        {showModal && (
          <div className={`university-flag-panel ${showModal ? 'open' : ''}`}>
            <div className="university-flag-panel-header">
              <h4>{TitlesAndDescriptions[selectedFlag]?.title}</h4>
              <button className="close-button" onClick={() => setShowModal(false)}>×</button>
            </div>
            <div className="selection-container">
              <div className="score-box-container">
                <div className="score-pill">
                  <span className="score-label">Score:</span>
                  <span className="score-value">
                    {selectedFlagData?.length || 0}
                  </span>
                </div>
              </div>
              <input type="text" placeholder="Search..." className="search-input" />
            </div>
            
            <div className="university-flag-panel-content">
              <AuthorCard authors={selectedFlagData} />
            </div>
          </div>
        )}
      </>
    </div>
  );
};

export default UniversityReport;
