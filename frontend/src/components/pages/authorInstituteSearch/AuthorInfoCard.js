import '../../../styles/components/_authorInfoCard.scss';
import Scholar from '../../../assets/icons/Scholar.svg';
import University from '../../../assets/icons/University.svg';
import Research from '../../../assets/icons/Research.svg';
import { Divider } from '@mui/material';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import CloudDownloadIcon from '@mui/icons-material/CloudDownload';
import { getTrailingWhiteSpace } from '../../../utils/helper';
import DownloadDropdown from '../../common/tools/button/downloadButton';
import { useEffect, useState } from 'react';
import useOutsideClick from '../../../hooks/useOutsideClick';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import EmailIcon from '@mui/icons-material/Email';
import MarkEmailReadIcon from '@mui/icons-material/MarkEmailRead';

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
  TextField,
} from '@mui/material';

import { reportService } from '../../../api/reports';

const AuthorInfoCard = ({
  result,
  verifyWithDisclosure,
  downloadProps,
  filterProps,
  loading,
  reportId,
}) => {
  const [downloadDropdownOpen, setDownloadDropdownOpen] = useState(false);
  const [filterDropdownOpen, setFilterDropdownOpen] = useState(false);
  const [isFilterActive, setIsFilterActive] = useState(false);
  const [emailUpdateDiaLogBox, setEmailUpdateDiaLogBox] = useState(false);
  const [emailUpdateHookId, setEmailUpdateHookId] = useState('');
  const [emailFrequency, setEmailFrequency] = useState('');
  const [customDays, setCustomDays] = useState('');
  const [customDaysError, setCustomDaysError] = useState('');
  const [isCustom, setIsCustom] = useState(false);
  const [hasExistingSubscription, setHasExistingSubscription] = useState(false);
  const [emailUpdateHookDisabled, setEmailUpdateHookDisabled] = useState(false);

  useEffect(() => {
    // check if the email update hook is disabled
    const checkEmailUpdateHookDisabled = async () => {
      try {
        const hooks = await reportService.listHooks();
        if (hooks.some((hook) => hook.Type === 'AuthorReportUpdateNotifier')) {
          setEmailUpdateHookDisabled(false);
          setHasExistingSubscription(false);
        } else {
          setEmailUpdateHookDisabled(true);
        }
      } catch (error) {
        console.error('Error checking email update hook:', error);
        setEmailUpdateHookDisabled(true);
        setHasExistingSubscription(false);
      }
    };
    checkEmailUpdateHookDisabled();

    if (!emailUpdateHookDisabled) {
      // fetch the email update hooks for the report
      const fetchHooks = async () => {
        const hooks = await reportService.getHooks(reportId);
        const authorReportEmailUpdateHook = hooks.find(
          (hook) => hook.Action === 'AuthorReportUpdateNotifier'
        );
        if (authorReportEmailUpdateHook) {
          setHasExistingSubscription(true);
          setEmailFrequency((authorReportEmailUpdateHook.Interval / (24 * 60 * 60)).toString());
          setEmailUpdateHookId(authorReportEmailUpdateHook.Id);
        }
      };
      fetchHooks();
    }
  }, []);

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
    if (!isCustom) {
      setCustomDays('');
      setCustomDaysError('');
    }
  };

  const handleEmailUpdateSubmit = async () => {
    try {
      const interval = parseInt(isCustom ? customDays : emailFrequency);
      if (!customDaysError) {
        const res = await reportService.createHook(reportId, {
          action: 'AuthorReportUpdateNotifier',
          interval: interval * 24 * 60 * 60, // Convert days to seconds
        });
        setEmailFrequency(interval.toString());
        setCustomDays('');
        setHasExistingSubscription(true);
        setEmailUpdateDiaLogBox(false);
        setEmailUpdateHookId(res.Id);
      }
    } catch (error) {
      console.error('Error creating email hook:', error);
    }
  };

  const handleUnsubscribe = async () => {
    try {
      await reportService.deleteHook(reportId, emailUpdateHookId);
      setHasExistingSubscription(false);
      setEmailUpdateDiaLogBox(false);
    } catch (error) {
      console.error('Error deleting email hook:', error);
    }
  };

  return (
    <div className="text-start card-container" style={{ width: '100%' }}>
      <div className="card-top" style={{ padding: '20px 20px 4px 30px', flexGrow: 1 }}>
        <div className="info-row">
          <div className="title-container">
            <img src={Scholar} alt="Scholar" className="icon scholar" />
            <h5 className="title">{result.AuthorName}</h5>
            <button
              className={`email-updates-button ${emailUpdateHookDisabled ? 'disabled' : ''}`}
              onClick={() => setEmailUpdateDiaLogBox(true)}
              title={
                emailUpdateHookDisabled ? 'Email Updates Unavailable' : 'Subscribe to Email Updates'
              }
            >
              {hasExistingSubscription ? <MarkEmailReadIcon /> : <EmailIcon />}
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
        onClose={() => setEmailUpdateDiaLogBox(false)}
        className="email-dialog"
      >
        <DialogTitle className="email-dialog-title">
          <span className="email-dialog-icon">
            {emailUpdateHookDisabled ? (
              <EmailIcon className="disabled" />
            ) : hasExistingSubscription ? (
              <MarkEmailReadIcon />
            ) : (
              <EmailIcon />
            )}
          </span>
          <span className="email-dialog-title-text">
            {emailUpdateHookDisabled
              ? 'Email Updates Unavailable'
              : hasExistingSubscription
                ? 'Email Updates Active'
                : 'Set Up Email Updates'}
          </span>
        </DialogTitle>
        <DialogContent className="email-dialog-content">
          {emailUpdateHookDisabled ? (
            <div className="disabled-message">
              Please contact your administrator to enable this feature.
            </div>
          ) : hasExistingSubscription ? (
            <div className="subscription-enabled-message">
              <span className="check-icon">✓</span>
              You are currently receiving email updates{' '}
              {emailFrequency === '15'
                ? 'bi-weekly'
                : emailFrequency === '30'
                  ? 'monthly'
                  : emailFrequency === '90'
                    ? 'quarterly'
                    : `every ${emailFrequency} days`}
            </div>
          ) : (
            <FormControl fullWidth className="email-frequency-select">
              <InputLabel>Email Frequency</InputLabel>
              <Select
                value={emailFrequency}
                label="Email Frequency"
                onChange={handleFrequencyChange}
              >
                <MenuItem value="15">Bi-weekly</MenuItem>
                <MenuItem value="30">Monthly</MenuItem>
                <MenuItem value="90">Quarterly</MenuItem>
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
          )}
        </DialogContent>
        <DialogActions className="email-dialog-actions">
          {emailUpdateHookDisabled ? (
            <Button onClick={() => setEmailUpdateDiaLogBox(false)} className="cancel-button">
              Close
            </Button>
          ) : hasExistingSubscription ? (
            <Button onClick={handleUnsubscribe} className="unsubscribe-button">
              Unsubscribe
            </Button>
          ) : (
            <>
              <Button onClick={() => setEmailUpdateDiaLogBox(false)} className="cancel-button">
                Cancel
              </Button>
              <Button
                onClick={handleEmailUpdateSubmit}
                className="submit-button"
                disabled={!emailFrequency || (isCustom && (!!customDaysError || !customDays))}
              >
                Subscribe
              </Button>
            </>
          )}
        </DialogActions>
      </Dialog>
    </div>
  );
};

export default AuthorInfoCard;
