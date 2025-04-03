// import '../../../styles/components/_authorInfoCard.scss';
import Scholar from '../../../assets/icons/Scholar.svg';
import University from '../../../assets/icons/University.svg';
import Research from '../../../assets/icons/Research.svg';

const AuthorInfoCard = ({ result }) => {
    console.log("Result in AuthorInfoCard", result);

    return (
        <div className="text-start" style={{ padding: '20px 60px 10px 30px', width: "100%" }}>
            <div className="info-row">
                <img src={Scholar} alt="Scholar" className="icon scholar" />
                <h5 className="title">{result.AuthorName}</h5>
            </div>

            <div className="info-row" style={{ marginTop: '10px' }}>
                <img src={University} alt="Affiliation" className="icon" />
                <span className="content">
                    <span className="content-research">{result.Institutions[0]}</span>
                    {result.Institutions.length > 1 &&
                        ', ' + result.Institutions.slice(1).join(', ')}
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
        </div>
    )
}

export default AuthorInfoCard;