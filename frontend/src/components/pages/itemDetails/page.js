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
} from '../../../constants/constants.js';
import ConcernVisualizer, { BaseFontSize, getFontSize } from '../../ConcernVisualization.js';
import Graph from '../../common/graph/graph.js';
import Tabs from '../../common/tools/Tabs.js';
import DownloadButton from '../../common/tools/button/downloadButton.js';
import { reportService } from '../../../api/reports.js';
import styled from 'styled-components';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import ArrowRightIcon from '@mui/icons-material/ArrowRight';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import Shimmer from './Shimmer.js';
import MuiAlert from '@mui/material/Alert';
import { Snackbar } from '@mui/material';
import useGoBack from '../../../hooks/useGoBack.js';
import useOutsideClick from '../../../hooks/useOutsideClick.js';

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

  const [yearDropdownOpen, setYearDropdownOpen] = useState(false);
  const [downloadDropdownOpen, setDownloadDropdownOpen] = useState(false);
  const [reportContent, setReportContent] = useState({});
  const [authorName, setAuthorName] = useState('');
  const [institutions, setInstitutions] = useState([]);
  const [initialReprtContent, setInitialReportContent] = useState({});
  const [isDisclosureChecked, setDisclosureChecked] = useState(false);
  const [valueFontSize, setValueFontSize] = useState(`${BaseFontSize}px`);

  // box shadow for disclosed/undisclosed buttons
  const greenBoxShadow = '0 0px 10px rgb(0, 183, 46)';
  const redBoxShadow = '0 0px 10px rgb(255, 0, 0)';

  const toggleSortOrder = () => {
    setSortOrder((prevOrder) => (prevOrder === 'asc' ? 'desc' : 'asc'));
  };

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
      setReportContent(result.Content);
      setInitialReportContent(result.Content);
      setDisclosureChecked(true);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
      setNotification({
        open: true,
        severity: 'success',
        message: 'Disclosure check succeeded!',
      });
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
    const poll = async () => {
      let inProgress = true;
      const report = await reportService.getReport(report_id);
      if (report.Content && Object.keys(report.Content).length > 0 && isMounted) {
        console.log('Report', report);
        setAuthorName(report.AuthorName);
        setReportContent(report.Content);
        setInitialReportContent(report.Content);
        setLoading(false);

        const newFontSize = `${getFontSize(
          Math.max(...FLAG_ORDER.map((flag) => report.Content.Flags[flag]?.length || 0))
        )}px`;

        inProgress = report.Status === 'queued' || report.Status === 'in-progress';
      }

      if (inProgress) {
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
  const [activeTab, setActiveTab] = useState(0);

  const [filterMessage, setFilterMessage] = useState('');

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
      setReportContent(initialReprtContent);
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
      if (initialReprtContent[flag]) {
        filteredContent[flag] = initialReprtContent[flag].filter((item) => {
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

    setFilterMessage(`${displayStart} to ${displayEnd}`);

    setStartDate('');
    setEndDate('');

    setReportContent(filteredContent);

    // font size change maybe needed
    const newFontSize = `${getFontSize(
      Math.max(...FLAG_ORDER.map((flag) => filteredContent[flag]?.length || 0))
    )}px`;
    setValueFontSize(newFontSize);
  };

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
      <div
        key={index}
        className="p-3 px-5 w-75 detail-item"
        style={{
          boxShadow: !isDisclosureChecked ? 'none' : flag.Disclosed ? greenBoxShadow : redBoxShadow,
        }}
      >
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
      </div>
    );
  }

  function funderFlag(flag, index) {
    return (
      <div
        key={index}
        className="p-3 px-5 w-75 detail-item"
        style={{
          boxShadow: !isDisclosureChecked ? 'none' : flag.Disclosed ? greenBoxShadow : redBoxShadow,
        }}
      >
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
      </div>
    );
  }

  function publisherFlag(flag, index) {
    return (
      <div
        key={index}
        className="p-3 px-5 w-75 detail-item"
        style={{
          boxShadow: !isDisclosureChecked ? 'none' : flag.Disclosed ? greenBoxShadow : redBoxShadow,
        }}
      >
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
      </div>
    );
  }

  function coauthorFlag(flag, index) {
    return (
      <div
        key={index}
        className="p-3 px-5 w-75 detail-item"
        style={{
          boxShadow: !isDisclosureChecked ? 'none' : flag.Disclosed ? greenBoxShadow : redBoxShadow,
        }}
      >
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
      </div>
    );
  }

  function coauthorAffiliationFlag(flag, index) {
    return (
      <div
        key={index}
        className="p-3 px-5 w-75 detail-item"
        style={{
          boxShadow: !isDisclosureChecked ? 'none' : flag.Disclosed ? greenBoxShadow : redBoxShadow,
        }}
      >
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
      </div>
    );
  }

  function authorAffiliationFlag(flag, index) {
    return (
      <div
        key={index}
        className="p-3 px-5 w-75 detail-item"
        style={{
          boxShadow: !isDisclosureChecked ? 'none' : flag.Disclosed ? greenBoxShadow : redBoxShadow,
        }}
      >
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
      </div>
    );
  }

  function acknowledgementFlag(flag, index) {
    return (
      <div
        key={index}
        className="p-3 px-5 w-75 detail-item"
        style={{
          boxShadow: !isDisclosureChecked ? 'none' : flag.Disclosed ? greenBoxShadow : redBoxShadow,
        }}
      >
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
          <p>{}</p>
        </p>
      </div>
    );
  }

  function universityFacultyFlag(flag, index) {
    return (
      <div
        key={index}
        className="p-3 px-5 w-75 detail-item"
        style={{
          boxShadow: !isDisclosureChecked ? 'none' : flag.Disclosed ? greenBoxShadow : redBoxShadow,
        }}
      >
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
      </div>
    );
  }

  const [showDisclosed, setShowDisclosed] = useState(false);
  const [showUndisclosed, setShowUndisclosed] = useState(false);
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
      <div
        key={index}
        className="p-3 px-5 w-75 detail-item"
        style={{
          boxShadow: !isDisclosureChecked ? 'none' : flag.Disclosed ? greenBoxShadow : redBoxShadow,
        }}
      >
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
              padding: '10px 0', // reduce padding
            }}
          >
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
            <div style={{ textAlign: 'center' }}>
              <h5 className="m-0">{authorName}</h5>
              <b className="m-0 p-0" style={{ fontSize: 'small' }}>
                {institutions.join(', ')}
              </b>
            </div>
            <div>
              <div className="dropdown" ref={dropdownFilterRef}>
                <style>
                  {`
                    .form-control::placeholder {
                      color: #888;
                    }
                  `}
                </style>
                <button
                  className="btn dropdown-toggle"
                  type="button"
                  onClick={() => setYearDropdownOpen(!yearDropdownOpen)}
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
                    className="dropdown-menu show p-2"
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
                    <div className="form-group" style={{ marginBottom: '10px', width: '100%' }}>
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
                    <div className="form-group" style={{ marginBottom: '10px', width: '100%' }}>
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
                )}
              </div>
            </div>
          </div>
          {/* Render the Tabs with the filterMessage */}
          <Tabs
            activeTab={activeTab}
            handleTabChange={handleTabChange}
            filterMessage={filterMessage}
          />
        </div>
        {activeTab === 0 && (
          <div className="d-flex justify-content-end mt-2 gap-2 px-2">
            <StyledWrapper>
              <button className="cssbuttons-io-button" onClick={handleFileUploadClick}>
                Verify with Disclosures
              </button>
            </StyledWrapper>
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
                isOpen={downloadDropdownOpen}
                setIsOpen={() => setDownloadDropdownOpen(!downloadDropdownOpen)}
              />
            </div>
          </div>
        )}
      </div>

      {activeTab === 0 &&
        (loading ? (
          <div style={{ width: '100%', height: '300px' }}>
            <Shimmer />
          </div>
        ) : (
          <>
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
            {review && (
              <div style={{ width: '100%', textAlign: 'center', marginTop: '50px' }}>
                {(() => {
                  const items = reportContent[review] || [];
                  const hasDates = items.some(
                    (item) =>
                      item?.Work?.PublicationDate &&
                      !isNaN(new Date(item.Work.PublicationDate).getTime())
                  );
                  if (!hasDates) return null;
                  return (
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
                      <div onClick={toggleSortOrder} style={{ cursor: 'pointer' }}>
                        {sortOrder === 'asc' ? (
                          <ArrowUpwardIcon style={{ color: 'black' }} />
                        ) : (
                          <ArrowDownwardIcon style={{ color: 'black' }} />
                        )}
                      </div>
                    </div>
                  );
                })()}

                {isDisclosureChecked ? (
                  <>
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'center',
                        gap: '20px',
                        margin: '10px auto',
                        width: 'fit-content',
                      }}
                    >
                      {/* Disclosed Button */}
                      {disclosedItems.length > 0 ? (
                        <button
                          onClick={() => {
                            setShowDisclosed(!showDisclosed);
                            if (!showDisclosed) setShowUndisclosed(false);
                          }}
                          style={{
                            backgroundColor: 'transparent',
                            color: 'green',
                            boxShadow: showDisclosed ? '0 0px 10px rgb(0, 183, 46)' : 'none',
                            borderRadius: '20px',
                            border: '2px solid green',
                            padding: '10px 10px',
                            cursor: 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            width: '200px',
                            fontSize: '16px',
                            transition: 'background-color 0.3s, color 0.3s',
                          }}
                        >
                          <strong>Disclosed ({disclosedItems.length})</strong>
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
                      ) : (
                        <div
                          style={{
                            color: 'green',
                            textAlign: 'center',
                            padding: '10px 20px',
                            border: '2px solid green',
                            borderRadius: '20px',
                            width: '200px',
                          }}
                        >
                          <strong>Disclosed (0)</strong>
                        </div>
                      )}

                      {/* Undisclosed Button */}
                      {undisclosedItems.length > 0 ? (
                        <button
                          onClick={() => {
                            setShowUndisclosed(!showUndisclosed);
                            if (!showUndisclosed) setShowDisclosed(false);
                          }}
                          style={{
                            backgroundColor: 'transparent',
                            color: 'red',
                            boxShadow: showUndisclosed ? '0 0px 10px rgb(255, 0, 0)' : 'none',
                            borderRadius: '20px',
                            border: '2px solid red',
                            padding: '10px 10px',
                            cursor: 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            width: '200px',
                            fontSize: '16px',
                            transition: 'background-color 0.3s, color 0.3s',
                          }}
                        >
                          <strong>Undisclosed ({undisclosedItems.length})</strong>
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
                      ) : (
                        <div
                          style={{
                            color: 'red',
                            textAlign: 'center',
                            padding: '10px 20px',
                            border: '2px solid red',
                            borderRadius: '20px',
                            width: '200px',
                          }}
                        >
                          <strong>Undisclosed (0)</strong>
                        </div>
                      )}
                    </div>

                    {/* Content areas for disclosed and undisclosed items */}
                    {/* Display flags below buttons */}
                    <div style={{ width: '100%', marginTop: '20px' }}>
                      {showDisclosed && (
                        <div
                          style={{
                            width: '100%',
                            maxWidth: '1200px',
                            margin: '10px auto',
                            display: 'flex',
                            flexDirection: 'column',
                            alignItems: 'center',
                          }}
                        >
                          {renderFlags(disclosedItems)}
                        </div>
                      )}

                      {showUndisclosed && (
                        <div
                          style={{
                            width: '100%',
                            maxWidth: '1200px',
                            margin: '10px auto',
                            display: 'flex',
                            flexDirection: 'column',
                            alignItems: 'center',
                          }}
                        >
                          {renderFlags(undisclosedItems)}
                        </div>
                      )}
                    </div>
                  </>
                ) : (
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
        ))}

      {activeTab === 1 && <Graph authorName={authorName} reportContent={reportContent} />}
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
