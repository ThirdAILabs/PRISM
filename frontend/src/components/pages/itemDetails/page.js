// src/ItemDetails.js
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
} from '../../../constants/constants.js';
import ConcernVisualizer from '../../ConcernVisualization.js';
import Graph from '../../common/graph/graph.js';
import Tabs from '../../common/tools/Tabs.js';
import DownloadButton from '../../common/tools/button/downloadButton.js';
import { Dialog, DialogTitle, DialogContent, DialogActions, Button, Divider } from '@mui/material';
import { reportService } from '../../../api/reports.js';
import styled from 'styled-components';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import ArrowRightIcon from '@mui/icons-material/ArrowRight';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import Shimmer from './Shimmer.js';

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
    desc: 'These papers are funded by talent programs run by foreign governents that intend to recruit high-performing researchers.',
  },
  [ASSOCIATIONS_WITH_DENIED_ENTITIES]: {
    title: 'Funding from Denied Entities',
    desc: 'Some of the parties involved in these papers are in denied entity lists of government agencies.',
  },
  [HIGH_RISK_FUNDERS]: {
    title: 'High Risk Funding Sources',
    desc: 'These papers are funded by sources that have close ties to high-risk foreign governments.',
  },
  [AUTHOR_AFFILIATIONS]: {
    title: 'Affiliations with High Risk Foreign Institutes',
    desc: 'These papers list the queried author as being affiliated with a high-risk foreign institute.',
  },
  [POTENTIAL_AUTHOR_AFFILIATIONS]: {
    title: 'Appointments at High Risk Foreign Institutes*',
    desc: 'The author may have an appointment at a high-risk foreign institute.\n\n*Collated information from the web, might contain false positives.',
  },
  [MISC_HIGH_RISK_AFFILIATIONS]: {
    title: 'Miscellaneous High Risk Connections*',
    desc: 'The author or an associate may be mentioned in a press release.\n\n*Collated information from the web, might contain false positives.',
  },
  [COAUTHOR_AFFILIATIONS]: {
    title: "Co-authors' affiliations with High Risk Foreign Institutes",
    desc: 'Co-authors in these papers are affiliated with high-risk foreign institutes.',
  },
};

const get_paper_url = (flag) => {
  return (
    <>
      <a href={flag.Work.WorkUrl} target="_blank" rel="noopener noreferrer">
        {flag.Work.DisplayName}
      </a>
      {flag.Work.OaUrl && (
        <text>
          {' '}
          [
          <a href={flag.Work.OaUrl} target="_blank" rel="noopener noreferrer">
            full text
          </a>
          ]
        </text>
      )}
    </>
  );
};

