import VisualEmpty from '../../../assets/images/visualEmpty.svg';
import VisualColored from '../../../assets/images/VisualColored.svg';
import '../../../styles/components/_scoreCard.scss';
import '../../../styles/components/_primaryButton.scss';

const ScoreCard = ({ score = 0, setActiveTab, loading }) => {
  const backgroundImage = score === 0 ? VisualEmpty : VisualColored;

  return (
    <div className="score-card" style={{ backgroundImage: `url(${backgroundImage})` }}>
      <div className="score-card-content">
        <div className="score-card-content-left">
          <span className="score-card-title">
            {'Your Assessment '}
            {loading && 'is in '}
            <span className={loading ? 'progress-text' : ''}>
              {loading ? 'Progress...' : 'in One View'}
              {loading && (
                <div
                  className="spinner-border text-primary"
                  style={{ width: '2rem', height: '2rem' }}
                  role="status"
                ></div>
              )}
            </span>
          </span>
        </div>
        <div className="score-card-content-right">
          <span className={`score-card-score ${score > 0 ? 'has-score' : ''}`}>{score}</span>
          <button
            className="button button-3d score-card-button"
            style={{ padding: '8px 48px' }}
            disabled={loading || !score}
            onClick={() => setActiveTab(1)}
          >
            Visualise
          </button>
        </div>
      </div>
    </div>
  );
};

export default ScoreCard;
