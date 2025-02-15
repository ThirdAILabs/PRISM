// src/ItemDetails.js
import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import {
  AUTHOR_AFFIL_EOC,
  FUNDER_EOC,
  PUBLISHER_EOC,
  ACK_EOC,
  COAUTHOR_AFFIL_EOC,
  COAUTHOR_EOC,
  MULTI_AFFIL,
  UNI_FACULTY_EOC,
  DOJ_PRESS_RELEASES_EOC
} from "../../../constants/constants.js";
import { useConclusions } from '../../../hooks/useConclusions.js';
import ConcernVisualizer from '../../ConcernVisualization.js';
import { useVisualizationData } from '../../../hooks/useVisualizationData.js';
import LoadedPapers from '../../common/cards/LoadedPapers.js'
import apiService from '../../../services/apiService.jsx';
import RelationShipGraph3 from '../../common/relationShipGraph/Relation-Graph3.js';
import Tabs from '../../common/tools/Tabs.js';

import { setStreamFlags } from '../../../services/streamStore.js';
import { reportService } from '../../../api/reports.js';


const FLAG_ID_ORDER = [
  AUTHOR_AFFIL_EOC, FUNDER_EOC, PUBLISHER_EOC, ACK_EOC, COAUTHOR_AFFIL_EOC, COAUTHOR_EOC, MULTI_AFFIL, UNI_FACULTY_EOC, DOJ_PRESS_RELEASES_EOC
];

const get_paper_url = (flag) => {
  return (
    <>
      <a href={flag.Work.WorkUrl} target="_blank" rel="noopener noreferrer">{flag.Work.DisplayName}</a>
      {
        flag.Work.OaUrl && <text> [<a href={flag.Work.OaUrl} target="_blank" rel="noopener noreferrer">full text</a>]</text>
      }
    </>
  );
}

