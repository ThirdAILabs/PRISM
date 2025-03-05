import * as React from 'react';
import { Speedometer } from './common/tools/Speedometer';
import '../App.css';

const Hover = ({ text }) => {
  return (
    <div className="popover-container">
      <button
        type="button"
        className="btn btn-info btn-circle ml-2"
        style={{
          marginLeft: '5px',
          width: '14px',
          height: '14px',
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

export default function ConcernVisualizer({ title, hoverText, value, scale, onReview, selected }) {
  return (
    <div style={{ width: '200px', height: '300px', position: 'relative' }}>
      <Speedometer scale={scale || [0, 1, 2, 3, 5, 10, 20]} value={value} />

      <div className="mt-3 mb-4 text-dark" style={{ height: '50px' }}>
        {title} <Hover text={hoverText} />
      </div>
      {onReview && (
        <button
          className={`btn btn-dark rounded rounded-5 px-4`}
          style={{
            background: !value ? 'ghostwhite' : 'lightgrey',
            border: 'none',
            color: 'black',
            boxShadow: selected ? '0 0px 12px rgb(57, 57, 57)' : 'none',
          }}
          disabled={!value}
          onClick={onReview}
        >
          Review
        </button>
      )}
    </div>
  );
}
