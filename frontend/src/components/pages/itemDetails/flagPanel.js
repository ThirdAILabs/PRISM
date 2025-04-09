import React, { useState, useEffect } from 'react';
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
import { getRawTextFromXML } from '../../../utils/helper.js';
import useOutsideClick from '../../../hooks/useOutsideClick.js';
import '../../../styles/components/_flagPanel.scss';
import { Divider } from '@mui/material';
import { IoMdClose } from 'react-icons/io';
import { ChevronDown } from 'lucide-react';
import FlagContainer from './flagContainer.js';
import { createHighlights, applyHighlighting, hasValidTriangulationData } from './ack_utils.js';

const FlagPanel = ({ reportContent, review, setReview, authorName, isDisclosureChecked }) => {
  const [isRendered, setIsRendered] = useState(false);
  const [activeTab, setActiveTab] = useState('all');
  const [isSortByDropdownOpen, setIsSortByDropdownOpen] = useState(false);
  const [sortOrder, setSortOrder] = useState('Latest To Oldest');

  useEffect(() => {
    if (review) {
      // Small delay to ensure DOM is ready
      requestAnimationFrame(() => {
        setIsRendered(true);
      });
    } else {
      setIsRendered(false);
    }
  }, [review]);

  const sidepanelRef = useOutsideClick(() => {
    setIsRendered(false);
    setTimeout(() => setReview(''), 300);
  });

  const toggleDropdown = () => {
    setIsSortByDropdownOpen(!isSortByDropdownOpen);
  };

  const selectOption = (option) => {
    setSortOrder(option);
    setIsSortByDropdownOpen(false);
  };

  const get_paper_url = (flag) => {
    // getRawTextFromXML(flag.Work.DisplayName);
    return (
      <>
        <a href={flag.Work.WorkUrl} target="_blank" rel="noopener noreferrer">
          {getRawTextFromXML(flag.Work.DisplayName)}
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
      return sortOrder === 'Latest To Oldest' ? dateA - dateB : dateB - dateA;
    });

    return sortedItems.map((flag, index) => {
      let flagContent;
      switch (review) {
        case TALENT_CONTRACTS:
          flagContent = acknowledgementFlag(flag, index);
          break;
        case ASSOCIATIONS_WITH_DENIED_ENTITIES:
          flagContent = acknowledgementFlag(flag, index);
          break;
        case HIGH_RISK_FUNDERS:
          flagContent = funderFlag(flag, index);
          break;
        case AUTHOR_AFFILIATIONS:
          flagContent = authorAffiliationFlag(flag, index);
          break;
        case POTENTIAL_AUTHOR_AFFILIATIONS:
          flagContent = universityFacultyFlag(flag, index);
          break;
        case MISC_HIGH_RISK_AFFILIATIONS:
          flagContent = PRFlag(flag, index);
          break;
        case COAUTHOR_AFFILIATIONS:
          flagContent = coauthorAffiliationFlag(flag, index);
          break;
        default:
          return null;
      }

      return (
        <FlagContainer
          key={index}
          isDisclosureChecked={isDisclosureChecked}
          isDisclosed={flag.Disclosed}
        >
          {flagContent}
        </FlagContainer>
      );
    });
  }

  function withPublicationDate(header, flag) {
    const publicationDateStr = flag?.Work?.PublicationDate;
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
      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        {header}
        <span className="fw-bold mt-3">{formattedDate}</span>
      </div>
    );
  }

  function acknowledgementSection(flag, authorName, index) {
    // Quick validation of required data
    if (!Array.isArray(flag.RawAcknowledgements) || flag.RawAcknowledgements.length === 0) {
      return null;
    }

    // Check for triangulation data
    const triangulationResult = hasValidTriangulationData(flag.FundCodeTriangulation);
    const notContainPR = triangulationResult.notContainPR;
    const containPR = triangulationResult.containPR;
    const hasTriangulationData = notContainPR || containPR;

    // Create highlights configuration once
    const highlights = hasTriangulationData
      ? createHighlights(flag.FundCodeTriangulation, authorName)
      : [];

    return (
      <div className="flag-sub-container">
        <div className="acknowledgement-header">
          <strong>Acknowledgements Text</strong>
          {hasTriangulationData && (
            <div className="triangulation-indicators">
              {notContainPR && <span className="triangulation-tag success">Not primary recipient</span>}
              {containPR && <span className="triangulation-tag danger">Primary recipient</span>}
            </div>
          )}
        </div>

        <ul className="bulleted-list">
          {flag.RawAcknowledgements.map((item, itemIndex) => {
            const key = `ack-${index}-${itemIndex}`;
            return (
              <li key={key}>
                {hasTriangulationData ? applyHighlighting(item, highlights) : item}
              </li>
            );
          })}
        </ul>
      </div>
    );
  }

  function acknowledgementFlag(flag, index) {
    return (
      <div>
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
          {acknowledgementSection(flag, authorName, index)}
        </p>
      </div>
    );
  }


  function funderFlag(flag, index) {
    return (
      <div>
        {withPublicationDate(
          <h5 className="fw-bold mt-3">Funder is an Entity of Concern</h5>,
          flag
        )}
        <p>{get_paper_url(flag)} is funded by the entities of concern</p>

        {Array.isArray(flag.Funders) && flag.Funders.length > 0 && (
          <div className="flag-sub-container">
            <strong>Concerned entities</strong>
            <div className="concerned-tags">
              {flag.Funders.map((item, index2) => {
                const key = `${index} ${index2}`;
                return (
                  <span key={key} className="concerned-tag-item">
                    {item}
                  </span>
                );
              })}
            </div>
          </div>
        )}
        {acknowledgementSection(flag, authorName, index)}
      </div>
    );
  }

  function publisherFlag(flag, index) {
    return (
      <div>
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
      <div>
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
      <div>
        {withPublicationDate(
          <h5 className="fw-bold mt-">Co-authors are affiliated with Entities of Concern</h5>,
          flag
        )}
        <p>Some authors of {get_paper_url(flag)} are affiliated with entities of concern:</p>
        <div className="flag-sub-container">
          <strong>Concerning entity</strong>
          <div className="concerned-tags">
            {flag.Affiliations.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <span key={key} className="concerned-tag-item">
                  {item}
                </span>
              );
            })}
          </div>
        </div>
        <div className="flag-sub-container">
          <strong>Affiliated author(s)</strong>
          <div className="concerned-tags">
            {flag.Coauthors.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <span key={key} className="concerned-tag-item">
                  {item}
                </span>
              );
            })}
          </div>
        </div>
      </div>
    );
  }

  function authorAffiliationFlag(flag, index) {
    return (
      <div>
        {withPublicationDate(
          <h5 className="fw-bold mt-3">Author is affiliated with an Entity of Concern</h5>,
          flag
        )}
        <div>
          {authorName} is affiliated with an entity of concern in {get_paper_url(flag)}.<p></p>
          <div className="flag-sub-container">
            <strong>Detected Affiliations</strong>
            <div className="concerned-tags">
              {flag.Affiliations.map((item, index2) => {
                const key = `${index} ${index2}`;
                return (
                  <span key={key} className="concerned-tag-item">
                    {item}
                  </span>
                );
              })}
            </div>
          </div>
        </div>
      </div>
    );
  }

  function universityFacultyFlag(flag, index) {
    return (
      <div>
        <h5 className="fw-bold mt-3">
          The author may potentially be linked with an Entity of Concern
        </h5>
        <div className="flag-sub-container">
          <strong>Concering entity</strong>
          <div className="concerned-tags">
            <span className="concerned-tag-item">{flag.University}</span>
          </div>
        </div>
        <div className="flag-sub-container">
          <strong>Relevant Webpage</strong>
          <a href={flag.UniversityUrl} target="_blank" rel="noopener noreferrer">
            {flag.UniversityUrl}
          </a>
        </div>
      </div>
    );
  }

  function PRFlag(flag, index) {
    const connections = flag.Connections || [];

    function pressRelease() {
      return (
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
      );
    }

    function relevantDocuments() {
      return (
        <>
          <strong>Relevant Document(s)</strong>
          <ul className="bulleted-list">
            {connections.map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <li key={key}>
                  <a href={item.DocUrl} target="_blank" rel="noopener noreferrer">
                    {item.DocTitle}
                  </a>
                </li>
              );
            })}
          </ul>
        </>
      );
    }
    return (
      <div>
        <h5 className="flag-header">
          {connections.length == 0
            ? 'The author or an associate may be mentioned in a Press Release'
            : connections.length == 1
              ? "The author's associate may be mentioned in a Press Release"
              : connections.length == 2
                ? 'The author may potentially be connected to an entity/individual mentioned in a Press Release'
                : ''}
        </h5>
        <p>{flag.Message}</p>
        {connections.length == 1 && (
          <div className="flag-sub-container">
            {flag.FrequentCoauthor ? (
              <>Frequent Coauthor: {flag.FrequentCoauthor}</>
            ) : (
              relevantDocuments()
            )}
          </div>
        )}
        {connections.length == 2 && <div className="flag-sub-container">{relevantDocuments()}</div>}
        <div className="flag-sub-container">{pressRelease()}</div>
        <div className="flag-sub-container">
          <strong>Entity/individual mentioned</strong>
          <div className="concerned-tags">
            {[flag.EntityMentioned].map((item, index2) => {
              const key = `${index} ${index2}`;
              return (
                <span key={key} className="concerned-tag-item">
                  {item}
                </span>
              );
            })}
          </div>
        </div>
        {flag.DocEntities && flag.DocEntities.length > 0 && (
          <div className="flag-sub-container">
            <strong>Potential affiliate(s)</strong>
            <div className="concerned-tags">
              {flag.DocEntities.map((item, index2) => {
                const key = `${index} ${index2}`;
                return (
                  <span key={key} className="concerned-tag-item">
                    {item}
                  </span>
                );
              })}
            </div>
          </div>
        )}
      </div>
    );
  }

  function sortByComponent() {
    const items = reportContent[review] || [];
    const hasDates = items.some(
      (item) => item?.Work?.PublicationDate && !isNaN(new Date(item.Work.PublicationDate).getTime())
    );
    if (!hasDates) return null;

    return (<div className="sort-dropdown">
    <div className="sort-dropdown__toggle" onClick={toggleDropdown}>
      <span className="sort-dropdown__label">Sort By</span>
      <ChevronDown className="sort-dropdown__icon" />
    </div>
    
    {isSortByDropdownOpen && (
      <div className="sort-dropdown__menu">
        <div 
          className="sort-dropdown__option"
          onClick={() => selectOption('Latest To Oldest')}
        >
          Latest To Oldest
        </div>
        <div 
          className="sort-dropdown__option"
          onClick={() => selectOption('Oldest To Latest')}
        >
          Oldest To Latest
        </div>
      </div>
    )}
  </div>
);
  }

  return (
    <div ref={sidepanelRef} className={`flag-panel ${isRendered ? 'open' : ''}`}>
      <div className="flag-panel-header">
        <h4 className="flag-panel-title">{TitlesAndDescriptions[review].title}</h4>
        <button className="flag-panel-close-button">
          <IoMdClose />
        </button>
      </div>
      <Divider className="divider" />
      <div className="selection-container">
        <div className="score-box-container">
          <div className="score-pill">
            <span className="score-label">Score:</span>
            <span className="score-value">
              {reportContent[review] ? reportContent[review].length : 0}
            </span>
          </div>

          <div className="tab-group">
            <button
              className={`tab-button ${activeTab === 'all' ? 'active' : ''}`}
              onClick={() => {
                setActiveTab('all');
              }}
            >
              All
            </button>
            {isDisclosureChecked && (
              <>
                <button
                  className={`tab-button ${activeTab === 'disclosed' ? 'active' : ''}`}
                  onClick={() => {
                    setActiveTab('disclosed');
                  }}
                >
                  Disclosed
                </button>
                <button
                  className={`tab-button ${activeTab === 'undisclosed' ? 'active' : ''}`}
                  onClick={() => {
                    setActiveTab('undisclosed');
                  }}
                >
                  Undisclosed
                </button>
              </>
            )}
          </div>

          {sortByComponent()}
        </div>
      </div>
      {(() => {
        let itemsToRender = reportContent[review];

        if (activeTab === 'disclosed') {
          itemsToRender = itemsToRender.filter((item) => item.Disclosed);
        } else if (activeTab === 'undisclosed') {
          itemsToRender = itemsToRender.filter((item) => !item.Disclosed);
        }

        return (
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
            {renderFlags(itemsToRender)}
          </div>
        );
      })()}
    </div>
  );
};

export default FlagPanel;
