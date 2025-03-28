import React from 'react';
import '../../../styles/components/_noResultFound.scss';

const NoResultsFound = (text) => {
  console.log("text.msg",text.msg);
  return (
    <div className="no-results">
      <div className="no-results-icon">🔍</div>
      <h3 className='no-results-heading'>{text.msg || 'No Result Found'}</h3>
      <p className='no-results-text'>{text.submsg || 'Please refine your search criteria and try again.'}</p>
    </div>
  );
};

export default NoResultsFound;
