import React from 'react';
import { TbReportSearch } from 'react-icons/tb';
import { reportService } from '../../../api/reports';
import './AuthorCard.scss';

const AuthorCard = ({ authors }) => {
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
  );
};

export default AuthorCard;
