import React from 'react';
import { TbReportSearch } from 'react-icons/tb';
import { reportService } from '../../../api/reports';
import '../../../styles/pages/_authorCard.scss';

const AuthorCard = ({ score, authors }) => {
  const handleClick = async (authorId, authorName, Source) => {
    const report = await reportService.createReport({
      AuthorId: authorId,
      AuthorName: authorName,
      Source: Source,
      StartYear: 1990,
    });

    window.open('/report/' + report.Id, '_blank');
  };

  return (
    <>
      <div className="selection-container">
        <div className="score-box-container">
          <div className="score-pill">
            <span className="score-label">Score:</span>
            <span className="score-value">{score}</span>
          </div>
        </div>
        <div className="search-container">
          <div className="search-icon">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path
                fillRule="evenodd"
                d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
                clipRule="evenodd"
              />
            </svg>
          </div>
          <input type="text" placeholder="Search..." className="search-input" />
        </div>

        <div className="controls-container">
          <button className="filters-button">
            Filters
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path
                fillRule="evenodd"
                d="M3 3a1 1 0 011-1h12a1 1 0 011 1v3a1 1 0 01-.293.707L12 11.414V15a1 1 0 01-.293.707l-2 2A1 1 0 018 17v-5.586L3.293 6.707A1 1 0 013 6V3z"
                clipRule="evenodd"
              />
            </svg>
          </button>
        </div>
      </div>
      <div className="author-card">
        <table className="author-table">
          <thead>
            <tr>
              <th>Author Name</th>
              <th>Flag Count</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {authors.map((author, index) => (
              <tr key={index}>
                <td>{author.AuthorName}</td>
                <td>{author.FlagCount}</td>
                <td>
                  <button
                    className="view-report-button"
                    onClick={() => handleClick(author.AuthorId, author.AuthorName, author.Source)}
                  >
                    <TbReportSearch />
                    <span>View Report</span>
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
};

export default AuthorCard;
