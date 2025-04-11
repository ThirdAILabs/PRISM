import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import useGoBack from '../../hooks/useGoBack.js';
import { IoIosClose } from 'react-icons/io';
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
import useOutsideClick from '../../hooks/useOutsideClick.js';
import '../../styles/pages/_universityReport.scss';
import { Divider } from '@mui/material';

import '../../styles/components/_primaryButton.scss';
import '../../styles/components/_authorInfoCard.scss';
import UniversityInfoCard from './university/UniversityInfoCard.js';
import ScoreCard from './university/UniversityScoreCard.js';
import Lottie from 'lottie-react';
import loadingAnimation from '../../assets/animations/Loader.json';

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
  const [loading, setLoading] = useState(true);
  const [valueFontSize, setValueFontSize] = useState(`${BaseFontSize}px`);
  const [isPanelVisible, setIsPanelVisible] = useState(false);
  const [universityInfo, setUniversityInfo] = useState(null);

  const handleReview = (flag) => {
    setSelectedFlag(flag);
    setSelectedFlagData(reportContent?.Flags[flag] || []);
    requestAnimationFrame(() => {
      setIsPanelVisible(true);
    });
  };

  const handleClosePanel = () => {
    setIsPanelVisible(false);
    // Wait for transition to complete before clearing selection
    setTimeout(() => {
      setSelectedFlag(null);
      setSelectedFlagData(null);
    }, 300);
  };

  const universityFlagPanelRef = useOutsideClick(handleClosePanel);

  useEffect(() => {
    let isMounted = true;

    const poll = async () => {
      const report = await universityReportService.getReport(report_id);

      if (report.Status === 'in-progress') {
        setInstituteName(report.UniversityName);
        setTotalResearchers(report.Content.TotalAuthors);
        setResearchersAssessed(report.Content.AuthorsReviewed);
        setUniversityInfo({
          name: report.UniversityName,
          address: report?.UniversityAddress,
        });
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
        setUniversityInfo({
          name: report.UniversityName,
          address: report?.UniversityAddress,
        });

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
    <div className="basic-setup" style={{ minHeight: '100vh', paddingBottom: '50px' }}>
      <div className={`panel-overlay ${isPanelVisible ? 'visible' : ''}`} />
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
            onClick={() => goBack()}
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
          <h5>University Assessment Result</h5>
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
      <>
        {universityInfo ? (
          <>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <div
                className="author-item"
                style={{
                  marginTop: '20px',
                  marginBottom: '20px',
                  marginLeft: '3%',
                  width: '45%',
                  height: '0%',
                }}
              >
                {universityInfo && <UniversityInfoCard result={universityInfo} />}
              </div>
              <div
                className="author-item"
                style={{ marginTop: '20px', marginBottom: '20px', marginRight: '3%', width: '45%' }}
              >
                <ScoreCard
                  reserachersAccessed={researchersAssessed}
                  totalResearcher={toatlResearchers}
                  loading={loading}
                />
              </div>
            </div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                flexWrap: 'wrap',
                marginTop: '40px',
                marginInline: '3%',
              }}
            >
              {reportContent?.Flags
                ? FLAG_ORDER.map((flag, index) => {
                    const flagData = reportContent.Flags[flag] || [];

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
                          title={TitlesAndDescriptions[flag].title}
                          hoverText={TitlesAndDescriptions[flag].desc}
                          value={flagData.length || 0}
                          speedometerHoverText={`${flagData.length} Authors`}
                          onReview={() => handleReview(flag)}
                          selected={flag === selectedFlag}
                          key={index}
                          valueFontSize={valueFontSize}
                        />
                      </div>
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

        {selectedFlag && (
          <div
            className={`university-flag-panel ${isPanelVisible ? 'open' : ''}`}
            ref={universityFlagPanelRef}
          >
            <div className="university-flag-panel-header">
              <span>{TitlesAndDescriptions[selectedFlag]?.title}</span>
              <button className="close-button" onClick={handleClosePanel}>
                <IoIosClose />
              </button>
            </div>
            <Divider className="university-flag-panel-divider" />
            <div className="university-flag-panel-content">
              <AuthorCard score={selectedFlagData.length} authors={selectedFlagData} />
            </div>
          </div>
        )}
      </>
    </div>
  );
};

export default UniversityReport;
