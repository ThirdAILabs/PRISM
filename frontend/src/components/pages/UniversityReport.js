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
import styled from 'styled-components';

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

const ItemDetails = () => {
  const navigate = useNavigate();
  const { report_id } = useParams();

  const [reportContent, setReportContent] = useState({});
  const [instituteName, setInstituteName] = useState('');
  // const [institutions, setInstitutions] = useState([]);

  useEffect(() => {
    let isMounted = true;
    const poll = async () => {
      const report = await universityReportService.getReport(report_id);
      if (report.Status === 'complete' && isMounted) {
        console.log('Report', report);
        setInstituteName(report.AuthorName);
        setReportContent(report.Content);
        setLoading(false);
      } else if (isMounted) {
        // setTimeout(poll, 2000);
      }
    };

    poll();

    return () => {
      isMounted = false;
    };
  }, []);

  const [loading, setLoading] = useState(true);

  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [yearDropdownOpen, setYearDropdownOpen] = useState(false);
  const handleStartDateChange = (e) => setStartDate(e.target.value);
  const handleEndDateChange = (e) => setEndDate(e.target.value);

  const toggleYearDropdown = () => {
    setYearDropdownOpen(!yearDropdownOpen);
  };

  const [review, setReview] = useState();

  return (
    <div className="basic-setup">
      <div className="grid grid-cols-2 gap-4">
        <div className="flex flex-row">
          <div className="detail-header">
            <button
              onClick={() => navigate('/')}
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

            <div className="d-flex w-80">
              <div className="text-start px-5">
                <div className="d-flex align-items-center mb-2">
                  <h5 className="m-0">{'Sample Institute'}</h5>
                </div>
                {/* <b className="m-0 p-0" style={{ fontSize: 'small' }}>
                                    {institutions.join(', ')}
                                </b> */}
              </div>
            </div>

            {/* <div>
                            <div className="dropdown">
                                <style>
                                    {` 
                    .form-control::placeholder { 
                        color: #888; 
                    }`}
                                </style>
                                <button
                                    className="btn dropdown-toggle"
                                    type="button"
                                    onClick={toggleYearDropdown}
                                    style={{
                                        backgroundColor: 'rgb(160, 160, 160)',
                                        border: 'none',
                                        color: 'white',
                                        width: '200px',
                                        fontWeight: 'bold',
                                        fontSize: '14px',
                                    }}
                                >
                                    Filter by Timeline
                                </button>
                                {yearDropdownOpen && (
                                    <div
                                        className="dropdown-menu show p-3"
                                        style={{
                                            backgroundColor: 'rgb(160, 160, 160)',
                                            border: 'none',
                                            right: 0,
                                            marginTop: '5px',
                                            color: 'white',
                                            fontWeight: 'bold',
                                            fontSize: '14px',
                                            justifyContent: 'center',
                                            alignItems: 'center',
                                            display: 'flex',
                                            flexDirection: 'column',
                                        }}
                                    >
                                        <div className="form-group mb-2">
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
                                                    marginTop: '10px',
                                                    width: '100%',
                                                }}
                                            />
                                        </div>
                                        <div style={{ height: '10px' }} />
                                        <div className="form-group">
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
                                                    marginTop: '10px',
                                                }}
                                            />
                                        </div>
                                        <div
                                            style={{
                                                marginTop: '20px',
                                                display: 'flex',
                                                flexDirection: 'column',
                                                gap: '10px',
                                            }}
                                        >
                                            <button
                                                className="form-control"
                                                onClick={handleResetFilter}
                                                style={{
                                                    backgroundColor: 'rgb(220, 220, 220)',
                                                    border: 'none',
                                                    color: 'black',
                                                    width: '100px',
                                                    fontWeight: 'bold',
                                                    fontSize: '14px',
                                                    cursor: 'pointer',
                                                    transition: 'background-color 0.3s',
                                                }}
                                            >
                                                Reset
                                            </button>

                                            <button
                                                className="form-control"
                                                type="submit"
                                                onClick={handleDateFilter}
                                                disabled={!(startDate || endDate)}
                                                style={{
                                                    backgroundColor: 'black',
                                                    border: 'none',
                                                    color: 'white',
                                                    width: '100px',
                                                    fontWeight: 'bold',
                                                    fontSize: '14px',
                                                    cursor: startDate || endDate ? 'pointer' : 'default',
                                                    transition: 'background-color 0.3s',
                                                }}
                                            >
                                                Submit
                                            </button>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div> */}
          </div>
        </div>
      </div>

      <>
        <div className="d-flex w-100 flex-column align-items-center">
          <div className="d-flex w-100 px-5 align-items-center my-2 mt-3 justify-content-between">
            <div style={{ width: '20px' }}>
              {loading && (
                <div className="spinner-border text-primary spinner-border-sm" role="status"></div>
              )}
            </div>
          </div>
        </div>

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
            return (
              <ConcernVisualizer
                title={TitlesAndDescriptions[flag].title}
                hoverText={TitlesAndDescriptions[flag].desc}
                value={reportContent[flag] ? reportContent[flag].length : 0}
                onReview={() => setReview(flag)}
                key={index}
              />
            );
          })}
        </div>
        {review && <div>HUE HUE HEU</div>}
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

export default ItemDetails;
