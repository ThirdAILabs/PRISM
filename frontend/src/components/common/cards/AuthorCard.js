import React, { useState } from 'react';
import { TbReportSearch } from 'react-icons/tb';
import { reportService } from '../../../api/reports';
import '../../../styles/pages/_authorCard.scss';
import { CiSearch } from 'react-icons/ci';
import { IoFilter } from 'react-icons/io5';

const AuthorCard = ({ score, authors }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const filteredAuthors = authors.filter(author => 
    author.AuthorName.toLowerCase().includes(searchTerm.toLowerCase())
  );

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
            <CiSearch />
          </div>
          <input 
            type="text" 
            placeholder="Search..." 
            className="search-input"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>
      <div className="author-card">
        <table className="author-table">
          <thead>
            <tr>
              <th>Author Name</th>
              <th>Flag Count</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {filteredAuthors.map((author, index) => (
              <tr key={index}>
                <td>{author.AuthorName}</td>
                <td>{author.FlagCount}</td>
                <td>
                  <button
                    class="button button-3d view-report-button"
                    onClick={() => handleClick(author.AuthorId, author.AuthorName, author.Source)}
                  >
                    View Report
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