const ItemDetails = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const reportId  = location.state.response.Id;
  // const reportId = "3b872f16-842b-499e-8001-1f8f617007c9"

  const [authorName, setAuthorName] = useState("")
  const [institutions, setInstitutions] = useState(["TODO INSTITUTIONS"])
  const [graphData, setGraphData] = useState(null)

  useEffect(() => {
    let isMounted = true;

    const poll = async () => {
      const report = await reportService.getReport(reportId)
      console.log("REPORT:", report)
      if (report.Status === "complete" && isMounted) {
        setIdToFlags(report.Content.type_to_flag)
        setAuthorName(report.AuthorName)
        setGraphData(report.Content)
        setLoading(false)
      } else if (isMounted) {
        setTimeout(poll, 500);
      }
    };

    poll();

    return () => {
      isMounted = false;
    };
  }, [])

  const [idToFlags, setIdToFlags] = useState({});


  const [loading, setLoading] = useState(true);
  const [worksCount, setWorksCount] = useState(0);
  const { formalRelations, loading: conclusionLoading } = useConclusions(authorName, idToFlags, /* threshold= */ 3);

  const [startYear, setStartYear] = useState('');
  const [endYear, setEndYear] = useState('');
  const [yearDropdownOpen, setYearDropdownOpen] = useState(false);
  const [activeTab, setActiveTab] = useState(0);

  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };
  const handleStartYearChange = (e) => setStartYear(e.target.value);
  const handleEndYearChange = (e) => setEndYear(e.target.value);
  const toggleYearDropdown = () => setYearDropdownOpen(!yearDropdownOpen);

  const [instDropdownOpen, setInstDropdownOpen] = useState(false);
  const toggleInstDropdown = () => setInstDropdownOpen(!instDropdownOpen);

  const { data, addAffiliations, setWeight, AffiliationChecklist } = useVisualizationData(authorName, idToFlags, formalRelations, worksCount, startYear, endYear);
  const [review, setReview] = useState();

  const [titles, setTitles] = useState();
  // console.log("Data coming from backend: ", data);
  // const streamFlags = useMemo(() => (result) => {
  //   const protocol = window.location.protocol.replace("http", "ws");  // Works whether it's http or https / ws or wss.
  //   const host = window.location.host.replace(":" + window.location.port, ":" + process.env.REACT_APP_BACKEND_PORT);
  //   const ws = new WebSocket(protocol + "//" + host + "/flags");
  //   console.log("ws", ws);

  //   websocket.current = ws;

  //   ws.onopen = () => {
  //     ws.send(JSON.stringify({ id: result.id, source: result.source, display_name: authorName, start_year: '', end_year: '' }));
  //   };

  //   ws.onmessage = (event) => {
  //     const response = JSON.parse(event.data);
  //     console.log('WebSocket response:', response);

  //     if (response.status === 'success' && response.flags) {
  //       console.log("About to set flags");
  //       setStreamFlags(response.flags);
  //       console.log("Set flags");
  //     }


  //     if (response.status === 'success' && response.update) {
  //       updateMessage();
  //       if (response.end_of_stream) {
  //         setMessage(248_000_000 + Math.floor(Math.random() * 2_000_000))
  //         setLoading(false);
  //       } else {
  //         if (response.update.message) {
  //           if ((response.update.message || '').includes("Loaded")) {
  //             console.log(response.update.message);
  //           }
  //           const match = response.update.message.match(/^Loaded (\d+) works$/);
  //           if (match) {
  //             setWorksCount(prev => Math.max(prev, parseInt(match[1])));
  //           }
  //         }
  //         if (response.update.flags) {
  //           setIdToFlags((prev) => {
  //             const newFlags = Object.fromEntries(Object.keys(prev).map(key => {
  //               return [key, [...prev[key]]];
  //             }));
  //             for (const flag of response.update.flags || []) {
  //               if (!newFlags[flag.flagger_id]) {
  //                 newFlags[flag.flagger_id] = [];
  //               }
  //               newFlags[flag.flagger_id].push(flag);
  //               addAffiliations(flag.affiliations);
  //             }
  //             return newFlags;
  //           });
  //         }
  //       }
  //     }

  //   };

  //   return () => {
  //     if (websocket.current) {
  //       websocket.current.close();
  //     }
  //   };
  // }, [])

  // useEffect(() => {
  //   setLoading(true);
  //   if (result.source !== "scopus") {
  //     streamFlags(result)
  //     return
  //   }
  //   apiService.scopusPaperTitles(result.id, (batch) => setTitles(prev => [...(prev || []), ...batch])).then(titles => {
  //     const newResult = { ...result, id: JSON.stringify(titles) };
  //     streamFlags(newResult);
  //   });


  // }, []);


  function withYear(header, flag) {
    return <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      {header}
      <span className='fw-bold mt-3'>{flag.Work.PublicationYear}</span>
    </div>
  }


  function multipleAffiliationsFlag(flag, index) {
    return (
      <li key={index} className='p-3 px-5 w-75 detail-item'>
        {withYear(<h5 className='fw-bold mt-3'>Author has multiple affiliations</h5>, flag)}
        <p>
          {flag.AuthorName} is affiliated with multiple institutions in {" "}
          {get_paper_url(flag)}
          . Detected affiliations:
          <ul className='bulleted-list'>
            {flag.Affiliations.map((item, index2) => {
              const key = `${index} ${index2}`;
              return <li key={key}><a>{item}</a></li>
            })}
          </ul>
        </p>
      </li>
    );
  }

  function funderFlag(flag, index) {
    return (
      <li key={index} className='p-3 px-5 w-75 detail-item'>
        {withYear(<h5 className='fw-bold mt-3'>Funder is an entity of concern</h5>, flag)}
        <p>
          {get_paper_url(flag)}
          {" "} is funded by the following entities of concern:
          <ul className='bulleted-list'>
            {flag.Funders.map((item, index2) => {
              const key = `${index} ${index2}`;
              return <li key={key}><a>{item}</a></li>
            })}
          </ul>
        </p>
      </li>
    );
  }

  function publisherFlag(flag, index) {
    return (
      <li key={index} className='p-3 px-5 w-75 detail-item'>
        {withYear(<h5 className='fw-bold mt-3'>Publisher is an entity of concern</h5>, flag)}
        <p>
          {get_paper_url(flag)}
          {" "} is published by the following entities of concern:
          <ul className='bulleted-list'>
            {flag.Publishers.map((item, index2) => {
              const key = `${index} ${index2}`;
              return <li key={key}><a>{item}</a></li>
            })}
          </ul>
        </p>
      </li>
    );
  }

  function coauthorFlag(flag, index) {
    return (
      <li key={index} className='p-3 px-5 w-75 detail-item'>
        {withYear(<h5 className='fw-bold mt-3'>Co-authors are high-risk entities</h5>, flag)}
        <p>
          The following co-authors of {" "}
          {get_paper_url(flag)}
          {" "} are high-risk entities:
          <ul className='bulleted-list'>
            {flag.Coauthors.map((item, index2) => {
              const key = `${index} ${index2}`;
              return <li key={key}><a>{item}</a></li>
            })}
          </ul>
        </p>
      </li>
    );
  }

  function coauthorAffiliationFlag(flag, index) {
    return (
      <li key={index} className='p-3 px-5 w-75 detail-item'>
        {withYear(<h5 className='fw-bold mt-3'>Co-authors are affiliated with entities of concern</h5>, flag)}
        <p>
          Some authors of {" "}
          {get_paper_url(flag)}
          {" "} are affiliated with entities of concern:
          <ul className='bulleted-list'>
            {flag.Institutions.map((item, index2) => {
              const key = `${index} ${index2}`;
              return <li key={key}><a>{item}</a></li>
            })}
          </ul>
          Affiliated authors:
          <ul className='bulleted-list'>
            {flag.Authors.map((item, index2) => {
              const key = `${index} ${index2}`;
              return <li key={key}><a>{item}</a></li>
            })}
          </ul>
        </p>
      </li>
    );
  }

  function authorAffiliationFlag(flag, index) {
    return (
      <li key={index} className='p-3 px-5 w-75 detail-item'>
        {withYear(<h5 className='fw-bold mt-3'>Author is affiliated with entities of concern</h5>, flag)}
        <p>
          {authorName} is affiliated with an entity of concern in {" "}
          {get_paper_url(flag)}
          . Detected affiliations:
          <ul className='bulleted-list'>
            {flag.Institutions.map((item, index2) => {
              const key = `${index} ${index2}`;
              return <li key={key}><a>{item}</a></li>
            })}
          </ul>
        </p>
      </li>
    );
  }

  function acknowledgementFlag(flag, index) {
    return (
      <li key={index} className='p-3 px-5 w-75 detail-item'>
        {withYear(<h5 className='fw-bold mt-3'>Acknowledgements contain foreign influences</h5>, flag)}
        <p>
          {
            flag.metadata.entities.length > 0
              ? <>
                {get_paper_url(flag)}
                {" "} acknowledges the following entities of concern:
              </>
              : <>
                Some acknowledged entities in {" "} {get_paper_url(flag)}
                {" "} may be foreign entities.
              </>
          }
          <ul className='bulleted-list'>
            {flag.metadata.entities.map((item, index2) => {
              const key = `${index} ${index2}`;
              return <li key={key}>
                <a>
                  "{item["entity"]}"
                  {" was detected in "}
                  <text style={{ "fontWeight": "bold", textDecoration: "underline" }}>{item["lists"].join(", ")}</text>
                  {" as "}
                  {item["aliases"].map(element => `"${element}"`).join(", ")}
                </a>
              </li>
            })}
          </ul>
          <p style={{ "marginTop": "20px", "fontWeight": "bold" }}>Acknowledgement text:</p>
          {
            flag.metadata.raw_acknowledgements.map((item, index3) => {
              return <p key={index3}>{item}</p>
            })
          }
          <p>{ }</p>
        </p>
        { }
      </li>
    );
  }

  function universityFacultyFlag(flag, index) {
    return (
      <li key={index} className='p-3 px-5 w-75 detail-item'>
        <h5 className='fw-bold mt-3'>The author may potentially be linked with an Entity of Concern</h5>
        <p>
          {flag.FlagMessage}
          <br />
          Relevant Webpage: <a href={flag.UniversityUrl} target="_blank" rel="noopener noreferrer">{flag.UniversityUrl}</a>
        </p>
      </li>
    );
  }

  function dojPRFlag(flag, index) {
    return (
      <li key={index} className='p-3 px-5 w-75 detail-item'>
        {true && (
          <>
            {flag.ConnectionLevel === 'primary' ? (
              <>
                <h5 className='fw-bold mt-3'>The author or an associate may be mentioned in a Press Release</h5>
              </>
            ) : flag.ConnectionLevel === 'secondary' ? (
              <>
                <h5 className='fw-bold mt-3'>The author's associate may be mentioned in a Press Release</h5>
              </>
            ) : flag.ConnectionLevel === 'tertiary' ? (
              <>
                <h5 className='fw-bold mt-3'>The author may potentially be connected to an entity/individual mentioned in a Press Release</h5>
              </>
            ) : null}
          </>
        )}
        <p>
          {flag.FlagMessage}
          {true && (
            <>
              <br />
              {flag.ConnectionLevel === 'primary' ? (
                <>
                  Press Release: <a href={flag.DocUrl} target="_blank" rel="noopener noreferrer">{flag.DocTitle}</a>
                  <br />
                  {/* Press Release: <a href={flag.metadata.doj_url} target="_blank" rel="noopener noreferrer">{flag.metadata.doj_title}</a>
                <br /> */}
                </>
              ) : flag.ConnectionLevel === 'secondary' ? (
                <>
                  {flag.FrequentCoauthor ? (
                    <>
                      Frequent Coauthor: {flag.FrequentCoauthor}
                    </>
                  ) :
                    <>
                      Relevant Document: <a href={flag.Nodes[0].DocUrl} target="_blank" rel="noopener noreferrer">{flag.Nodes[0].DocTitle}</a>
                    </>
                  }
                  <br />
                  Press Release: <a href={flag.DocUrl} target="_blank" rel="noopener noreferrer">{flag.DocTitle}</a>
                  <br />
                </>
              ) : flag.metadata.connection === 'tertiary' ? (
                <>
                  Relevant Documents: <a href={flag.Nodes[0].DocUrl} target="_blank" rel="noopener noreferrer">{flag.Nodes[0].DocTitle}</a>, <a href={flag.Nodes[1].DocUrl} target="_blank" rel="noopener noreferrer">{flag.Nodes[1].DocTitle}</a>
                  <br />
                  Press Release: <a href={flag.DocUrl} target="_blank" rel="noopener noreferrer">{flag.DocTitle}</a>
                  <br />
                </>
              ) : null}
            </>
          )}
          {/* <br />
          Relevant Document: <a href={flag.metadata.url} target="_blank" rel="noopener noreferrer">{flag.metadata.title}</a>
          {flag.metadata && flag.metadata.doj_url && flag.metadata.doj_title && (
            <>
              <br />
              Press Release: <a href={flag.metadata.doj_url} target="_blank" rel="noopener noreferrer">{flag.metadata.doj_title}</a>
            </>
          )} */}
        </p>
        <p>
          {true && (
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
          )}

          {/* Author/Frequent Collaborator:
        <ul className='bulleted-list'>
        {flag.metadata.associate_mentioned.map((item, index2) => {
            const key = `${index} ${index2}`;
            return <li key={key}><a>{item}</a></li>
          })}
        </ul> */}
          <strong>Potential affiliate(s)</strong>
          <ul className='bulleted-list'>
            {flag.DocEntities.map((item, index2) => {
              const key = `${index} ${index2}`;
              return <li key={key}><a>{item}</a></li>
            })}
          </ul>
        </p>
      </li>
    );
  }

  // function formalRelationFlag(flag, index) {
  //   return (
  //     <li key={index} className='p-3 px-5 w-75 detail-item'>
  //       {withYear(<h5 className='fw-bold mt-3'>{flag.title}</h5>, flag)}
  //       <p>{wrapLinks(flag.message)}</p>
  //     </li>
  //   )
  // }

  function makeFlag(flag, index) {
    switch (flag.FlagType) {
      case MULTI_AFFIL:
        return multipleAffiliationsFlag(flag, index);
      case FUNDER_EOC:
        return funderFlag(flag, index);
      case PUBLISHER_EOC:
        return publisherFlag(flag, index);
      case COAUTHOR_EOC:
        return coauthorFlag(flag, index);
      case COAUTHOR_AFFIL_EOC:
        return coauthorAffiliationFlag(flag, index);
      case AUTHOR_AFFIL_EOC:
        return authorAffiliationFlag(flag, index);
      case ACK_EOC:
        return acknowledgementFlag(flag, index);
      case UNI_FACULTY_EOC:
        return universityFacultyFlag(flag, index);
      case DOJ_PRESS_RELEASES_EOC:
        return dojPRFlag(flag, index);
      // case "formal_relations":
      //   return formalRelationFlag(flag, index);
      default:
        return (
          <li key={index} className='p-3 px-5 w-75 detail-item'>
            <h5 className='fw-bold mt-3'>{flag.FlagTitle}</h5>
            <p>{flag.FlagMessage}</p>
            {
              flag.Work && <p><a href={flag.Work.WorkUrl} target="_blank" rel="noopener noreferrer">{flag.Work.DisplayName}</a></p>
            }
          </li>
        );
    }
  }

  const [showPopover, setShowPopover] = useState(false);

  const togglePopover = () => {
    setShowPopover(!showPopover);
  };
  function wrapLinks(origtext) {
    const linkStart = Math.max(origtext.indexOf("https://"), origtext.indexOf("http://"));
    if (linkStart == -1) {
      return [origtext];
    }
    const message = origtext.slice(0, linkStart)
    const link = origtext.slice(linkStart)
    return [message, <a href={link} target="_blank" rel="noopener noreferrer">{link}</a>];
  }

  return (
    <div className='basic-setup'>
      <div className='flex flex-row'>
        <div className='detail-header'>
          <button onClick={() => navigate(-1)} className='btn text-light mb-3' style={{ minWidth: '80px', display: 'flex', alignItems: 'center' }}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" style={{ marginRight: '8px' }}>
              <path d="M10 19L3 12L10 5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
              <path d="M3 12H21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
            Back
          </button>

          <div className='d-flex w-80'>
            <div className='text-start px-5'>
              <div className='d-flex align-items-center mb-2'>
                <h5 className='m-0'>{authorName}</h5>
              </div>
              <b className='m-0 p-0' style={{ fontSize: 'small' }}>{institutions.join(', ')}</b>
            </div>
          </div>

          <div>

            <div className="dropdown">
              <style>
                {` 
                    .form-control::placeholder { 
                        color: #888; 
                    }`
                }
              </style>
              <button
                className="btn dropdown-toggle"
                type="button"
                onClick={toggleYearDropdown}
                style={{
                  backgroundColor: '#323232',
                  border: 'none',
                  color: 'white',
                  width: "200px",
                  fontWeight: 'bold',
                  fontSize: '14px'
                }}
              >
                Filter by Year
              </button>
              {yearDropdownOpen && (
                <div
                  className="dropdown-menu show p-3"
                  style={{
                    backgroundColor: '#323232',
                    border: 'none',
                    right: 0,
                    marginTop: "10px",
                    color: 'white',
                    fontWeight: 'bold',
                    fontSize: '14px'
                  }}
                >
                  <div className="form-group mb-2">
                    <label>Start Year</label>
                    <input
                      type="text"
                      value={startYear}
                      onChange={handleStartYearChange}
                      className="form-control"
                      placeholder="Enter start year"
                      style={{
                        backgroundColor: '#444444',
                        border: 'none',
                        outline: 'none',
                        color: 'white',
                        marginTop: '10px',
                      }}
                    />
                  </div>
                  <div style={{ height: "10px" }} />
                  <div className="form-group">
                    <label>End Year</label>
                    <input
                      type="text"
                      value={endYear}
                      onChange={handleEndYearChange}
                      className="form-control"
                      placeholder="Enter end year"
                      style={{
                        backgroundColor: '#444444',
                        border: 'none',
                        outline: 'none',
                        color: 'white',
                        marginTop: '10px',
                      }}
                    />
                  </div>
                </div>
              )}
            </div>

            <div style={{ height: "10px" }} />

            <div className="dropdown">
              <button
                className="btn dropdown-toggle"
                type="button"
                onClick={toggleInstDropdown}
                style={{
                  backgroundColor: '#323232',
                  border: 'none',
                  color: 'white',
                  width: "200px",
                  fontWeight: 'bold',
                  fontSize: '14px'
                }}
              >
                Filter by Institution
              </button>
              {instDropdownOpen && <AffiliationChecklist />}
            </div>
          </div>

        </div>
        {/* Comment the following to get rid of the graph tab */}
        <Tabs activeTab={activeTab} handleTabChange={handleTabChange} />
      </div>

      {activeTab === 0 && <>
        <div className='d-flex w-100 flex-column align-items-center'>
          <div className='d-flex w-75 align-items-center my-2 mt-3'>
            {(loading || conclusionLoading) && <div className="spinner-border text-primary spinner-border-sm" role="status"></div>}
            {/* {message && <h5 className='text-light m-0 ms-2' style={{ fontSize: 'small' }}>{(loading || conclusionLoading) ? `Scanned ${formattedMessage} out of 250M documents` : "Analysis complete"}</h5>} */}
            {/* {
            !(loading || conclusionLoading) && 
            <button
              type="button"
              className="btn btn-info btn-circle ml-2"
              onClick={togglePopover}
              style={buttonStyles}
            >
              ?
              {showPopover && (
                <div className="popover" style={popoverStyles}>
                  <div className="popover-header">Disclaimer</div>
                  <div className="popover-body">These flags are meant to act as an aid in detecting foreign influence. Please double check the results.</div>
                </div>
              )}
            </button>
          } */}
          </div>
        </div>

        {
          titles && <div className='d-flex w-100 flex-column align-items-center'>
            <LoadedPapers titles={titles} />
          </div>
        }

        <div className='d-flex w-100 flex-column align-items-center' style={{ color: "white", marginTop: "50px" }}>
          <div style={{ fontSize: "large", fontWeight: "bold" }}>
            Total Score
          </div>
          <div style={{ fontSize: "60px", fontWeight: "bold" }}>
            {
              Object.keys(data)
                .map(name => data[name].weight ? data[name].weight * data[name].details.length : 0)
                .reduce((prev, curr) => prev + curr, 0)
            }
          </div>
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-around', flexWrap: 'wrap', height: '500px', marginTop: '20px' }}>
          <ConcernVisualizer
            title={data.foreign_talent_programs.display_name}
            hoverText={data.foreign_talent_programs.desc}
            value={data.foreign_talent_programs.details.length}
            weight={data.foreign_talent_programs.weight}
            setWeight={(w) => setWeight("foreign_talent_programs", w)}
            onReview={() => setReview(["foreign_talent_programs", data.foreign_talent_programs.display_name])}
          />
          <ConcernVisualizer
            title={data.denied_entities.display_name}
            hoverText={data.denied_entities.desc}
            value={data.denied_entities.details.length}
            weight={data.denied_entities.weight}
            setWeight={(w) => setWeight("denied_entities", w)}
            onReview={() => setReview(["denied_entities", data.denied_entities.display_name])}
          />
          <ConcernVisualizer
            title={data.high_risk_funding_sources.display_name}
            hoverText={data.high_risk_funding_sources.desc}
            value={data.high_risk_funding_sources.details.length}
            weight={data.high_risk_funding_sources.weight}
            setWeight={(w) => setWeight("high_risk_funding_sources", w)}
            onReview={() => setReview(["high_risk_funding_sources", data.high_risk_funding_sources.display_name])}
          />
          <ConcernVisualizer
            title={data.high_risk_foreign_institutions.display_name}
            hoverText={data.high_risk_foreign_institutions.desc}
            value={data.high_risk_foreign_institutions.details.length}
            weight={data.high_risk_foreign_institutions.weight}
            setWeight={(w) => setWeight("high_risk_foreign_institutions", w)}
            onReview={() => setReview(["high_risk_foreign_institutions", data.high_risk_foreign_institutions.display_name])}
          />
          <ConcernVisualizer
            title={data.high_risk_appointments.display_name}
            hoverText={data.high_risk_appointments.desc}
            value={data.high_risk_appointments.details.length}
            weight={data.high_risk_appointments.weight}
            setWeight={(w) => setWeight("high_risk_appointments", w)}
            onReview={() => setReview(["high_risk_appointments", data.high_risk_appointments.display_name])}
          />
          <ConcernVisualizer
            title={data.university_faculty_appointments.display_name}
            hoverText={data.university_faculty_appointments.desc}
            value={data.university_faculty_appointments.details.length}
            weight={data.university_faculty_appointments.weight}
            setWeight={(w) => setWeight("university_faculty_appointments", w)}
            onReview={() => setReview(["university_faculty_appointments", data.university_faculty_appointments.display_name])}
          />
          <ConcernVisualizer
            title={data.doj_press_releases.display_name}
            hoverText={data.doj_press_releases.desc}
            value={data.doj_press_releases.details.length}
            weight={data.doj_press_releases.weight}
            setWeight={(w) => setWeight("doj_press_releases", w)}
            onReview={() => setReview(["doj_press_releases", data.doj_press_releases.display_name])}
          />
        </div>
        {/* Comment the following to get rid of the collapsible components */}
        {/* <CustomCollapsible /> */}
        {review && <ul className='d-flex flex-column align-items-center p-0' style={{ color: "white" }}>
          <h5 className='fw-bold mt-3'>Reviewing {" "} {review[1]}</h5>
          {data[review[0]].details.toSorted((a, b) => b.year - a.year).map((flag, index) => makeFlag(flag, index))}
        </ul>
        }
      </>
      }

      {activeTab === 1 && <>
        <div className='flex flex-row bg-black'>
          {/* <RelationShipGraph2 /> */}
          <RelationShipGraph3 graphData={graphData} />
        </div>
      </>}
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

export default ItemDetails;