const ItemDetails = () => {
  const navigate = useNavigate();
  const { report_id } = useParams();
  const [dropdownOpen, setDropdownOpen] = useState(0);
  const [reportContent, setReportContent] = useState({});
  const [authorName, setAuthorName] = useState('');
  const [institutions, setInstitutions] = useState([]);
  const [initialReprtContent, setInitialReportContent] = useState({});
  const [isDisclosureChecked, setDisclosureChecked] = useState(false);

  // Add new states
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedFiles, setSelectedFiles] = useState([]);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadError, setUploadError] = useState(null);

  // Add handlers
  const handleDropdownChange = (index) => {
    if (index === dropdownOpen) setDropdownOpen(0);
    else setDropdownOpen(index);
  };

  const handleOpenDialog = () => setOpenDialog(true);
  const handleCloseDialog = () => {
    handleDropdownChange(0);
    setOpenDialog(false);
    setSelectedFiles([]);
    setUploadError(null);
  };

  const handleDrop = (event) => {
    event.preventDefault();
    const files = Array.from(event.dataTransfer.files);
    setSelectedFiles(files);
  };

  const handleFileSelect = (event) => {
    const files = Array.from(event.target.files);
    setSelectedFiles(files);
  };

  const handleSubmit = async () => {
    if (selectedFiles.length === 0) {
      setUploadError('Please select at least one file');
      return;
    }

    setIsUploading(true);
    try {
      const result = await reportService.checkDisclosure(report_id, selectedFiles);
      setReportContent(result.Content);
      setInitialReportContent(result.Content);
      setDisclosureChecked(true);
      handleCloseDialog();
    } catch (error) {
      setUploadError(error.message || 'Failed to check disclosure');
    } finally {
      setIsUploading(false);
    }
  };

  useEffect(() => {
    let isMounted = true;
    const poll = async () => {
      const report = await reportService.getReport(report_id);
      if (report.Status === 'complete' && isMounted) {
        console.log('Report', report);
        setAuthorName(report.AuthorName);
        setReportContent(report.Content);
        setInitialReportContent(report.Content);
        setLoading(false);
      } else if (isMounted) {
        setTimeout(poll, 2000);
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
  const [isDownloadOpen, setIsDownloadOpen] = useState(false);
  const [activeTab, setActiveTab] = useState(0);

  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };
  const handleStartDateChange = (e) => setStartDate(e.target.value);
  const handleEndDateChange = (e) => setEndDate(e.target.value);
  const toggleYearDropdown = () => {
    setYearDropdownOpen(!yearDropdownOpen);
    setIsDownloadOpen(false); // Close download dropdown
  };

  const handleDateFilter = () => {
    if (!startDate && !endDate) {
      setReportContent(initialReprtContent);
      setYearDropdownOpen(false);
      return;
    }

    let start = startDate ? new Date(startDate) : null;
    let end = endDate ? new Date(endDate) : null;

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
      if (initialReprtContent[flag]) {
        filteredContent[flag] = initialReprtContent[flag].filter((item) => {
          if (!item?.Work?.PublicationDate) return true;
          const pubDate = new Date(item.Work.PublicationDate);

          if (pubDate < start) return false;
          if (pubDate > end) return false;

          return true;
        });
      } else {
        filteredContent[flag] = null;
      }
    });

    setReportContent(filteredContent);
    setYearDropdownOpen(false);
  };
  const [instDropdownOpen, setInstDropdownOpen] = useState(false);
  const toggleInstDropdown = () => setInstDropdownOpen(!instDropdownOpen);

  const [review, setReview] = useState();

  function withPublicationDate(header, flag) {
    const publicationDateStr = flag.Work && flag.Work.PublicationDate;
    let formattedDate = 'N/A';
    if (publicationDateStr) {
      const publicationDate = new Date(publicationDateStr);
      formattedDate = publicationDate.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      });
    }
    return (
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        {header}
        <span className="fw-bold mt-3">{formattedDate}</span>
      </div>
    );
  }

  function multipleAffiliationsFlag(flag, index) {
    return (
      <div key={index} className="p-3 px-5 w-75 detail-item">
        {withPublicationDate(
          <h5 className="fw-bold mt-3">Author has multiple affiliations</h5>,
          flag
        )}
        <p>
          {authorName} is affiliated with multiple institutions in {get_paper_url(flag)}. Detected
          affiliations:
          <ul className="bulleted-list">
            {flag.Affiliations.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <li key={key}>
                  <a>{item}</a>
                </li>
              );
            })}
          </ul>
        </p>
        {isDisclosureChecked &&
          (flag.Disclosed ? (
            <button type="button" className="btn btn-success">
              Disclosed
            </button>
          ) : (
            <button type="button" className="btn btn-danger">
              Undisclosed
            </button>
          ))}
      </div>
    );
  }

  function funderFlag(flag, index) {
    return (
      <div key={index} className="p-3 px-5 w-75 detail-item">
        {withPublicationDate(
          <h5 className="fw-bold mt-3">Funder is an Entity of Concern</h5>,
          flag
        )}
        <p>
          {get_paper_url(flag)} is funded by the following entities of concern:
          <ul className="bulleted-list">
            {flag.Funders.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <li key={key}>
                  <a>{item}</a>
                </li>
              );
            })}
          </ul>
          {flag.RawAcknowledements.length > 0 && (
            <>
              <strong>Acknowledgements Text</strong>
              <ul className="bulleted-list">
                {flag.RawAcknowledements.map((item, index2) => {
                  const key = `ack-${index} ${index2}`;
                  return <li key={key}>{item}</li>;
                })}
              </ul>
            </>
          )}
        </p>
        {isDisclosureChecked &&
          (flag.Disclosed ? (
            <button type="button" className="btn btn-success">
              Disclosed
            </button>
          ) : (
            <button type="button" className="btn btn-danger">
              Undisclosed
            </button>
          ))}
      </div>
    );
  }

  function publisherFlag(flag, index) {
    return (
      <div key={index} className="p-3 px-5 w-75 detail-item">
        {withPublicationDate(
          <h5 className="fw-bold mt-3">Publisher is an Entity of Concern</h5>,
          flag
        )}
        <p>
          {get_paper_url(flag)} is published by the following entities of concern:
          <ul className="bulleted-list">
            {flag.Publishers.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <li key={key}>
                  <a>{item}</a>
                </li>
              );
            })}
          </ul>
        </p>
        {isDisclosureChecked &&
          (flag.Disclosed ? (
            <button type="button" className="btn btn-success">
              Disclosed
            </button>
          ) : (
            <button type="button" className="btn btn-danger">
              Undisclosed
            </button>
          ))}
      </div>
    );
  }

  function coauthorFlag(flag, index) {
    return (
      <div key={index} className="p-3 px-5 w-75 detail-item">
        {withPublicationDate(
          <h5 className="fw-bold mt-3">Co-authors are high-risk entities</h5>,
          flag
        )}
        <p>
          The following co-authors of {get_paper_url(flag)} are high-risk entities:
          <ul className="bulleted-list">
            {flag.Coauthors.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <li key={key}>
                  <a>{item}</a>
                </li>
              );
            })}
          </ul>
        </p>
        {isDisclosureChecked &&
          (flag.Disclosed ? (
            <button type="button" className="btn btn-success">
              Disclosed
            </button>
          ) : (
            <button type="button" className="btn btn-danger">
              Undisclosed
            </button>
          ))}
      </div>
    );
  }

  function coauthorAffiliationFlag(flag, index) {
    return (
      <div key={index} className="p-3 px-5 w-75 detail-item">
        {withPublicationDate(
          <h5 className="fw-bold mt-3">Co-authors are affiliated with Entities of Concern</h5>,
          flag
        )}
        <p>
          Some authors of {get_paper_url(flag)} are affiliated with entities of concern:
          <ul className="bulleted-list">
            {flag.Affiliations.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <li key={key}>
                  <a>{item}</a>
                </li>
              );
            })}
          </ul>
          <strong>Affiliated Authors</strong>
          <ul className="bulleted-list">
            {flag.Coauthors.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <li key={key}>
                  <a>{item}</a>
                </li>
              );
            })}
          </ul>
        </p>
        {isDisclosureChecked &&
          (flag.Disclosed ? (
            <button type="button" className="btn btn-success">
              Disclosed
            </button>
          ) : (
            <button type="button" className="btn btn-danger">
              Undisclosed
            </button>
          ))}
      </div>
    );
  }

  function authorAffiliationFlag(flag, index) {
    return (
      <div key={index} className="p-3 px-5 w-75 detail-item">
        {withPublicationDate(
          <h5 className="fw-bold mt-3">Author is affiliated with an Entity of Concern</h5>,
          flag
        )}
        <div>
          {authorName} is affiliated with an entity of concern in {get_paper_url(flag)}.<p></p>
          <strong>Detected Affiliations</strong>
          <ul className="bulleted-list">
            {flag.Affiliations.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <li key={key}>
                  <a>{item}</a>
                </li>
              );
            })}
          </ul>
        </div>
        {isDisclosureChecked &&
          (flag.Disclosed ? (
            <button type="button" className="btn btn-success">
              Disclosed
            </button>
          ) : (
            <button type="button" className="btn btn-danger">
              Undisclosed
            </button>
          ))}
      </div>
    );
  }

  function acknowledgementFlag(flag, index) {
    return (
      <div key={index} className="p-3 px-5 w-75 detail-item">
        {withPublicationDate(
          <h5 className="fw-bold mt-3">Acknowledgements possibly contain Talent Contracts</h5>,
          flag
        )}
        <p>
          {flag.Entities > 0 ? (
            <>{get_paper_url(flag)} acknowledges the following entities of concern:</>
          ) : (
            <>Some acknowledged entities in {get_paper_url(flag)} may be foreign entities.</>
          )}
          <ul className="bulleted-list">
            {flag.Entities.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <li key={key}>
                  <a>
                    "{item.Entity}"{' was detected in '}
                    <text style={{ fontWeight: 'bold', textDecoration: 'underline' }}>
                      {item.Sources.join(', ')}
                    </text>
                    {' as '}
                    {item.Aliases.map((element) => `"${element}"`).join(', ')}
                  </a>
                </li>
              );
            })}
          </ul>
          <strong>Acknowledgement Text</strong>
          {flag.RawAcknowledements.map((item, index3) => {
            return <p key={index3}>{item}</p>;
          })}
          <p>{ }</p>
        </p>
        { }
        {isDisclosureChecked &&
          (flag.Disclosed ? (
            <button type="button" className="btn btn-success">
              Disclosed
            </button>
          ) : (
            <button type="button" className="btn btn-danger">
              Undisclosed
            </button>
          ))}
      </div>
    );
  }

  function universityFacultyFlag(flag, index) {
    return (
      <div key={index} className="p-3 px-5 w-75 detail-item">
        <h5 className="fw-bold mt-3">
          The author may potentially be linked with an Entity of Concern
        </h5>
        <p>
          {flag.Message}
          Relevant Webpage:{' '}
          <a href={flag.UniversityUrl} target="_blank" rel="noopener noreferrer">
            {flag.UniversityUrl}
          </a>
        </p>
        {isDisclosureChecked &&
          (flag.Disclosed ? (
            <button type="button" className="btn btn-success">
              Disclosed
            </button>
          ) : (
            <button type="button" className="btn btn-danger">
              Undisclosed
            </button>
          ))}
      </div>
    );
  }

  const [showDisclosed, setShowDisclosed] = useState(true);
  const [showUndisclosed, setShowUndisclosed] = useState(true);
  const disclosedItems = (reportContent[review] || []).filter((item) => item.Disclosed);
  const undisclosedItems = (reportContent[review] || []).filter((item) => !item.Disclosed);
  const [sortOrder, setSortOrder] = useState('desc');

  function renderFlags(items) {
    if (!items) {
      setReview(false);
      return null;
    }

    const sortedItems = [...items].sort((a, b) => {
      const dateA =
        a.Work && a.Work.PublicationDate ? new Date(a.Work.PublicationDate).getTime() : 0;
      const dateB =
        b.Work && b.Work.PublicationDate ? new Date(b.Work.PublicationDate).getTime() : 0;
      return sortOrder === 'asc' ? dateA - dateB : dateB - dateA;
    });

    return sortedItems.map((flag, index) => {
      switch (review) {
        case TALENT_CONTRACTS:
          return acknowledgementFlag(flag, index);
        case ASSOCIATIONS_WITH_DENIED_ENTITIES:
          return acknowledgementFlag(flag, index);
        case HIGH_RISK_FUNDERS:
          return funderFlag(flag, index);
        case AUTHOR_AFFILIATIONS:
          return authorAffiliationFlag(flag, index);
        case POTENTIAL_AUTHOR_AFFILIATIONS:
          return universityFacultyFlag(flag, index);
        case MISC_HIGH_RISK_AFFILIATIONS:
          return PRFlag(flag, index);
        case COAUTHOR_AFFILIATIONS:
          return coauthorAffiliationFlag(flag, index);
        default:
          return null;
      }
    });
  }

  function PRFlag(flag, index) {
    const connections = flag.Connections || [];
    return (
      <div key={index} className="p-3 px-5 w-75 detail-item">
        {true && (
          <>
            {connections.length == 0 ? (
              <>
                <h5 className="fw-bold mt-3">
                  The author or an associate may be mentioned in a Press Release
                </h5>
              </>
            ) : connections.length == 1 ? (
              <>
                <h5 className="fw-bold mt-3">
                  The author's associate may be mentioned in a Press Release
                </h5>
              </>
            ) : connections.length === 2 ? (
              <>
                <h5 className="fw-bold mt-3">
                  The author may potentially be connected to an entity/individual mentioned in a
                  Press Release
                </h5>
              </>
            ) : null}
          </>
        )}
        <p>
          {flag.Message}
          <p />
          <>
            {connections.length === 0 ? (
              <>
                <strong>Press Release</strong>
                <ul className="bulleted-list">
                  <li>
                    <a href={flag.DocUrl} target="_blank" rel="noopener noreferrer">
                      {flag.DocTitle}
                    </a>
                  </li>
                </ul>
              </>
            ) : connections.length === 1 ? (
              <>
                {flag.FrequentCoauthor ? (
                  <>Frequent Coauthor: {flag.FrequentCoauthor}</>
                ) : (
                  <>
                    <strong>Relevant Document</strong>
                    <ul className="bulleted-list">
                      <li>
                        <a
                          href={flag.Connections[0].DocUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          {flag.Connections[0].DocTitle}
                        </a>{' '}
                      </li>
                    </ul>
                  </>
                )}
                <strong>Press Release</strong>
                <ul className="bulleted-list">
                  <li>
                    <a href={flag.DocUrl} target="_blank" rel="noopener noreferrer">
                      {flag.DocTitle}
                    </a>
                  </li>
                </ul>
              </>
            ) : connections.length == 2 ? (
              <>
                <strong>Relevant Documents</strong>
                <ul className="bulleted-list">
                  <li>
                    <a href={flag.Connections[0].DocUrl} target="_blank" rel="noopener noreferrer">
                      {flag.Connections[0].DocTitle}
                    </a>
                  </li>
                  <li>
                    <a href={flag.Connections[1].DocUrl} target="_blank" rel="noopener noreferrer">
                      {flag.Connections[1].DocTitle}
                    </a>
                  </li>
                </ul>
                <strong>Press Release</strong>
                <ul className="bulleted-list">
                  <li>
                    <a href={flag.DocUrl} target="_blank" rel="noopener noreferrer">
                      {flag.DocTitle}
                    </a>
                  </li>
                </ul>
              </>
            ) : null}
          </>
        </p>
        <p>
          <>
            <strong>Entity/individual mentioned</strong>
            <ul className="bulleted-list">
              {[flag.EntityMentioned].map((item, index2) => {
                const key = `${index} ${index2}`;
                return (
                  <li key={key}>
                    <a>{item}</a>
                  </li>
                );
              })}
            </ul>
          </>
          <strong>Potential affiliate(s)</strong>
          <ul className="bulleted-list">
            {flag.DocEntities.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <li key={key}>
                  <a>{item}</a>
                </li>
              );
            })}
          </ul>
        </p>
        {isDisclosureChecked &&
          (flag.Disclosed ? (
            <button type="button" className="btn btn-success">
              Disclosed
            </button>
          ) : (
            <button type="button" className="btn btn-danger">
              Undisclosed
            </button>
          ))}
      </div>
    );
  }

  // function formalRelationFlag(flag, index) {
  //   return (
  //     <li key={index} className='p-3 px-5 w-75 detail-item'>
  //       {withPublicationDate(<h5 className='fw-bold mt-3'>{flag.title}</h5>, flag)}
  //       <p>{wrapLinks(flag.message)}</p>
  //     </li>
  //   )
  // }

  const [showPopover, setShowPopover] = useState(false);

  const handleResetFilter = () => {
    setStartDate('');
    setEndDate('');
    setReportContent(initialReprtContent);
  };

  const togglePopover = () => {
    setShowPopover(!showPopover);
  };
  function wrapLinks(origtext) {
    const linkStart = Math.max(origtext.indexOf('https://'), origtext.indexOf('http://'));
    if (linkStart === -1) {
      return [origtext];
    }
    const message = origtext.slice(0, linkStart);
    const link = origtext.slice(linkStart);
    return [
      message,
      <a href={link} target="_blank" rel="noopener noreferrer">
        {link}
      </a>,
    ];
  }

  const items = review ? reportContent[review] || [] : [];
  const hasDates = items.some(
    (item) => item?.Work?.PublicationDate && !isNaN(new Date(item.Work.PublicationDate).getTime())
  );

  return (
    (
      <div className="basic-setup">
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

              <div className="d-flex w-80">
                <div className="text-start px-5">
                  <div className="d-flex align-items-center mb-2">
                    <h5 className="m-0">{authorName}</h5>
                  </div>
                  <b className="m-0 p-0" style={{ fontSize: 'small' }}>
                    {institutions.join(', ')}
                  </b>
                </div>
              </div>

              <div>
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
                    onClick={() => {
                      handleDropdownChange(1);
                    }}
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
                  {dropdownOpen === 1 && (
                    <div
                      className="dropdown-menu show p-3"
                      style={{
                        width: '200px',
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
              </div>
            </div>
            {/* Comment the following to get rid of the graph tab */}
            <Tabs activeTab={activeTab} handleTabChange={handleTabChange} />
          </div>
          {activeTab === 0 && (
            <div className="d-flex justify-content-end mt-2 gap-2 px-2">
              <>
                <StyledWrapper>
                  <button
                    className="cssbuttons-io-button"
                    onClick={() => {
                      handleDropdownChange(3);
                    }}
                  >
                    Verify with Disclosures
                  </button>
                </StyledWrapper>
                <Dialog
                  open={dropdownOpen === 3}
                  onClose={handleCloseDialog}
                  maxWidth="sm"
                  fullWidth
                >
                  <DialogTitle sx={{ m: 0, p: 2 }} id="customized-dialog-title">
                    Select files to check for disclosure
                  </DialogTitle>
                  <Divider sx={{ color: 'black', backgroundColor: '#000000' }} />
                  <DialogContent>
                    <div
                      className="container"
                      onDrop={handleDrop}
                      onDragOver={(e) => e.preventDefault()}
                    >
                      <div className="header">
                        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path
                            d="M7 10V9C7 6.23858 9.23858 4 12 4C14.7614 4 17 6.23858 17 9V10C19.2091 10 21 11.7909 21 14C21 15.4806 20.1956 16.8084 19 17.5M7 10C4.79086 10 3 11.7909 3 14C3 15.4806 3.8044 16.8084 5 17.5M7 10C7.43285 10 7.84965 10.0688 8.24006 10.1959M12 12V21M12 12L15 15M12 12L9 15"
                            stroke="#000000"
                            strokeWidth="1.5"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          />
                        </svg>
                        <p>Drag & drop your file!</p>
                      </div>
                      <label htmlFor="file" className="footer">
                        <p>
                          {selectedFiles.length
                            ? `${selectedFiles.length} files selected`
                            : 'Click here to upload your file.'}
                        </p>
                      </label>
                      <input
                        id="file"
                        type="file"
                        multiple
                        onChange={handleFileSelect}
                        accept=".txt,.doc,.docx,.pdf"
                      />
                    </div>
                    {uploadError && (
                      <div style={{ color: 'red', marginTop: '10px' }}>{uploadError}</div>
                    )}
                  </DialogContent>
                  <DialogActions>
                    <Button onClick={handleCloseDialog}>Cancel</Button>
                    <Button onClick={handleSubmit} disabled={isUploading} variant="contained">
                      {isUploading ? 'Uploading...' : 'Submit'}
                    </Button>
                  </DialogActions>
                </Dialog>
              </>
              <DownloadButton
                reportId={report_id}
                isOpen={dropdownOpen === 2}
                setIsOpen={() => {
                  handleDropdownChange(2);
                }}
              />
            </div>
          )}
        </div>

        {activeTab === 0 &&
          // loading ? (<div style={{ width: '100%', height: '300px' }}>
          //   <Shimmer />
          // </div>) : 
          (
            <>
              {/* <div className="d-flex w-100 flex-column align-items-center">
                <div className="d-flex w-100 px-5 align-items-center my-2 mt-3 justify-content-between">
                  <div style={{ width: '20px' }}>
                    {loading && (
                      <div
                        className="spinner-border text-primary spinner-border-sm"
                        role="status"
                      ></div>
                    )}
                  </div>
                </div>
              </div> */}

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
                      onReview={() => setReview(flag)}
                      key={index}
                      selected={isSelected}
                    />
                  );
                })}
              </div>
              {review && (
                <div style={{ width: '100%', textAlign: 'center', marginTop: '50px' }}>
                  {hasDates && (
                    <div
                      style={{
                        marginBottom: '20px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        gap: '10px',
                      }}
                    >
                      <span style={{ marginRight: '10px' }}>Sort by Date</span>
                      <ArrowUpwardIcon
                        onClick={() => setSortOrder('asc')}
                        style={{
                          cursor: 'pointer',
                          color: sortOrder === 'asc' ? 'lightgray' : 'black',
                        }}
                      />
                      <ArrowDownwardIcon
                        onClick={() => setSortOrder('desc')}
                        style={{
                          cursor: 'pointer',
                          color: sortOrder === 'desc' ? 'lightgray' : 'black',
                        }}
                      />
                    </div>
                  )}
                  {isDisclosureChecked ? (
                    <>
                      {disclosedItems.length > 0 ? (
                        <>
                          <button
                            onClick={() => setShowDisclosed(!showDisclosed)}
                            style={{
                              backgroundColor: showDisclosed ? 'green' : 'transparent',
                              color: showDisclosed ? 'white' : 'green',
                              borderRadius: '20px',
                              border: '2px solid green',
                              padding: '10px 20px',
                              margin: '10px auto',
                              cursor: 'pointer',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              width: '200px',
                              fontSize: '16px',
                              transition: 'background-color 0.3s, color 0.3s',
                            }}
                          >
                            Disclosed ({disclosedItems.length})
                            {showDisclosed ? (
                              <ArrowDropDownIcon
                                style={{ verticalAlign: 'middle', marginLeft: '8px' }}
                              />
                            ) : (
                              <ArrowRightIcon
                                style={{ verticalAlign: 'middle', marginLeft: '8px' }}
                              />
                            )}
                          </button>
                          {showDisclosed && (
                            <div
                              style={{
                                width: '100%',
                                maxWidth: '1200px',
                                margin: '0 auto',
                                display: 'flex',
                                flexDirection: 'column',
                                alignItems: 'center',
                              }}
                            >
                              {renderFlags(disclosedItems)}
                            </div>
                          )}
                        </>
                      ) : (
                        <div
                          style={{
                            color: 'green',
                            margin: '10px auto',
                            width: '200px',
                            textAlign: 'center',
                            padding: '10px 20px',
                            border: '2px solid green',
                            borderRadius: '20px',
                          }}
                        >
                          Disclosed (0)
                        </div>
                      )}

                      {undisclosedItems.length > 0 ? (
                        <>
                          <button
                            onClick={() => setShowUndisclosed(!showUndisclosed)}
                            style={{
                              backgroundColor: showUndisclosed ? 'red' : 'transparent',
                              color: showUndisclosed ? 'white' : 'red',
                              borderRadius: '20px',
                              border: '2px solid red',
                              padding: '10px 20px',
                              margin: '10px auto',
                              cursor: 'pointer',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              width: '200px',
                              fontSize: '16px',
                              transition: 'background-color 0.3s, color 0.3s',
                            }}
                          >
                            Undisclosed ({undisclosedItems.length})
                            {showUndisclosed ? (
                              <ArrowDropDownIcon
                                style={{ verticalAlign: 'middle', marginLeft: '8px' }}
                              />
                            ) : (
                              <ArrowRightIcon
                                style={{ verticalAlign: 'middle', marginLeft: '8px' }}
                              />
                            )}
                          </button>
                          {showUndisclosed && (
                            <div
                              style={{
                                width: '100%',
                                maxWidth: '1200px',
                                margin: '0 auto',
                                display: 'flex',
                                flexDirection: 'column',
                                alignItems: 'center',
                              }}
                            >
                              {renderFlags(undisclosedItems)}
                            </div>
                          )}
                        </>
                      ) : (
                        <div
                          style={{
                            color: 'red',
                            margin: '10px auto',
                            width: '200px',
                            textAlign: 'center',
                            padding: '10px 20px',
                            border: '2px solid red',
                            borderRadius: '20px',
                          }}
                        >
                          Undisclosed (0)
                        </div>
                      )}
                    </>
                  ) : (
                    // When no disclosure check has been done, just show all items in one list.
                    <div
                      style={{
                        width: '100%',
                        maxWidth: '1200px',
                        margin: '0 auto',
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                      }}
                    >
                      {renderFlags(reportContent[review])}
                    </div>
                  )}
                </div>
              )}
            </>
          )}

        {activeTab === 1 && (
          <>
            <Graph authorName={authorName} reportContent={reportContent} />
          </>
        )}
      </div>
    )
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
