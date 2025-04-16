import React from 'react';
import NoResultsFoundIcon from '../../../assets/icons/Not_Found_llustration.svg';
const NoResultsFound = (text) => {
  return (
    <div className="no-results">
      <div className="no-results-icon">
        <img
          src={NoResultsFoundIcon}
          alt="No Results Found"
          style={{ width: '120px', height: 'auto' }}
        />
      </div>
      <h6 style={{ fontWeight: 600 }}>{text.msg || 'Result Not Found'}</h6>
      <p>{text.submsg || 'Please refine your search criteria.'}</p>
    </div>
  );
};

export default NoResultsFound;
