// src/TodoListComponent.js
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { reportService } from '../../../api/reports';
import './TodoListComponent.css';

const TodoListComponent = ({ results, setResults, canLoadMore, loadMore }) => {
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
        canGoBack: true
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
      <>
        <ul className="result-wrapper">
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
          <div className="show-more-results-button">
            <button className="button" onClick={getMoreResults}>
              Show More
            </button>
          </div>
        )}
      </>
    </div>
  );
};

export default TodoListComponent;
