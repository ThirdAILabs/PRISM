import '../../../styles/components/_authorInfoCard.scss';
import Scholar from '../../../assets/icons/Scholar.svg';
import University from '../../../assets/icons/University.svg';

const UniversityInfoCard = ({ result }) => {
  return (
    <div
      className="text-start"
      style={{ padding: '20px 60px 4px 30px', width: '100%', height: '100px' }}
    >
      <div className="info-row">
        <img src={Scholar} alt="Scholar" className="icon scholar" />
        <h5 className="title">{result.name}</h5>
      </div>

      <div className="info-row" style={{ marginTop: '10px' }}>
        <img src={University} alt="Affiliation" className="icon" />
        <span className="content">6100 Main St, Houston, TX 77005, USA</span>
      </div>
    </div>
  );
};

export default UniversityInfoCard;
