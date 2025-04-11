import '../../../styles/components/_authorInfoCard.scss';
import Scholar from '../../../assets/icons/Scholar.svg';
import University from '../../../assets/icons/University.svg';
import Research from '../../../assets/icons/Research.svg';
import { Divider } from '@mui/material';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import CloudDownloadIcon from '@mui/icons-material/CloudDownload';
import { getTrailingWhiteSpace } from '../../../utils/helper';
import DownloadDropdown from '../../common/tools/button/downloadButton';
import { useState } from 'react';
import useOutsideClick from '../../../hooks/useOutsideClick';
import { TextField, Popover, Button } from '@mui/material';
import { LocalizationProvider, DatePicker } from '@mui/x-date-pickers';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';

const AuthorInfoCard = ({ result, verifyWithDisclosure, downloadProps, filterProps }) => {

    const [downloadDropdownOpen, setDownloadDropdownOpen] = useState(false);
    const [filterDropdownOpen, setFilterDropdownOpen] = useState(false);
    const [isFilterActive, setIsFilterActive] = useState(false);

    const dropdownDownloadRef = useOutsideClick(() => {
        setDownloadDropdownOpen(false);
    });

    const filterRef = useOutsideClick(() => {
        setFilterDropdownOpen(false);
    });

    const handleFilterClick = () => {
        setFilterDropdownOpen(!filterDropdownOpen);
    };

    const handleApplyFilter = () => {
        setFilterDropdownOpen(false);
        setIsFilterActive(true);
        filterProps.handleDateFilter();
    };

    const handleClearFilter = () => {
        setFilterDropdownOpen(false);
        setIsFilterActive(false);
        filterProps.handleClearFilter();
    }
    return (
        <div className="text-start" style={{ padding: '20px 20px 4px 30px', width: '100%' }}>
            <div className="info-row">
                <div className="title-container">
                    <img src={Scholar} alt="Scholar" className="icon scholar" />
                    <h5 className="title">{result.AuthorName}</h5>
                </div>
                <div className={`filter-container ${isFilterActive ? 'active' : ''}`}
                    ref={filterRef}
                >
                    <div onClick={handleFilterClick}>

                        {isFilterActive ? (
                            <div>
                                <span className='filter-container-content filter-icon active' >Filter By Date</span>
                                <KeyboardArrowDownIcon
                                    className="filter-icon active"
                                />
                            </div>
                        ) : (
                            <div>
                                <span className='filter-container-content' >Filter By Date</span>
                                <KeyboardArrowDownIcon
                                    className="filter-icon"
                                />
                            </div>
                        )}
                    </div>
                    {filterDropdownOpen && <div
                        className="dropdown-menu"
                        style={{
                            backgroundColor: 'rgb(255, 255, 255)',
                            border: '1px solid rgb(230, 230, 230)',
                            borderRadius: '8px',
                            marginTop: '6px',
                            color: 'white',
                            fontWeight: 'bold',
                            fontSize: '14px',
                            justifyContent: 'center',
                            alignItems: 'center',
                            display: 'flex',
                            flexDirection: 'column',
                            marginLeft: '-44px'
                        }}
                    >
                        <div
                            className="form-group"
                            style={{ marginBottom: '4px', width: '100%', padding: '7px', paddingTop: '0px' }}
                        >
                            <span style={{
                                color: 'black',
                                fontWeight: '400',
                                fontSize: '12px',
                                lineHeight: '18px',
                                letterSpacing: '0.25px'
                            }}>
                                Start Date
                            </span>
                            <input
                                type="date"
                                className="form-control"
                                value={filterProps.startDate}
                                max={filterProps.todayStr}
                                onChange={filterProps.handleStartDateChange}
                                style={{
                                    backgroundColor: 'rgb(255, 255, 255)',
                                    border: '1px solid rgb(230, 230, 230)',
                                    borderRadius: '8px',
                                    outline: 'none',
                                    color: 'black',
                                    marginTop: '5px',
                                    width: '100%',
                                }}
                            />
                        </div>
                        <div
                            className="form-group"
                            style={{ marginBottom: '8px', width: '100%', padding: '0px 7px' }}
                        >
                            <span style={{
                                color: 'black',
                                fontWeight: '400',
                                fontSize: '12px',
                                lineHeight: '18px',
                                letterSpacing: '0.25px'
                            }}>
                                End Date
                            </span>
                            <input
                                type="date"
                                value={filterProps.endDate}
                                max={filterProps.todayStr}
                                onChange={filterProps.handleEndDateChange}
                                className="form-control"
                                style={{
                                    backgroundColor: 'rgb(255, 255, 255)',
                                    border: '1px solid rgb(230, 230, 230)',
                                    borderRadius: '8px',
                                    outline: 'none',
                                    color: 'black',
                                    marginTop: '5px',
                                    width: '100%',
                                }}
                            />
                        </div>
                        <div className='filter-container-buttonGroup'>
                            <button
                                className="form-control"
                                type="submit"
                                onClick={handleClearFilter}
                                style={{
                                    backgroundColor: 'rgb(130, 130, 130)',
                                    border: 'none',
                                    color: 'white',
                                    border: '1px solid rgb(230, 230, 230)',
                                    borderRadius: '8px',
                                    width: '70px',
                                    fontWeight: 'bold',
                                    fontSize: '14px',
                                    transition: 'background-color 0.3s',
                                    marginTop: '4px',
                                }}
                            >
                                Clear
                            </button>
                            <button
                                className="form-control hover:bg-blue-500"
                                type="submit"
                                onClick={handleApplyFilter}
                                disabled={!(filterProps.startDate || filterProps.endDate)}
                                style={{
                                    backgroundColor: '#64b6f7',
                                    color: 'white',
                                    border: '1px solid rgb(230, 230, 230)',
                                    borderRadius: '8px',
                                    width: '70px',
                                    fontWeight: 'bold',
                                    fontSize: '14px',
                                    cursor: filterProps.startDate || filterProps.endDate ? 'pointer' : 'not-allowed',
                                    transition: 'background-color 0.3s',
                                    marginTop: '4px',
                                }}
                            >
                                Apply
                            </button>
                        </div>
                    </div>}
                </div>
            </div>

            <div className="info-row" style={{ marginTop: '10px' }}>
                <img src={University} alt="Affiliation" className="icon" />
                <span className="content">
                    <span className="content-research">{result.Institutions[0]}</span>
                    {result.Institutions.length > 1 && ', ' + result.Institutions.slice(1).join(', ')}
                </span>
            </div>

            {result.Interests && result.Interests.length > 0 && (
                <div className="info-row">
                    <img src={Research} alt="Research" className="icon" />
                    <span className="content content-research">
                        {result.Interests.slice(0, 3).join(', ')}
                    </span>
                </div>
            )}
            <Divider
                sx={{
                    backgroundColor: 'black',
                    height: '1px',
                    width: '100%',
                    opacity: 0.1,
                    margin: '12px -30px 0px -30px', // Negative margins to counter parent padding
                    borderRadius: '8px',
                    '&.MuiDivider-root': {
                        width: 'calc(100% + 50px)', // Compensate for parent padding
                    },
                }}
            />
            <div className="button-group">
                <span onClick={verifyWithDisclosure.handleFileUploadClick}>
                    {'Verify With Disclosures' + getTrailingWhiteSpace(2)} <UploadFileIcon color="info" />
                    <input
                        type="file"
                        ref={verifyWithDisclosure.fileInputRef}
                        style={{ display: 'none' }}
                        multiple
                        accept=".txt,.pdf"
                        onChange={verifyWithDisclosure.handleFileSelect}
                    />
                </span>
                <span onClick={() => { setDownloadDropdownOpen(!downloadDropdownOpen) }} ref={dropdownDownloadRef}>
                    {'Download Report' + getTrailingWhiteSpace(4)}
                    <CloudDownloadIcon color="info" />
                    <DownloadDropdown
                        reportId={downloadProps.reportId}
                        metadata={downloadProps.metadata}
                        content={downloadProps.content}
                        isOpen={downloadDropdownOpen}
                        setIsOpen={setDownloadDropdownOpen}
                    />
                </span>
            </div>
        </div>
    );
};

export default AuthorInfoCard;
