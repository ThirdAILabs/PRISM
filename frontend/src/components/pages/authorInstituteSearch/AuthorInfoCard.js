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
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import EmailRoundedIcon from '@mui/icons-material/EmailRounded';
import MailOutlineRoundedIcon from '@mui/icons-material/MailOutlineRounded';
import { TextField } from '@mui/material';

import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
} from '@mui/material';


const AuthorInfoCard = ({ result, verifyWithDisclosure, downloadProps, filterProps, loading }) => {
  const [downloadDropdownOpen, setDownloadDropdownOpen] = useState(false);
  const [filterDropdownOpen, setFilterDropdownOpen] = useState(false);
  const [isFilterActive, setIsFilterActive] = useState(false);
  const [emailUpdateDiaLogBox, setEmailUpdateDiaLogBox] = useState(false);
  const [emailFrequency, setEmailFrequency] = useState(0);
  const [customDays, setCustomDays] = useState('');
  const [customDaysError, setCustomDaysError] = useState('');
  const [isCustom, setIsCustom] = useState(false);

  const emeialUpdateDialogBoxRef = useOutsideClick(() => {
    setEmailUpdateDiaLogBox(false);
  });

  const dropdownDownloadRef = useOutsideClick(() => {
    setDownloadDropdownOpen(false);
  });

  const filterRef = useOutsideClick(() => {
    setFilterDropdownOpen(false);
  });

  const handleFilterClick = () => {
    if (!loading) setFilterDropdownOpen(!filterDropdownOpen);
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
  };

  const handleEmailDialogOpen = () => {
    setEmailUpdateDiaLogBox(true);
  };

  const handleEmailDialogClose = () => {
    setEmailUpdateDiaLogBox(false);
  };

  const handleCustomDaysChange = (event) => {
    const value = event.target.value;
    setCustomDays(value);
    
    if (value && (isNaN(value) || parseInt(value) < 15 || parseInt(value) > 365)) {
      setCustomDaysError('Must be a number between 15 and 365');
    } else {
      setCustomDaysError('');
    }
  };

  const handleFrequencyChange = (event) => {
    const value = event.target.value;
    setEmailFrequency(value);
    setIsCustom(value === 'custom');
    if (value !== 'custom') {
      setCustomDays('');
      setCustomDaysError('');
    }
  };

  const handleEmailUpdateSubmit = () => {
    // Handle the email update subscription here
    console.log('Email frequency:', emailFrequency);
    setEmailUpdateDiaLogBox(false);
  };

  return (
    <div className="text-start card-container" style={{ width: '100%' }}>
      <div className="card-top" style={{ padding: '20px 20px 4px 30px', flexGrow: 1 }}>
        <div className="info-row">
          <div className="title-container">
            <img src={Scholar} alt="Scholar" className="icon scholar" />
            <h5 className="title">{result.AuthorName}</h5>
            <button
              className="email-updates-button"
              onClick={handleEmailDialogOpen}
              disabled={loading}
              title="Subscribe to Email Updates"
            >
              <EmailRoundedIcon />
            </button>
          </div>

          <div className={`filter-container ${isFilterActive ? 'active' : ''}`} ref={filterRef}>
            <div onClick={handleFilterClick}>
              {isFilterActive ? (
                <div>
                  <span className="filter-container-content filter-icon active">
                    Filter By Date
                  </span>
                  <KeyboardArrowDownIcon className="filter-icon active" />
                </div>
              ) : (
                <div>
                  <span className="filter-container-content">Filter By Date</span>
                  <KeyboardArrowDownIcon className="filter-icon" />
                </div>
              )}
            </div>
            {filterDropdownOpen && (
              <div
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
                  marginLeft: '-44px',
                }}
              >
                <div
                  className="form-group"
                  style={{ marginBottom: '4px', width: '100%', padding: '7px', paddingTop: '0px' }}
                >
                  <span
                    style={{
                      color: 'black',
                      fontWeight: '400',
                      fontSize: '12px',
                      lineHeight: '18px',
                      letterSpacing: '0.25px',
                    }}
                  >
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
                  <span
                    style={{
                      color: 'black',
                      fontWeight: '400',
                      fontSize: '12px',
                      lineHeight: '18px',
                      letterSpacing: '0.25px',
                    }}
                  >
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
                <div className="filter-container-buttonGroup">
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
                      cursor:
                        filterProps.startDate || filterProps.endDate ? 'pointer' : 'not-allowed',
                      transition: 'background-color 0.3s',
                      marginTop: '4px',
                    }}
                  >
                    Apply
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>

        <div className="info-row" style={{ marginTop: '10px' }}>
          <img src={University} alt="Affiliation" className="icon" />
          <span className="content">
            {result.Institutions && result.Institutions[0] !== '' ? (
              <>
                <span className="content-research">{result.Institutions[0]}</span>
                {result.Institutions.length > 1 && ', ' + result.Institutions.slice(1).join(', ')}
              </>
            ) : (
              <span className="content-research-placeholder">
                No institute information available
              </span>
            )}
          </span>
        </div>

        {result.Interests && result.Interests[0] !== '' ? (
          <div className="info-row">
            <img src={Research} alt="Research" className="icon" />
            <span className="content content-research">
              {result.Interests.slice(0, 3).join(', ')}
            </span>
          </div>
        ) : (
          <div className="info-row">
            <img src={Research} alt="Research" className="icon" />
            <span className="content content-research-placeholder">
              No research interests available
            </span>
          </div>
        )}
      </div>
      <Divider
        sx={{
          backgroundColor: 'black',
          height: '1px',
          opacity: 0.1,
          margin: '12px 0px 0px 0px',
          borderRadius: '8px',
        }}
      />
      <div
        className="card-footer"
        style={{ height: '40px', display: 'flex', alignItems: 'center' }}
      >
        <div className="button-group" style={{ width: '100%', height: '36px' }}>
          <span onClick={verifyWithDisclosure.handleFileUploadClick}>
            {'Verify With Disclosures' + getTrailingWhiteSpace(2)}{' '}
            <UploadFileIcon color={loading ? 'disabled' : 'info'} />
            <input
              type="file"
              ref={verifyWithDisclosure.fileInputRef}
              style={{ display: 'none' }}
              multiple
              accept=".txt,.pdf"
              onChange={verifyWithDisclosure.handleFileSelect}
              disabled={loading}
            />
          </span>
          <span
            onClick={() => {
              if (!loading) setDownloadDropdownOpen(!downloadDropdownOpen);
            }}
            ref={dropdownDownloadRef}
          >
            {'Download Report' + getTrailingWhiteSpace(4)}
            <CloudDownloadIcon color={loading ? 'disabled' : 'info'} />
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

      <Dialog 
        open={emailUpdateDiaLogBox} 
        onClose={handleEmailDialogClose}
        maxWidth="sm"
        fullWidth
        className='email-dialog'
      >
        <DialogTitle className="email-dialog-title">
          <span className="email-dialog-icon">
            <MailOutlineRoundedIcon />
          </span>
          <span className="email-dialog-title-text">
            Set Up Email Updates
          </span>
        </DialogTitle>
        <DialogContent className="email-dialog-content">
          <FormControl fullWidth className="email-frequency-select">
            <InputLabel>Email Frequency</InputLabel>
            <Select
              value={emailFrequency}
              label="Email Frequency"
              onChange={handleFrequencyChange}
            >
              <MenuItem value={15}>Bi-weekly</MenuItem>
              <MenuItem value={30}>Monthly</MenuItem>
              <MenuItem value = {90}>Quarterly</MenuItem>
              <MenuItem value="custom">Custom</MenuItem>
            </Select>
            {isCustom && (
              <TextField
                className="custom-days-input"
                label="Number of days"
                type="number"
                value={customDays}
                onChange={handleCustomDaysChange}
                error={!!customDaysError}
                helperText={customDaysError}
              />
            )}
            <div className="helper-text">
              Select how often you'd like to receive assessment updates via email
            </div>
          </FormControl>
        </DialogContent>
        <DialogActions className="email-dialog-actions">
          <Button 
            onClick={handleEmailDialogClose}
            className="cancel-button"
          >
            Cancel
          </Button>
          <Button 
            onClick={handleEmailUpdateSubmit} 
            className="submit-button"
            disabled={isCustom && (!!customDaysError || !customDays)}
          >
            Set Up Updates
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
};

export default AuthorInfoCard;
