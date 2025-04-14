import React from 'react';

const Tooltip = ({ text }) => {
  return (
    <div className="popover-container">
      <button
        type="button"
        className="btn btn-info btn-circle ml-2"
        style={{
          marginLeft: '5px',
          width: '16px',
          height: '16px',
          padding: '1px 0',
          borderRadius: '7.5px',
          textAlign: 'center',
          fontSize: '8px',
          lineHeight: '1.42857',
          border: '1px solid grey',
          borderWidth: '1px',
          backgroundColor: 'transparent',
          color: 'grey',
          position: 'relative',
          boxShadow: 'none',
        }}
      >
        ?
      </button>
      <div className="popover">
        <div className="popover-body" style={{ whiteSpace: 'pre-wrap' }}>
          {text}
        </div>
      </div>
    </div>
  );
};

export default Tooltip;
