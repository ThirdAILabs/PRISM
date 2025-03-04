import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { reportService } from '../api/reports';

const TodoListComponent = ({ results, setResults, canLoadMore, loadMore, noResultFound }) => {
  const navigate = useNavigate();
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  const handleItemClick = async (result) => {
    const response = await reportService.createReport({
      AuthorId: result.AuthorId,
      AuthorName: result.AuthorName,
      Source: result.Source,
      StartYear: 1990,
    });
    navigate(`/report/${response.Id}`);
    return;
  };

  const getMoreResults = async () => {
    if (canLoadMore) {
      setIsLoadingMore(true);
      try {
        const moreResults = await loadMore();
        setResults(results.concat(moreResults));
      } catch (error) {
        console.error("Error loading more results:", error);
      } finally {
        setIsLoadingMore(false);
      }
    }
  };

  return (
    <div className="d-flex flex-column align-items-center w-100 ">
      <>
        {results.length > 0 ? (
          <>
            <ul className="result-wrapper d-flex flex-wrap mt-3">
              {results.map((result, index) => (
                <li key={index} onClick={() => handleItemClick(result)} className="result-item">
                  <div className="text-start px-5">
                    <div className="d-flex align-items-center mb-2">
                      <h5 className="m-0">{result.AuthorName}</h5>
                    </div>
                    <p className="m-0 p-0" style={{ fontSize: 'small' }}>
                      <b>Affiliations: </b>
                      {result.Institutions.join(', ')}
                    </p>
                    {result.Interests && result.Interests.length > 0 && (
                      <div>
                        <p className="m-0 p-0 pt-1" style={{ fontSize: 'small' }}>
                          <b>Research Interests: </b>
                          {result.Interests.slice(0, 3).join(', ')}
                        </p>
                      </div>
                    )}
                  </div>
                </li>
              ))}
            </ul>
            {canLoadMore && (
              <div className="d-flex flex-column align-items-center pb-4">
                <button 
                  className="button" 
                  onClick={getMoreResults} 
                  disabled={isLoadingMore}
                >
                  {isLoadingMore ? "Loading..." : "Show More"}
                </button>
                {isLoadingMore && (
                  <div className="mt-2">
                    <div className="spinner-border text-primary" role="status">
                      <span className="visually-hidden">Loading...</span>
                    </div>
                  </div>
                )}
              </div>
            )}
          </>
        ) : (
          <div className="d-flex justify-content-center mt-4" style={{ textAlign: 'center', padding: '2rem', color: '#4a5568', fontSize: '1.1rem', width: '100%' }}>
            <p>{noResultFound ? "No results found" : "Loading..."}</p>
          </div>
        )}
      </>
    </div>
  );
};

export default TodoListComponent;