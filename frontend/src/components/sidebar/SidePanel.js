import { useEffect, useState } from 'react';
import { reportService } from '../../api/reports';
import { universityReportService } from '../../api/universityReports';
import './Sidepanel.css';
import RandomAvatar from '../../assets/images/RandomAvatar.jpg';
import PRISM_LOGO from '../../assets/images/prism.png';
import { FaRegUserCircle, FaUniversity, FaSearch } from 'react-icons/fa';
import UserService from '../../services/userService';
import { CiLogout } from 'react-icons/ci';
import { TbReportSearch } from 'react-icons/tb';
import { CiCircleList, CiCircleCheck } from 'react-icons/ci';
import { CgSpinner } from 'react-icons/cg';
import { useNavigate } from 'react-router-dom';
import { useUser } from '../../store/userContext';
import { MdDelete } from 'react-icons/md';
import { useLocation } from 'react-router-dom';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import ArrowRightIcon from '@mui/icons-material/ArrowRight';

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
      }, 1000);

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
        canGoBack: true
      },
    });
  };
  const handleUniversityReportClick = (universityReport) => {
    onClose();
    navigate(`/university/report/${universityReport.Id}`, {
      state: {
        canGoBack: true
      },
    });
  };
  return (
    <>
      {/* {!isOpen && (
        <div className="mini-nav" style={{
          position: 'fixed',
          left: 0,
          top: '80px',
          width: '60px',
          backgroundColor: 'white',
          boxShadow: '2px 0 5px rgba(0,0,0,0.1)',
          zIndex: 100,
          height: '100%'
        }}>
          <nav className="navigation">
            <ul className="nav-list">
              <li
                className={`nav-item ${currentLocation === '/' ? 'active' : ''}`}
                onClick={handleIndividualClick}
                title="Individual Assessment"
              >
                <span className="nav-icon">
                  <FaRegUserCircle size={24} />
                </span>
              </li>
              <li
                className={`nav-item ${currentLocation === '/university' ? 'active' : ''}`}
                onClick={handleUniversityClick}
                title="University Assessment"
              >
                <span className="nav-icon">
                  <FaUniversity size={24} />
                </span>
              </li>
              <li
                className={`nav-item ${currentLocation === '/entity-lookup' ? 'active' : ''}`}
                onClick={handleEntityClick}
                title="Entity Lookup"
              >
                <span className="nav-icon">
                  <FaSearch size={24} />
                </span>
              </li>
            </ul>
          </nav>
        </div>
      )} */}

      <div className={`side-panel ${isOpen ? 'open' : ''}`}>
        <div className="panel-content">
          {/* Header */}
          <div
            className="side-panel-header"
            style={{
              display: 'flex',
              alignItems: 'center',
              padding: '10px 20px',
              paddingTop: '0',
              borderBottom: '1px solid #e0e0e0',
              gap: '20px',
            }}
          >
            <img
              src={PRISM_LOGO}
              alt="PRISM"
              style={{ width: '150px', height: '30px', marginLeft: '10%' }}
            />
          </div>
          {/* Navigation */}
          <nav className="navigation">
            <ul className="nav-list">
              <li
                className={`nav-item ${currentLocation === '/' ? 'active' : ''}`}
                onClick={handleIndividualClick}
              >
                <span className="nav-icon">
                  <FaRegUserCircle />
                </span>
                <span className="nav-text">Individual Assessment</span>
              </li>
              <li
                className={`nav-item ${currentLocation === '/university' ? 'active' : ''}`}
                onClick={handleUniversityClick}
              >
                <span className="nav-icon">
                  <FaUniversity />
                </span>
                <span className="nav-text">University Assessment</span>
              </li>
              <li
                className={`nav-item ${currentLocation === '/entity-lookup' ? 'active' : ''}`}
                onClick={handleEntityClick}
              >
                <span className="nav-icon">
                  <FaSearch />
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
            <div className="collapsible-header">
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <div
                  onClick={() => setShowAuthorReports(!showAuthorReports)}
                  className="collapsible-icon"
                >
                  {showAuthorReports ? <ArrowDropDownIcon /> : <ArrowRightIcon />}
                </div>
                <span style={{ fontSize: 'medium', fontWeight: 'normal', marginLeft: '10px' }}>
                  Author Report
                </span>
              </div>
            </div>
            {showAuthorReports && (
              <div className="collapsible-content">
                {reports.map(
                  (report, index) =>
                    index < 10 && (
                      <div
                        key={report.Id}
                        className="report-item"
                        onClick={handleReportClick.bind(null, report)}
                      >
                        <span>{report.AuthorName}</span>
                        {/* <span><MdDelete style={15} /></span> */}
                        <span className={`status ${report.Status}`}>{status[report.Status]}</span>
                      </div>
                    )
                )}
              </div>
            )}
            <div
              className="collapsible-header"
            // style={{ marginTop: '10px' }}
            >
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <div
                  onClick={() => setShowUniversityReports(!showUniversityReports)}
                  className="collapsible-icon"
                >
                  {showUniversityReports ? <ArrowDropDownIcon /> : <ArrowRightIcon />}
                </div>
                <span style={{ fontSize: 'medium', marginLeft: '10px' }}>University Report</span>
              </div>
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
                        <span>{universityReport.UniversityName}</span>
                        {/* <span><MdDelete style={15} /></span> */}
                        <span className={`status ${universityReport.Status}`}>
                          {status[universityReport.Status]}
                        </span>
                      </div>
                    )
                )}
              </div>
            )}
          </div>
        </div>

        {/* User Info */}
        <div className="user-info">
          <img
            src={user.avatar}
            alt="User"
            style={{ width: '40px', height: '40px', borderRadius: '100%' }}
          />
          <div>
            <h5>{user.username}</h5>
          </div>
        </div>

        {/* Logout */}
        <div className="logout-section">
          <button
            className="btn btn-dark w-100 d-flex align-items-center justify-content-center gap-2 border"
            onClick={UserService.doLogout}
          >
            <CiLogout /> Logout
          </button>
        </div>
      </div>
    </>
  );
};

export default SidePanel;
