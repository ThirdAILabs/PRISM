import React, { useState } from 'react';

const Collapsible = ({ title, children, initiallyExpanded = false }) => {
  const [isExpanded, setIsExpanded] = useState(initiallyExpanded);

  const toggleExpand = () => {
    setIsExpanded(!isExpanded);
  };

  return (
    <div className="collapsible-container mb-3">
      <div className="d-flex">
        <button
          type="button"
          className="btn btn-outline-dark w-25 d-flex justify-content-between align-items-center mb-2"
          onClick={toggleExpand}
          aria-expanded={isExpanded}
        >
          <span className="mx-auto">{title}</span>
          <span>
            <i className={`fa fa-chevron-${isExpanded ? 'up' : 'down'}`}></i>
          </span>
        </button>
      </div>

      {isExpanded && <div className="collapsible-content mt-2 w-100">{children}</div>}
    </div>
  );
};

export default Collapsible;
