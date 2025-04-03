// src/TodoListComponent.js
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { reportService } from '../../../api/reports';
import '../../../styles/components/_todoListComponent.scss';
import Scholar from '../../../assets/icons/Scholar.svg';
import University from '../../../assets/icons/University.svg';
import Research from '../../../assets/icons/Research.svg';

import NoResultsFound from '../../common/tools/NoResultsFound';

const TodoListComponent = ({ results, setResults, canLoadMore, loadMore, isLoadingMore }) => {
  const navigate = useNavigate();

  const handleItemClick = async (result) => {
    const response = await reportService.createReport({
      AuthorId: result.AuthorId,
      AuthorName: result.AuthorName,
      Source: result.Source,
      StartYear: 1990,
    });
    navigate(`/report/${response.Id}`, {
      state: {
        canGoBack: true,
      },
    });
    return;
  };

  const getMoreResults = async () => {
    if (canLoadMore) {
      setResults(results.concat(await loadMore()));
    }
  };
  return (
    <div className="d-flex flex-column w-100 ">
      {results.length === 0 ? (
        <NoResultsFound />
      ) : (
        <>
          <ul className="result-wrapper mt-3">
            {results.map((result, index) => (
              <li key={index} onClick={() => handleItemClick(result)} className="result-item">
                <div className="text-start px-5">
                  <div className="info-row">
                    <img src={Scholar} alt="Scholar" className="icon scholar" />
                    <h5 className="title">{result.AuthorName}</h5>
                  </div>

                  <div className="info-row">
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
              </li>
            ))}
          </ul>
          {canLoadMore && (
            <div className="show-more-results-button">
              <button
                className="button button-3d"
                onClick={getMoreResults}
                disabled={isLoadingMore}
              >
                {isLoadingMore ? <div className="spinner"></div> : 'Show More'}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default TodoListComponent;
