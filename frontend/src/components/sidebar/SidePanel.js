import { useEffect, useState } from 'react';
import { reportService } from '../../api/reports';
import { universityReportService } from '../../api/universityReports';
import RandomAvatar from '../../assets/images/RandomAvatar.jpg';
import PRISM_LOGO from '../../assets/images/prism.png';
import UserService from '../../services/userService';
import { FiLogOut } from "react-icons/fi";
import { TbReportSearch } from 'react-icons/tb';
import { CiCircleList, CiCircleCheck } from 'react-icons/ci';
import { CgSpinner } from 'react-icons/cg';
import { useNavigate } from 'react-router-dom';
import { useUser } from '../../store/userContext';
import { useLocation } from 'react-router-dom';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import ArrowRightIcon from '@mui/icons-material/ArrowRight';
import { Tooltip } from '@mui/material';
import { GRAPHICS } from '../../assets/icons/graphics';
import '../../styles/components/_sidepanel.scss';


const SidePanel = ({ isOpen, onClose }) => {
  const { userInfo } = useUser();
  const [reports, setReports] = useState([]);
  const [universityReports, setUniversityReports] = useState([]);
  const [showAuthorReports, setShowAuthorReports] = useState(false);
  const [showUniversityReports, setShowUniversityReports] = useState(false);

  const navigate = useNavigate();
  const user = {
    avatar: RandomAvatar,
    name: userInfo.name,
    email: userInfo.email,
    username: userInfo.username,
  };

  const status = {
    queued: <CiCircleList title="Queued" size={16} />,
    'in-progress': <CgSpinner className="animate-spin" title="In-Progress" size={16} />,
    complete: <CiCircleCheck title="Completed" size={16} />,
  };
  const location = useLocation();
  const currentLocation = location.pathname;

  useEffect(() => {
    const fetchReports = async () => {
      const data = await reportService.listReports();
      setReports(data);
    };
    const fetchUniversityReports = async () => {
      const data = await universityReportService.listReports();
      setUniversityReports(data);
    };
    if (isOpen) {
      fetchReports();
      fetchUniversityReports();

      const pollingInterval = setInterval(() => {
        fetchReports();
        fetchUniversityReports();
      }, 10000);

      return () => clearInterval(pollingInterval);
    }
  }, [isOpen]);

  const handleEntityClick = () => {
    onClose();
    navigate('/entity-lookup');
  };
  const handleIndividualClick = () => {
    onClose();
    navigate('/');
  };
  const handleUniversityClick = () => {
    onClose();
    navigate('/university');
  };
  const handleReportClick = (report) => {
    onClose();
    navigate(`/report/${report.Id}`, {
      state: {
        canGoBack: true,
      },
    });
  };
  const handleUniversityReportClick = (universityReport) => {
    onClose();
    navigate(`/university/report/${universityReport.Id}`, {
      state: {
        canGoBack: true,
      },
    });
  };
  const maximumAllowedStringLength = 25;
  const truncateString = (str) => {
    if (str.length > maximumAllowedStringLength)
      return str.substring(0, maximumAllowedStringLength) + '...';
    return str;
  };

  return (
    <>
      <div className={`side-panel ${isOpen ? 'open' : ''}`}>
        <div className="panel-content">
          {/* Header */}
          <div className="side-panel-header">
            <img src={PRISM_LOGO} alt="PRISM" className="side-panel-header__logo" />
          </div>
          {/* Navigation */}
          <nav className="navigation">
            <ul className="nav-list">
              <li
                className={`nav-item ${currentLocation === '/' ? 'active' : ''}`}
                onClick={handleIndividualClick}
              >
                <span className="nav-icon">
                  { GRAPHICS.individual_assessment }
                </span>
                <span className="nav-text">Individual Assessment</span>
              </li>
              <li
                className={`nav-item ${currentLocation === '/university' ? 'active' : ''}`}
                onClick={handleUniversityClick}
              >
                <span className="nav-icon">
                  { GRAPHICS.university }
                </span>
                <span className="nav-text">University Assessment</span>
              </li>
              <li
                className={`nav-item ${currentLocation === '/entity-lookup' ? 'active' : ''}`}
                onClick={handleEntityClick}
              >
                <span className="nav-icon">
                  {GRAPHICS.entity_lookup}
                </span>
                <span className="nav-text">Entity Lookup</span>
              </li>
            </ul>
          </nav>

          {/* Reports */}
          <div className="reports">
            <h5 className="report-header">
              <TbReportSearch /> Reports
            </h5>
            {reports?.length ? (
              <div className="collapsible-header">
                <div className="collapsible-subheader">
                  <div
                    onClick={() => setShowAuthorReports(!showAuthorReports)}
                    className="collapsible-icon"
                  >
                    {showAuthorReports ? <ArrowDropDownIcon /> : <ArrowRightIcon />}
                    <span className="collapsible-icon-text">Author Report</span>
                  </div>
                </div>
              </div>
            ) : null}
            {showAuthorReports && (
              <div className="collapsible-content">
                {reports.map(
                  (report, index) =>
                    index < 1000 && (
                      <div
                        key={report.Id}
                        className="report-item"
                        onClick={handleReportClick.bind(null, report)}
                      >
                        {report?.AuthorName?.length > maximumAllowedStringLength ? (
                          <Tooltip
                            title={report.AuthorName}
                            placement="right"
                            arrow
                            componentsProps={{
                              tooltip: {
                                sx: {
                                  bgcolor: 'rgba(60,60,60, 0.87)',
                                  '& .MuiTooltip-arrow': {
                                    color: 'rgba(60, 60, 60, 0.87)',
                                  },
                                  padding: '8px 12px',
                                  fontSize: '14px',
                                },
                              },
                            }}
                          >
                            <span className="text-start">{truncateString(report.AuthorName)}</span>
                          </Tooltip>
                        ) : (
                          <span className="text-start">{truncateString(report.AuthorName)}</span>
                        )}
                        <span className={`status ${report.Status}`}>{status[report.Status]}</span>
                      </div>
                    )
                )}
              </div>
            )}
            <div className="collapsible-header">
              {universityReports?.length ? (
                <div className="collapsible-subheader">
                  <div
                    onClick={() => setShowUniversityReports(!showUniversityReports)}
                    className="collapsible-icon"
                  >
                    {showUniversityReports ? <ArrowDropDownIcon /> : <ArrowRightIcon />}
                    <span className="collapsible-icon-text">University Report</span>
                  </div>
                </div>
              ) : null}
            </div>
            {showUniversityReports && (
              <div className="collapsible-content">
                {universityReports.map(
                  (universityReport, index) =>
                    index < 1000 && (
                      <div
                        key={universityReport.Id}
                        className="report-item"
                        onClick={handleUniversityReportClick.bind(null, universityReport)}
                      >
                        {universityReport?.UniversityName?.length > maximumAllowedStringLength ? (
                          <Tooltip
                            title={universityReport.UniversityName}
                            placement="right"
                            arrow
                            componentsProps={{
                              tooltip: {
                                sx: {
                                  bgcolor: 'rgba(60,60,60, 0.87)',
                                  '& .MuiTooltip-arrow': {
                                    color: 'rgba(60, 60, 60, 0.87)',
                                  },
                                  padding: '8px 12px',
                                  fontSize: '14px',
                                },
                              },
                            }}
                          >
                            <span className="text-start">
                              {truncateString(universityReport.UniversityName)}
                            </span>
                          </Tooltip>
                        ) : (
                          <span className="text-start">
                            {truncateString(universityReport.UniversityName)}
                          </span>
                        )}
                        <span className={`status ${universityReport.Status}`}>
                          {universityReport.Status === 'complete' &&
                          universityReport.Content.TotalAuthors !==
                            universityReport.Content.AuthorsReviewed
                            ? status['in-progress']
                            : status[universityReport.Status]}
                        </span>
                      </div>
                    )
                )}
              </div>
            )}
          </div>
        </div>

        {/* User Info */}

        <div className="user-card">
          <div className="user-card__profile">
            <img src={user.avatar} alt="User" className="user-card__avatar" />
            <div className="user-card__info">
              <h5 className="user-card__name">{user.username}</h5>
              <span className="user-card__email">{user.email}</span>
            </div>
          </div>
          <hr className="user-card__divider" />

          <button className="user-card__logout-button" onClick={UserService.doLogout}>
          <span>Logout</span>
          <FiLogOut/>
          </button>
        </div>
      </div>
    </>
  );
};

export default SidePanel;
