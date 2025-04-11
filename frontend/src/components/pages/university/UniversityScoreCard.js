import { useEffect, useState } from 'react';
import VisualEmpty from '../../../assets/images/visualEmpty.svg';
import '../../../styles/components/_univScoreCard.scss';
import { Divider } from '@mui/material';

const ScoreCard = ({ reserachersAccessed, totalResearcher, loading }) => {
  const backgroundImage = VisualEmpty;
  const [progress, setProgress] = useState(0);
  useEffect(() => {
    const targetProgress = (reserachersAccessed / totalResearcher) * 100;

    setProgress(0);
    const timer = setTimeout(() => {
      setProgress(targetProgress);
    }, 50);

    return () => clearTimeout(timer);
  }, [reserachersAccessed, totalResearcher]);

  return (
    <div className="univ-score-card" style={{ backgroundImage: `url(${backgroundImage})` }}>
      <div
        className={`progress-bar ${progress < 100 ? 'flowing' : ''}`}
        style={{ width: `${progress}%` }}
        aria-valuenow={progress}
        aria-valuemin="0"
        aria-valuemax="100"
      />
      <Divider
        sx={{
          backgroundColor: 'black',
          height: '1px',
          width: '100%',
          opacity: 0.1,
          position: 'relative',
          zIndex: 2,
        }}
      />
      <div className="univ-score-card-content">
        <span className="univ-score-card-title">
          <span className="univ-score-card-title-score-accessed">{reserachersAccessed}</span>
          <span>{' Researchers Accessed'}</span>
        </span>
        <span className="univ-score-card-title">
          <span className="univ-score-card-title-score">{totalResearcher}</span>
          <span>{'Total Researchers'}</span>
        </span>
      </div>
    </div>
  );
};

export default ScoreCard;
