import React from 'react';
import { Check } from 'lucide-react';
import '../../../styles/components/_flagPanel.scss';

export default function FlagContainer({ showDisclosure, isDisclosed = false, children }) {
  return (
    <div className="flag-container">
      <div className="flag-container-box">{children}</div>

      <div className={`flag-container-badge ${isDisclosed ? 'disclosed' : 'undisclosed'}`}>
        <span className="flag-container-badge-text">
          {isDisclosed ? 'Disclosed' : 'Undisclosed'}
        </span>
        <div className="flag-container-badge-check-circle">
          <Check color="white" strokeWidth={3} />
        </div>
      </div>
    </div>
  );
}

export function FlagSubContainer({ children }) {
  return <div className="flag-sub-container">{children}</div>;
}
