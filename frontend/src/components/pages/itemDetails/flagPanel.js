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
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import ArrowRightIcon from '@mui/icons-material/ArrowRight';
import { getRawTextFromXML } from '../../../utils/helper.js';
import useOutsideClick from '../../../hooks/useOutsideClick.js';
import '../../../styles/components/_flagPanel.scss';
import { Divider } from '@mui/material';
import { IoMdClose } from "react-icons/io";
import { IoIosArrowDown } from "react-icons/io";
import { IoIosArrowUp } from "react-icons/io";


const FlagPanel = ({ reportContent, review, setReview, authorName, isDisclosureChecked, disclosedItems, showDisclosed, setShowDisclosed, undisclosedItems, showUndisclosed, setShowUndisclosed}) => {    
  const [isRendered, setIsRendered] = useState(false);

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

    // box shadow for disclosed/undisclosed buttons
    const greenBoxShadow = '0 0px 10px rgb(0, 183, 46)';
    const redBoxShadow = '0 0px 10px rgb(255, 0, 0)';
    const [sortOrder, setSortOrder] = useState('desc');
  
    const toggleSortOrder = () => {
      setSortOrder((prevOrder) => (prevOrder === 'asc' ? 'desc' : 'asc'));
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
            <br />
            {flag.FundCodeTriangulation &&
            typeof flag.FundCodeTriangulation === 'object' &&
            Object.keys(flag.FundCodeTriangulation).length > 0 ? (
              <>
                {flag.RawAcknowledgements?.map((item, index2) => {
                  const key = `ack-${index} ${index2}`;
  
                  const highlights = Object.entries(flag.FundCodeTriangulation || {}).flatMap(
                    ([funderName, funderMap]) =>
                      Object.entries(funderMap).flatMap(([grantCode, isRecipient]) => {
                        if (typeof isRecipient === 'boolean') {
                          return [
                            {
                              regex: new RegExp(grantCode, 'i'),
                              color: isRecipient ? 'red' : 'green',
                              tooltip: `${authorName} is ${isRecipient ? 'likely' : 'likely NOT'} a primary recipient of this fund.`,
                            },
                          ];
                        }
                        return [];
                      })
                  );
  
                  let parts = [item];
  
                  highlights.forEach(({ regex, color, tooltip }) => {
                    parts = parts.flatMap((text, i) =>
                      typeof text === 'string'
                        ? text.split(regex).flatMap((part, j, arr) =>
                            j < arr.length - 1
                              ? [
                                  part,
                                  <span key={`${i}-${j}`} style={{ color }} title={tooltip}>
                                    <strong>{text.match(regex)[0]}</strong>
                                  </span>,
                                ]
                              : part
                          )
                        : text
                    );
                  });
  
                  return <p key={index2}>{parts}</p>;
                })}
                {triangulationLegend()}
              </>
            ) : (
              <>
                {flag.RawAcknowledgements?.map((item, index3) => {
                  return <p key={index3}>{item}</p>;
                })}
              </>
            )}
          </p>
        </div>
      );
    }

    function triangulationLegend() {
      return (
        <div className="mt-4 d-flex flex-column small">
          <span className="me-3">
            <span
              className="rounded-circle d-inline-block me-2"
              style={{ width: '8px', height: '8px', backgroundColor: 'green' }}
            ></span>
            The author likely <b>is not</b> a primary recipient of these high-risk grants.
          </span>
          <span>
            <span
              className="rounded-circle d-inline-block me-2"
              style={{ width: '8px', height: '8px', backgroundColor: 'red' }}
            ></span>
            The author likely <b>is</b> a primary recipient of these high-risk grants.
          </span>
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
                {Array.isArray(flag.Funders) &&
                    flag.Funders.length > 0 &&
                    flag.Funders.map((item, index2) => {
                    const key = `${index} ${index2}`;
                    return (
                        <li key={key}>
                        <a>{item}</a>
                        </li>
                    );
                    })}
                </ul>
                {Array.isArray(flag.RawAcknowledgements) && flag.RawAcknowledgements?.length > 0 && (
                <>
                    <strong>Acknowledgements Text</strong>
                    {flag.FundCodeTriangulation &&
                    typeof flag.FundCodeTriangulation === 'object' &&
                    Object.keys(flag.FundCodeTriangulation).length > 0 ? (
                    <>
                        <ul className="bulleted-list">
                        {flag.RawAcknowledgements?.map((item, index2) => {
                            const key = `ack-${index} ${index2}`;

                            const highlights = Object.entries(flag.FundCodeTriangulation || {}).flatMap(
                            ([funderName, funderMap]) =>
                                Object.entries(funderMap).flatMap(([grantCode, isRecipient]) => {
                                if (typeof isRecipient === 'boolean') {
                                    return [
                                    {
                                        regex: new RegExp(grantCode, 'i'),
                                        color: isRecipient ? 'red' : 'green',
                                        tooltip: `${authorName} is ${isRecipient ? 'likely' : 'likely NOT'} a primary recipient of this fund.`,
                                    },
                                    ];
                                }
                                return [];
                                })
                            );

                            let parts = [item];

                            highlights.forEach(({ regex, color, tooltip }) => {
                            parts = parts.flatMap((text, i) =>
                                typeof text === 'string'
                                ? text.split(regex).flatMap((part, j, arr) =>
                                    j < arr.length - 1
                                        ? [
                                            part,
                                            <span key={`${i}-${j}`} style={{ color }} title={tooltip}>
                                            <strong>{text.match(regex)[0]}</strong>
                                            </span>,
                                        ]
                                        : part
                                    )
                                : text
                            );
                            });

                            return <li key={key}>{parts}</li>;
                        })}
                        </ul>
                        {triangulationLegend()}
                    </>
                    ) : (
                    <ul className="bulleted-list">
                        {flag.RawAcknowledgements?.map((item, index2) => {
                        const key = `ack-${index} ${index2}`;
                        return <li key={key}>{item}</li>;
                        })}
                    </ul>
                    )}
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
                <br />
                {flag.FundCodeTriangulation &&
                typeof flag.FundCodeTriangulation === 'object' &&
                Object.keys(flag.FundCodeTriangulation).length > 0 ? (
                <>
                    {flag.RawAcknowledgements?.map((item, index2) => {
                    const key = `ack-${index} ${index2}`;

                    const highlights = Object.entries(flag.FundCodeTriangulation || {}).flatMap(
                        ([funderName, funderMap]) =>
                        Object.entries(funderMap).flatMap(([grantCode, isRecipient]) => {
                            if (typeof isRecipient === 'boolean') {
                            return [
                                {
                                regex: new RegExp(grantCode, 'i'),
                                color: isRecipient ? 'red' : 'green',
                                tooltip: `${authorName} is ${isRecipient ? 'likely' : 'likely NOT'} a primary recipient of this fund.`,
                                },
                            ];
                            }
                            return [];
                        })
                    );

                    let parts = [item];

                    highlights.forEach(({ regex, color, tooltip }) => {
                        parts = parts.flatMap((text, i) =>
                        typeof text === 'string'
                            ? text.split(regex).flatMap((part, j, arr) =>
                                j < arr.length - 1
                                ? [
                                    part,
                                    <span key={`${i}-${j}`} style={{ color }} title={tooltip}>
                                        <strong>{text.match(regex)[0]}</strong>
                                    </span>,
                                    ]
                                : part
                            )
                            : text
                        );
                    });

                    return <p key={index2}>{parts}</p>;
                    })}
                    {triangulationLegend()}
                </>
                ) : (
                <>
                    {flag.RawAcknowledgements?.map((item, index3) => {
                    return <p key={index3}>{item}</p>;
                    })}
                </>
                )}
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
              {flag.DocEntities && flag.DocEntities.length > 0 && (
                <>
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
                </>
              )}
            </p>
          </div>
        );
    }

    function sortByComponent(){
      const items = reportContent[review] || [];
      const hasDates = items.some(
        (item) =>
          item?.Work?.PublicationDate &&
          !isNaN(new Date(item.Work.PublicationDate).getTime())
      );
      if (!hasDates) 
        return null;

      return (
        <div className="sort-by-container">
          <span className="sort-label">Sort by Date</span>
          <div onClick={toggleSortOrder} className="sort-icon">
            {sortOrder === 'asc' ? (
              <IoIosArrowUp />
            ) : (
              <IoIosArrowDown />
            )}
          </div>
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
          <Divider className='divider'/>
          <div className='selection-container'>
            <div className="score-box-container">
              <div className="score-pill">
                <span className="score-label">Score:</span>
                <span className="score-value">{reportContent[review] ? reportContent[review].length : 0}</span>
              </div>
              { isDisclosureChecked && (
                <div className="tab-group">
                  <button className="tab-button active">All</button>
                  <button className="tab-button">Disclosed</button>
                  <button className="tab-button">Undisclosed</button>
                </div>
                )
              }
              {sortByComponent()}
            </div>
          </div>
          

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
                      <ArrowRightIcon style={{ verticalAlign: 'middle', marginLeft: '8px' }} />
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
                      <ArrowRightIcon style={{ verticalAlign: 'middle', marginLeft: '8px' }} />
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
        
    );
}
export default FlagPanel;