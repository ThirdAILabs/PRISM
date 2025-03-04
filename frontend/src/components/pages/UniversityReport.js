import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  TALENT_CONTRACTS,
  ASSOCIATIONS_WITH_DENIED_ENTITIES,
  HIGH_RISK_FUNDERS,
  AUTHOR_AFFILIATIONS,
  POTENTIAL_AUTHOR_AFFILIATIONS,
  MISC_HIGH_RISK_AFFILIATIONS,
  COAUTHOR_AFFILIATIONS,
} from '../../constants/constants.js';
import ConcernVisualizer from '../ConcernVisualization.js';

import { Dialog, DialogTitle, DialogContent, DialogActions, Button, Divider } from '@mui/material';
import { universityReportService } from '../../api/universityReports.js';
import AuthorCard from '../common/cards/AuthorCard.js';

import styled from 'styled-components';
import Loader from './university/Loader.js';

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
    desc: 'Authors in these papers are recruited by talent programs that have close ties to high-risk foreign governments.',
  },
  [ASSOCIATIONS_WITH_DENIED_ENTITIES]: {
    title: 'Funding from Denied Entities',
    desc: 'Some of the parties involved in these works are in the denied entity lists of U.S. government agencies.',
  },
  [HIGH_RISK_FUNDERS]: {
    title: 'High Risk Funding Sources',
    desc: 'These papers are funded by funding sources that have close ties to high-risk foreign governments.',
  },
  [AUTHOR_AFFILIATIONS]: {
    title: 'Affiliations with High Risk Foreign Institutes',
    desc: 'Papers that list the queried author as being affiliated with a high-risk foreign institution or web pages that showcase official appointments at high-risk foreign institutions.',
  },
  [POTENTIAL_AUTHOR_AFFILIATIONS]: {
    title: 'Appointments at High Risk Foreign Institutes*',
    desc: 'The author may have an appointment at a high-risk foreign institutions.\n\n*Collated information from the web, might contain false positives.',
  },
  [MISC_HIGH_RISK_AFFILIATIONS]: {
    title: 'Miscellaneous High Risk Connections*',
    desc: 'The author or an associate may be mentioned in a press release.\n\n*Collated information from the web, might contain false positives.',
  },
  [COAUTHOR_AFFILIATIONS]: {
    title: "Co-authors' affiliations with High Risk Foreign Institutes",
    desc: 'Coauthors in these papers are affiliated with high-risk foreign institutions.',
  },
};

//Mock University report.
const mockUniversityReport = {
  Id: 'e42ba4dd-f56b-4916-835b-034679df2d4b',
  CreatedAt: '2024-03-11T20:21:49.387547Z',
  UniversityId: 'stanford-university-123',
  UniversityName: 'Stanford University',
  Status: 'complete',
  FacultyCount: 200,
  Content: {
    TotalAuthors: 4,
    AuthorsReviewed: 3,
    Flags: {
      AssociationsWithDeniedEntities: [
        {
          AuthorId: 'auth_123',
          AuthorName: 'John Smith',
          Source: 'IEEE',
          FlagCount: 3,
        },
        {
          AuthorId: 'auth_123',
          AuthorName: 'John Smith',
          Source: 'IEEE',
          FlagCount: 3,
        },
        {
          AuthorId: 'auth_123',
          AuthorName: 'John Smith',
          Source: 'IEEE',
          FlagCount: 3,
        },
        {
          AuthorId: 'auth_123',
          AuthorName: 'John Smith',
          Source: 'IEEE',
          FlagCount: 3,
        },
        {
          AuthorId: 'auth_123',
          AuthorName: 'John Smith',
          Source: 'IEEE',
          FlagCount: 3,
        },
        {
          AuthorId: 'auth_123',
          AuthorName: 'John Smith',
          Source: 'IEEE',
          FlagCount: 3,
        },
        {
          AuthorId: 'auth_123',
          AuthorName: 'John Smith',
          Source: 'IEEE',
          FlagCount: 3,
        },
        {
          AuthorId: 'auth_123',
          AuthorName: 'John Smith',
          Source: 'IEEE',
          FlagCount: 3,
        },
        {
          AuthorId: 'auth_123',
          AuthorName: 'John Smith',
          Source: 'IEEE',
          FlagCount: 3,
        },
        {
          AuthorId: 'auth_123',
          AuthorName: 'John Smith',
          Source: 'IEEE',
          FlagCount: 3,
        },
      ],
      AuthorAffiliations: [
        {
          AuthorId: 'auth_124',
          AuthorName: 'Maria Garcia',
          Source: 'Scopus',
          FlagCount: 2,
        },
      ],
      CoauthorAffiliations: [
        {
          AuthorId: 'auth_125',
          AuthorName: 'David Chen',
          Source: 'Web of Science',
          FlagCount: 4,
        },
      ],
      HighRiskFunders: [
        {
          AuthorId: 'auth_126',
          AuthorName: 'Sarah Johnson',
          Source: 'NIH Database',
          FlagCount: 1,
        },
      ],
      MiscHighRiskAssociations: [
        {
          AuthorId: 'auth_127',
          AuthorName: 'Michael Brown',
          Source: 'ArXiv',
          FlagCount: 2,
        },
      ],
      PotentialAuthorAffiliations: [
        {
          AuthorId: 'auth_128',
          AuthorName: 'Lisa Anderson',
          Source: 'Google Scholar',
          FlagCount: 3,
        },
      ],
      TalentContracts: [
        {
          AuthorId: 'auth_129',
          AuthorName: 'Robert Wilson',
          Source: 'Research Gate',
          FlagCount: 2,
        },
      ],
    },
  },
};

