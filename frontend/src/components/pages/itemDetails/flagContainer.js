import React from 'react';
import '../../../styles/components/_flagContainer.scss';
import { FaRegCheckCircle } from 'react-icons/fa';

export default function FlagContainer({ isDisclosureChecked, isDisclosed, children }) {
  return (
    <div className="flag-container">
      <div className="flag-container-box">{children}</div>
      {isDisclosureChecked && (
        <div className={`flag-container-badge ${isDisclosed ? 'disclosed' : 'undisclosed'}`}>
          <span className="flag-container-badge-text">
            {isDisclosed ? 'Disclosed' : 'Undisclosed'}
          </span>
          {isDisclosed ? (
            <span className="flag-container-badge-check-circle">
              <FaRegCheckCircle />
            </span>
          ) : null}
        </div>
      )}
    </div>
  );
}
