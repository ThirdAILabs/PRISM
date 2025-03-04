import { useEffect, useState } from 'react';
import { reportService } from '../../api/reports';
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

const SidePanel = ({ isOpen, onClose }) => {
  const { userInfo } = useUser();
  const [reports, setReports] = useState([]);
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
    fetchReports();
  }, []);

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
    navigate(`/report/${report.Id}`);
  };

  const handleMouseLeave = () => {
    onClose();
  };

  return (
    <div 
      className={`side-panel ${isOpen ? 'open' : ''}`}
      onMouseLeave={handleMouseLeave}
    >
      <div className="panel-content">
        {/* Header */}
        <div className="side-panel-header">
          <img
            src={PRISM_LOGO}
            alt="PRISM"
            style={{ 
              height: '30px', 
              width: 'auto',
              objectFit: 'contain'
            }}
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
          {reports.map(
            (report, index) =>
              index < 10 && (
                <div
                  key={report.Id}
                  className="report-item"
                  onClick={handleReportClick.bind(null, report)}
                >
                  <p>{report.AuthorName}</p>
                  {/* <span><MdDelete style={15} /></span> */}
                  <span className={`status ${report.Status}`}>{status[report.Status]}</span>
                </div>
              )
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
  );
};

export default SidePanel;
