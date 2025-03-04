import React from 'react';
import './Shimmer.css';

const ShimmerCard = () => {
  return (
    <>
      <div className="shimmer-card-header"></div>
      <div style={{ display: 'flex', flexWrap: 'wrap', marginLeft: '5px', marginTop: '5rem' }}>
        {Array.from({ length: 7 }).map((_, idx) => (
          <div className="shimmer-card" key={idx}>
            <div className="shimmer-card-content shimmer"></div>
            <div className="shimmer-card-buttons">
              <div className="shimmer-button shimmer"></div>
            </div>
          </div>
        ))}
      </div>
    </>
  );
};

export default ShimmerCard;
