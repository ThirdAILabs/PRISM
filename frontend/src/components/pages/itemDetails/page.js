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
    COAUTHOR_AFFILIATIONS
} from "../../../constants/constants.js";
import ConcernVisualizer from '../../ConcernVisualization.js';
import RelationShipGraph3 from '../../common/relationShipGraph/Relation-Graph3.js';
import Tabs from '../../common/tools/Tabs.js';
import DownloadButton from '../../common/tools/button/downloadButton.js';
import { Dialog, DialogTitle, DialogContent, DialogActions, Button, Divider } from '@mui/material';
import { reportService } from '../../../api/reports.js';



const FLAG_ORDER = [
    TALENT_CONTRACTS, ASSOCIATIONS_WITH_DENIED_ENTITIES, HIGH_RISK_FUNDERS, AUTHOR_AFFILIATIONS, POTENTIAL_AUTHOR_AFFILIATIONS, MISC_HIGH_RISK_AFFILIATIONS, COAUTHOR_AFFILIATIONS
];

const TitlesAndDescriptions = {
    [TALENT_CONTRACTS]: {
        "title": "Talent Contracts",
        "desc": "Authors in these papers are recruited by talent programs that have close ties to high-risk foreign governments.",
    },
    [ASSOCIATIONS_WITH_DENIED_ENTITIES]: {
        "title": "Funding from Denied Entities",
        "desc": "Some of the parties involved in these works are in the denied entity lists of U.S. government agencies.",
    },
    [HIGH_RISK_FUNDERS]: {
        "title": "High Risk Funding Sources",
        "desc": "These papers are funded by funding sources that have close ties to high-risk foreign governments.",
    },
    [AUTHOR_AFFILIATIONS]: {
        "title": "Affiliations with High Risk Foreign Institutes",
        "desc": "Papers that list the queried author as being affiliated with a high-risk foreign institution or web pages that showcase official appointments at high-risk foreign institutions.",
    },
    [POTENTIAL_AUTHOR_AFFILIATIONS]: {
        "title": "Appointments at High Risk Foreign Institutes*",
        "desc": "The author may have an appointment at a high-risk foreign institutions.\n\n*Collated information from the web, might contain false positives.",
    },
    [MISC_HIGH_RISK_AFFILIATIONS]: {
        "title": "Miscellaneous High Risk Connections*",
        "desc": "The author or an associate may be mentioned in a press release.\n\n*Collated information from the web, might contain false positives.",
    },
    [COAUTHOR_AFFILIATIONS]: {
        "title": "Co-authors' affiliations with High Risk Foreign Institutes",
        "desc": "Coauthors in these papers are affiliated with high-risk foreign institutions."
    }
}

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
    const navigate = useNavigate();
    const { report_id } = useParams();

    const [reportContent, setReportContent] = useState({})
    const [authorName, setAuthorName] = useState("")
    const [institutions, setInstitutions] = useState([])
    const [initialReprtContent, setInitialReportContent] = useState({})
    // Add these states at the top with other states
    // const [selectedFiles, setSelectedFiles] = useState([]);
    // const [isUploading, setIsUploading] = useState(false);
    // const [uploadError, setUploadError] = useState(null);

    // // Add this function to handle file selection
    // const handleFileSelect = (event) => {
    //     setSelectedFiles([...event.target.files]);
    //     setUploadError(null);
    // };

    // // Add this function to handle check disclosure
    // const handleCheckDisclosure = async () => {
    //     if (selectedFiles.length === 0) {
    //         setUploadError('Please select at least one file');
    //         return;
    //     }

    //     setIsUploading(true);
    //     setUploadError(null);

    //     try {
    //         const result = await reportService.checkDisclosure(report_id, selectedFiles);
    //         setInitialReportContent(result.Content);
    //     } catch (error) {
    //         setUploadError(error.message || 'Failed to check disclosure');
    //     } finally {
    //         setIsUploading(false);
    //     }
    // };



    // Add new states
    const [openDialog, setOpenDialog] = useState(false);
    const [selectedFiles, setSelectedFiles] = useState([]);
    const [isUploading, setIsUploading] = useState(false);
    const [uploadError, setUploadError] = useState(null);

    // Add handlers
    const handleOpenDialog = () => setOpenDialog(true);
    const handleCloseDialog = () => {
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
            const report = await reportService.getReport(report_id)
            if (report.Status === "complete" && isMounted) {
                console.log("Report", report);
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
    }, [])




    const [loading, setLoading] = useState(true);

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
    const handleYearFilter = () => {
        const filteredContent = {};
        FLAG_ORDER.forEach((flag) => {
            if (initialReprtContent[flag]) {
                filteredContent[flag] = initialReprtContent[flag].filter((item) => {
                    return (item?.Work?.PublicationYear === undefined) || item?.Work?.PublicationYear >= startYear && item?.Work?.PublicationYear <= endYear;
                });
            }
            else
                filteredContent[flag] = null;
        });
        setReportContent(filteredContent);
        console.log("Unfiltered", initialReprtContent);
        console.log("Filtered", reportContent);
        setYearDropdownOpen(false);
    }
    const [instDropdownOpen, setInstDropdownOpen] = useState(false);
    const toggleInstDropdown = () => setInstDropdownOpen(!instDropdownOpen);

    const [review, setReview] = useState();

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
                    {authorName} is affiliated with multiple institutions in {" "}
                    {get_paper_url(flag)}
                    . Detected affiliations:
                    <ul className='bulleted-list'>
                        {flag.Affiliations.map((item, index2) => {
                            const key = `${index} ${index2}`;
                            return <li key={key}><a>{item}</a></li>
                        })}
                    </ul>
                </p>
                {flag.Disclosed ? (
                    <button type="button" className="btn btn-success">Disclosed</button>
                ) : (
                    <button type="button" className="btn btn-danger">Undisclosed</button>
                )}
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
                    {flag.Acknowledgements.length > 0 && (
                        <>
                            <p>Acknowledgements Text:</p>
                            <ul>
                                {flag.Acknowledgements.map((item, index2) => {
                                    const key = `ack-${index} ${index2}`;
                                    return <li key={key}>{item}</li>
                                })}
                            </ul>
                        </>
                    )}
                </p>
                {flag.Disclosed ? (
                    <button type="button" className="btn btn-success">Disclosed</button>
                ) : (
                    <button type="button" className="btn btn-danger">Undisclosed</button>
                )}
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
                {flag.Disclosed ? (
                    <button type="button" className="btn btn-success">Disclosed</button>
                ) : (
                    <button type="button" className="btn btn-danger">Undisclosed</button>
                )}
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
                {flag.Disclosed ? (
                    <button type="button" className="btn btn-success">Disclosed</button>
                ) : (
                    <button type="button" className="btn btn-danger">Undisclosed</button>
                )}
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
                        {flag.Affiliations.map((item, index2) => {
                            const key = `${index} ${index2}`;
                            return <li key={key}><a>{item}</a></li>
                        })}
                    </ul>
                    Affiliated authors:
                    <ul className='bulleted-list'>
                        {flag.Coauthors.map((item, index2) => {
                            const key = `${index} ${index2}`;
                            return <li key={key}><a>{item}</a></li>
                        })}
                    </ul>
                </p>
                {flag.Disclosed}
                {flag.Disclosed ? (
                    <button type="button" className="btn btn-success">Disclosed</button>
                ) : (
                    <button type="button" className="btn btn-danger">Undisclosed</button>
                )}
            </li>
        );
    }

    function authorAffiliationFlag(flag, index) {
        return (
            <li key={index} className='p-3 px-5 w-75 detail-item'>
                {withYear(<h5 className='fw-bold mt-3'>Author is affiliated with entities of concern</h5>, flag)}
                <div>
                    {authorName} is affiliated with an entity of concern in {" "}
                    {get_paper_url(flag)}
                    . Detected affiliations:
                    <ul className='bulleted-list'>
                        {flag.Affiliations.map((item, index2) => {
                            const key = `${index} ${index2}`;
                            return <li key={key}><a>{item}</a></li>
                        })}
                    </ul>
                </div>
                {flag.Disclosed}
                {flag.Disclosed ? (
                    <button type="button" className="btn btn-success">Disclosed</button>
                ) : (
                    <button type="button" className="btn btn-danger">Undisclosed</button>
                )}
            </li>
        );
    }

    function acknowledgementFlag(flag, index) {
        return (
            <li key={index} className='p-3 px-5 w-75 detail-item'>
                {withYear(<h5 className='fw-bold mt-3'>Acknowledgements contain foreign influences</h5>, flag)}
                <p>
                    {
                        flag.Entities > 0
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
                        {flag.Entities.map((item, index2) => {
                            const key = `${index} ${index2}`;
                            return <li key={key}>
                                <a>
                                    "{item.Entity}"
                                    {" was detected in "}
                                    <text style={{ "fontWeight": "bold", textDecoration: "underline" }}>{item.Sources.join(", ")}</text>
                                    {" as "}
                                    {item.Aliases.map(element => `"${element}"`).join(", ")}
                                </a>
                            </li>
                        })}
                    </ul>
                    <p style={{ "marginTop": "20px", "fontWeight": "bold" }}>Acknowledgement text:</p>
                    {
                        flag.RawAcknowledements.map((item, index3) => {
                            return <p key={index3}>{item}</p>
                        })
                    }
                    <p>{ }</p>
                </p>
                { }
                {flag.Disclosed ? (
                    <button type="button" className="btn btn-success">Disclosed</button>
                ) : (
                    <button type="button" className="btn btn-danger">Undisclosed</button>
                )}

                {/* DISCLOSURE */}
                {/* <hr />
                <div className="d-flex justify-content-between">
                    {flag.metadata?.dislosed === true ? (
                    <button type="button" className="btn btn-success">Disclosed</button>
                    ) : flag.metadata?.dislosed === false ? (
                    <button type="button" className="btn btn-danger">Undisclosed</button>
                    ) : (
                    // <button type="button" className="btn btn-secondary">No Disclosure Status</button> // Optional: Add a fallback state
                    <></>
                    )}
                    {flag.metadata?.is_primary_recipient === true ? (
                    <button type="button" className="btn btn-outline-success">Likely Not a Primary Recipient of Funds</button>
                    ) : flag.metadata?.is_primary_recipeitn === false ? (
                    <button type="button" className="btn btn-outline-danger">Likely a Primary Recipient of Funds</button>
                    ) : (
                    <></>
                    // <button type="button" className="btn btn-outline-secondary">No Grant Information</button>
                    )}
                </div> */}
            </li>
        );
    }

    function universityFacultyFlag(flag, index) {
        return (
            <li key={index} className='p-3 px-5 w-75 detail-item'>
                <h5 className='fw-bold mt-3'>The author may potentially be linked with an Entity of Concern</h5>
                <p>
                    {flag.Message}
                    <br />
                    Relevant Webpage: <a href={flag.UniversityUrl} target="_blank" rel="noopener noreferrer">{flag.UniversityUrl}</a>
                </p>
                {flag.Disclosed ? (
                    <button type="button" className="btn btn-success">Disclosed</button>
                ) : (
                    <button type="button" className="btn btn-danger">Undisclosed</button>
                )}
            </li>
        );
    }

    function PRFlag(flag, index) {
        const connections = flag.Connections || [];
        return (
            <li key={index} className='p-3 px-5 w-75 detail-item'>
                {true && (
                    <>
                        {connections.length == 0 ? (
                            <>
                                <h5 className='fw-bold mt-3'>The author or an associate may be mentioned in a Press Release</h5>
                            </>
                        ) : connections.length == 1 ? (
                            <>
                                <h5 className='fw-bold mt-3'>The author's associate may be mentioned in a Press Release</h5>
                            </>
                        ) : connections.length === 2 ? (
                            <>
                                <h5 className='fw-bold mt-3'>The author may potentially be connected to an entity/individual mentioned in a Press Release</h5>
                            </>
                        ) : null}
                    </>
                )}
                <p>
                    {flag.Message}
                    <>
                        <br />
                        {connections.length === 0 ? (
                            <>
                                Press Release: <a href={flag.DocUrl} target="_blank" rel="noopener noreferrer">{flag.DocTitle}</a>
                                <br />
                            </>
                        ) : connections.length === 1 ? (
                            <>
                                {flag.FrequentCoauthor ? (
                                    <>
                                        Frequent Coauthor: {flag.FrequentCoauthor}
                                    </>
                                ) :
                                    <>
                                        Relevant Document: <a href={flag.Connections[0].DocUrl} target="_blank" rel="noopener noreferrer">{flag.Connections[0].DocTitle}</a>
                                    </>
                                }
                                <br />
                                Press Release: <a href={flag.DocUrl} target="_blank" rel="noopener noreferrer">{flag.DocTitle}</a>
                                <br />
                            </>
                        ) : connections.length == 2 ? (
                            <>
                                Relevant Documents: <a href={flag.Connections[0].DocUrl} target="_blank" rel="noopener noreferrer">{flag.Connections[0].DocTitle}</a>, <a href={flag.Connections[1].DocUrl} target="_blank" rel="noopener noreferrer">{flag.Connections[1].DocTitle}</a>
                                <br />
                                Press Release: <a href={flag.DocUrl} target="_blank" rel="noopener noreferrer">{flag.DocTitle}</a>
                                <br />
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
                    <ul className='bulleted-list'>
                        {flag.DocEntities.map((item, index2) => {
                            const key = `${index} ${index2}`;
                            return <li key={key}><a>{item}</a></li>
                        })}
                    </ul>
                </p>
                {flag.Disclosed ? (
                    <button type="button" className="btn btn-success">Disclosed</button>
                ) : (
                    <button type="button" className="btn btn-danger">Undisclosed</button>
                )}
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
            <div className='grid grid-cols-2 gap-4'>
                <div className='flex flex-row'>
                    <div className='detail-header'>
                        <button onClick={() => navigate("/")} className='btn text-dark mb-3' style={{ minWidth: '80px', display: 'flex', alignItems: 'center' }}>
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
                                        backgroundColor: 'rgb(160, 160, 160)',
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
                                            backgroundColor: 'rgb(160, 160, 160)',
                                            border: 'none',
                                            right: 0,
                                            marginTop: "10px",
                                            color: 'white',
                                            fontWeight: 'bold',
                                            fontSize: '14px',
                                            justifyContent: 'center',
                                            alignItems: 'center',
                                            display: 'flex',
                                            flexDirection: 'column'
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
                                                    backgroundColor: 'rgb(220, 220, 220)',
                                                    border: 'none',
                                                    outline: 'none',
                                                    color: 'black',
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
                                                    backgroundColor: 'rgb(220, 220, 220)',
                                                    border: 'none',
                                                    outline: 'none',
                                                    color: 'black',
                                                    marginTop: '10px',
                                                }}
                                            />
                                        </div>
                                        <button
                                            className="form-control"
                                            type="submit"
                                            onClick={handleYearFilter}
                                            disabled={!(startYear && endYear)}
                                            style={{
                                                backgroundColor: 'rgb(220, 220, 220)',
                                                border: 'none',
                                                color: 'white',
                                                width: "100px",
                                                fontWeight: 'bold',
                                                fontSize: '14px',
                                                marginTop: '20px',
                                            }}
                                        >
                                            submit
                                        </button>
                                    </div>
                                )}
                            </div>
                            <>
                                <button
                                    className="form-control"
                                    onClick={handleOpenDialog}
                                    style={{
                                        backgroundColor: 'rgb(160, 160, 160)',
                                        border: 'none',
                                        color: 'white',
                                        width: "200px",
                                        fontWeight: 'bold',
                                        fontSize: '14px',
                                        marginTop: '20px',
                                    }}
                                >
                                    Check disclosure
                                </button>

                                <Dialog
                                    open={openDialog}
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
                                                    <path d="M7 10V9C7 6.23858 9.23858 4 12 4C14.7614 4 17 6.23858 17 9V10C19.2091 10 21 11.7909 21 14C21 15.4806 20.1956 16.8084 19 17.5M7 10C4.79086 10 3 11.7909 3 14C3 15.4806 3.8044 16.8084 5 17.5M7 10C7.43285 10 7.84965 10.0688 8.24006 10.1959M12 12V21M12 12L15 15M12 12L9 15" stroke="#000000" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                                                </svg>
                                                <p>Browse File to upload!</p>
                                            </div>
                                            <label htmlFor="file" className="footer">
                                                <p>{selectedFiles.length ? `${selectedFiles.length} files selected` : 'No file selected'}</p>
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
                                        <Button
                                            onClick={handleSubmit}
                                            disabled={isUploading}
                                            variant="contained"
                                        >
                                            {isUploading ? 'Uploading...' : 'Submit'}
                                        </Button>
                                    </DialogActions>
                                </Dialog>
                            </>
                        </div>

                    </div>
                    {/* Comment the following to get rid of the graph tab */}
                    <Tabs activeTab={activeTab} handleTabChange={handleTabChange} />
                </div>
                <div className='d-flex justify-content-end mt-2'>
                    <DownloadButton reportId={report_id} />
                </div>
            </div>


            {activeTab === 0 && <>
                <div className='d-flex w-100 flex-column align-items-center'>
                    <div className='d-flex w-100 px-5 align-items-center my-2 mt-3 justify-content-between'>
                        <div style={{ width: "20px" }}>
                            {loading && (
                                <div className="spinner-border text-primary spinner-border-sm" role="status"></div>
                            )}
                        </div>
                    </div>
                </div>

                <div className='d-flex w-100 flex-column align-items-center' style={{ color: "rgb(78, 78, 78)", marginTop: "0px" }}>
                    <div style={{ fontSize: "large", fontWeight: "bold" }}>
                        Total Score
                    </div>
                    <div style={{ fontSize: "60px", fontWeight: "bold" }}>
                        {
                            Object.keys(reportContent || {})
                                .map(name => (reportContent[name] || []).length)
                                .reduce((prev, curr) => prev + curr, 0)
                        }
                    </div>
                </div>

                <div style={{ display: 'flex', justifyContent: 'space-around', flexWrap: 'wrap', height: '500px', marginTop: '20px' }}>
                    {
                        FLAG_ORDER.map((flag, index) => {
                            return <ConcernVisualizer
                                title={TitlesAndDescriptions[flag].title}
                                hoverText={TitlesAndDescriptions[flag].desc}
                                value={reportContent[flag] ? reportContent[flag].length : 0}
                                weight={1}
                                onReview={() => setReview(flag)}
                                key={index}
                            />
                        })
                    }
                </div>
                {/* Comment the following to get rid of the collapsible components */}
                {/* <CustomCollapsible /> */}
                {review && <ul className='d-flex flex-column align-items-center p-0' style={{ color: "black" }}>
                    <h5 className='fw-bold mt-3'>Reviewing {" "} {TitlesAndDescriptions[review].title}</h5>{
                        (reportContent[review] || []).map(
                            (flag, index) => {
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
                                }
                            }
                        )
                    }
                </ul>
                }
            </>
            }

            {activeTab === 1 && <>
                <RelationShipGraph3 authorName={authorName} reportContent={reportContent} />
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