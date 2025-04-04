// import '../../../styles/components/_authorInfoCard.scss';
import Scholar from '../../../assets/icons/Scholar.svg';
import University from '../../../assets/icons/University.svg';
import Research from '../../../assets/icons/Research.svg';
import { Divider, Button } from '@mui/material';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import CloudDownloadIcon from '@mui/icons-material/CloudDownload';
import { getTrailingWhiteSpace } from '../../../utils/helper';

const AuthorInfoCard = ({ result }) => {
  console.log('Result in AuthorInfoCard', result);

  return (
    <div className="text-start" style={{ padding: '20px 60px 4px 30px', width: '100%' }}>
      <div className="info-row">
        <img src={Scholar} alt="Scholar" className="icon scholar" />
        <h5 className="title">{result.AuthorName}</h5>
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
            width: 'calc(100% + 90px)', // Compensate for parent padding
          },
        }}
      />
      <div className="button-group">
        <span>
          {'Verify With Disclosures' + getTrailingWhiteSpace(2)} <UploadFileIcon color="info" />
        </span>
        <span>
          {'Download Report' + getTrailingWhiteSpace(4)}
          <CloudDownloadIcon color="info" />
        </span>
      </div>
    </div>
  );
};

export default AuthorInfoCard;
