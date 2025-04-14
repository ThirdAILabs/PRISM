import React from 'react';
import { Check } from 'lucide-react';
import '../../../styles/components/_flagContainer.scss';

export default function FlagContainer({ isDisclosureChecked, isDisclosed, children }) {
  return (
    <div className="flag-container">
      <div className="flag-container-box">{children}</div>
      {/* This tag looks to be redudant. We can probably revisit this at a later time, so I'll leave the code here. -Ashutosh */}
      {/* {isDisclosureChecked && (
        <div className={`flag-container-badge ${isDisclosed ? 'disclosed' : 'undisclosed'}`}>
          <span className="flag-container-badge-text">
            {isDisclosed ? 'Disclosed' : 'Undisclosed'}
          </span>
          {isDisclosed ? (
            <span className="flag-container-badge-check-circle">
              <Check color="white" strokeWidth={3} />
            </span>
          ) : null}
        </div>
      )} */}
    </div>
  );
}
