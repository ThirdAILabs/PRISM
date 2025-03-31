import React from 'react';

const NoResultsFound = (text) => {
  return (
    <div className="no-results">
      <div className="no-results-icon">🔍</div>
      <h3>{text.msg || 'No Result Found'}</h3>
      <p>{text.submsg || 'Please refine your search criteria and try again.'}</p>
    </div>
  );
};

export default NoResultsFound;
