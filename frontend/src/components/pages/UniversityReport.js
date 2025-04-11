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
import Divider from '@mui/material/Divider';

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
  const [showModal, setShowModal] = useState(false);
  const [loading, setLoading] = useState(true);
  const [valueFontSize, setValueFontSize] = useState(`${BaseFontSize}px`);
  const [universityInfo, setUniversityInfo] = useState(null);

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
        setUniversityInfo({
          name: report.UniversityName,
          // address: report.UniversityAddress,
          address: '6100 Main St, Houston, TX 77005, USA',
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
          // address: report.UniversityAddress,
          address: '6100 Main St, Houston, TX 77005, USA',
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
        {universityInfo ?
          (<>
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
                        width: '13.5%'
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
          </>) : (
            <div style={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              height: 'calc(100vh - 100px)'
            }}>
              <Lottie
                animationData={loadingAnimation}
                loop={true}
                autoplay={true}
                style={{ width: 2000 }}
              />
            </div>
          )}


        {showModal && (
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              width: '100%',
            }}
          >
            <div
              style={{
                marginTop: '100px',
              }}
            >
              <AuthorCard authors={selectedFlagData} />
            </div>
          </div>
        )}
      </>
    </div>
  );
};

const popoverStyles = {
  position: 'absolute',
  top: '30px',
  left: '50%',
  transform: 'translateX(-50%)',
  zIndex: 1,
  backgroundColor: '#fff',
  border: '1px solid rgba(0, 0, 0, 0.2)',
  boxShadow: '0 0.5rem 1rem rgba(0, 0, 0, 0.15)',
  borderRadius: '0.3rem',
  padding: '0.5rem',
  width: '200px',
};

const buttonStyles = {
  marginLeft: '5px',
  width: '14px',
  height: '14px',
  padding: '1px 0',
  borderRadius: '7.5px',
  textAlign: 'center',
  fontSize: '8px',
  lineHeight: '1.42857',
  border: '1px solid grey',
  borderWidth: '1px',
  backgroundColor: 'transparent',
  color: 'grey',
  position: 'relative',
  boxShadow: 'none',
};

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
    transform: scale(1.009);
    border-color: #fff9;
  }
`;

export default UniversityReport;
