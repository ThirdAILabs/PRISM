import { useEffect, useState } from 'react';
import { reportService } from '../../api/reports';
import './Sidepanel.css';
import RandomAvatar from '../../assets/images/RandomAvatar.jpg'
// import PRISM_LOGO from '../../assets/icons/prism-logo.svg';
import PRISM_LOGO from '../../assets/images/prism.png';
import { FaBars, FaRegUserCircle, FaUniversity, FaSearch } from 'react-icons/fa';
import UserService from '../../services/userService';
import { CiLogout } from "react-icons/ci";
import { TbReportSearch } from "react-icons/tb";
import { CiCircleList, CiCircleCheck } from "react-icons/ci";
import { CgSpinner } from "react-icons/cg";
import { useNavigate } from 'react-router-dom';


const SidePanel = ({ isOpen, onClose }) => {
    const [reports, setReports] = useState([]);
    const navigate = useNavigate();
    const user = {
        avatar: RandomAvatar,
        name: 'Anand Kumar',
        // email: 'admin@mail.com'
    };

    const status = {
        'queued': <CiCircleList title="Queued" size={20} />,
        'in-progress': <CgSpinner className="animate-spin" title="In-Progress" size={20} />,
        'complete': <CiCircleCheck title="Completed" size={20} />
    };

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
    }
    const handleIndividualClick = () => {
        onClose();
        navigate('/');
    }
    const handleUniversityClick = () => {
        onClose();
        navigate('/university-assessment');
    }
    const handleReportClick = (report) => {
        onClose();
        navigate(`/report/${report.Id
            }`);
    };
    return (
        <div className={`side-panel ${isOpen ? 'open' : ''}`}>
            <div className="panel-content">
                <div style={{ display: 'flex', alignItems: 'center', padding: '10px 20px', paddingTop: '0', borderBottom: '1px solid #e0e0e0', gap: '20px' }}>
                    <img src={PRISM_LOGO} alt="PRISM" style={{ width: "150px", height: '30px', marginLeft: '10%' }} />
                </div>


                <nav className="navigation">
                    <ul className="nav-list">
                        <li className="nav-item" onClick={handleIndividualClick}>
                            <span className="nav-icon"><FaRegUserCircle /></span>
                            <span className="nav-text">Individual Assessment</span>
                        </li>
                        <li className="nav-item" onClick={handleUniversityClick}>
                            <span className="nav-icon"><FaUniversity /></span>
                            <span className="nav-text">University Assessment</span>
                        </li>
                        <li className="nav-item" onClick={handleEntityClick}>
                            <span className="nav-icon"><FaSearch /></span>
                            <span className="nav-text">Entity Lookup</span>
                        </li>
                    </ul>
                </nav>

                <div className="reports">
                    <h5 className='report-header'><TbReportSearch /> Reports</h5>
                    {reports.map(report => (
                        <div key={report.Id} className="report-item" onClick={handleReportClick.bind(null, report)}>
                            <p>{report.AuthorName}</p>
                            <span className={`status ${report.Status}`}>{status[report.Status]}</span>
                        </div>
                    ))}
                </div>

            </div>

            {/* User Info */}
            <div className="user-info">
                <img src={user.avatar} alt="User" style={{ width: "40px", height: '40px', borderRadius: '100%' }} />
                <div>
                    <h5>{user.name}</h5>
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