const UniversityReport = () => {
  const navigate = useNavigate();
  const { report_id } = useParams();

  const [reportContent, setReportContent] = useState({});
  const [instituteName, setInstituteName] = useState('');
  const [toatlResearchers, setTotalResearchers] = useState(0);
  const [researchersAssessed, setResearchersAssessed] = useState(0);
  const [selectedFlagData, setSelectedFlagData] = useState(null);
  const [showModal, setShowModal] = useState(false);

  const handleReview = (flag) => {
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
      }
      if (report.Status === 'complete' && isMounted) {
        console.log('Report', report);
        setInstituteName(report.UniversityName);
        setReportContent(report.Content);
        setLoading(false);
      } else if (isMounted) {
        setTimeout(poll, 10000);
      }
    };

    poll();

    return () => {
      isMounted = false;
    };
    // const report = mockUniversityReport;
    // console.log('Report', report);
    // setInstituteName(report.UniversityName);
    // setReportContent(report.Content);
    // setLoading(false);
  }, []);

  const [loading, setLoading] = useState(true);
  // if (loading) {
  //     return null;
  // }

  return (
    <div className="basic-setup" style={{ minHeight: '100vh', paddingBottom: '50px' }}>
      <div className="grid grid-cols-2 gap-4">
        <div className="flex flex-row">
          <div className="detail-header">
            <button
              onClick={() => navigate(-1)}
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
            <h5 style={{ marginRight: '43.5%' }}>{instituteName}</h5>
          </div>
        </div>
      </div>

      <>
        <div
          className="d-flex w-100 flex-column align-items-center"
          style={{ color: 'rgb(78, 78, 78)', marginTop: '50px' }}
        >
          <div style={{ fontSize: 'large', fontWeight: 'bold' }}>Total Researchers</div>
          <div style={{ fontSize: '60px', fontWeight: 'bold' }}>{toatlResearchers}</div>
          <div style={{ fontSize: 'medium', fontWeight: 'bold' }}>Researchers Assessed</div>
          <div style={{ fontSize: '50px', fontWeight: 'bold' }}>{researchersAssessed}</div>
        </div>

        {loading ? (
          <div style={{ marginTop: '10%', marginLeft: '62%' }}>
            <Loader />
          </div>
        ) : (
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-around',
              flexWrap: 'wrap',
              marginTop: '20px',
            }}
          >
            {reportContent?.Flags &&
              FLAG_ORDER.map((flag, index) => {
                let value = 0;
                const flagData = reportContent.Flags[flag] || [];

                if (flagData.length > 0) {
                  for (let authorIndex = 0; authorIndex < flagData.length; authorIndex++) {
                    const author = flagData[authorIndex];
                    value += author.FlagCount;
                  }
                }

                return (
                  <ConcernVisualizer
                    title={TitlesAndDescriptions[flag].title}
                    hoverText={TitlesAndDescriptions[flag].desc}
                    value={value || 0}
                    onReview={() => handleReview(flag)}
                    key={index}
                  />
                );
              })}
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
