import VisualEmpty from '../../../assets/images/visualEmpty.svg';
import VisualColored from '../../../assets/images/VisualColored.svg';
import '../../../styles/components/_scoreCard.scss';
import '../../../styles/components/_primaryButton.scss';

const ScoreCard = ({ score = 0 }) => {
    const backgroundImage = score === 0 ? VisualEmpty : VisualColored;

    return (
        <div
            className="score-card"
            style={{ backgroundImage: `url(${backgroundImage})` }}
        >
            <div className="score-card-content">
                <div className="score-card-content-left">
                    <span className='score-card-title'>
                        Your Assessment Insights in One View
                    </span>
                </div>
                <div className="score-card-content-right">
                    <span className={`score-card-score ${score > 0 ? 'has-score' : ''}`}>
                        {score}
                    </span>
                    <button
                        className="button button-3d"
                        style={{ padding: '8px 48px' }}
                        disabled={!score}
                    >
                        Visualise
                    </button>
                </div>
            </div>
        </div>
    );
};

export default ScoreCard;