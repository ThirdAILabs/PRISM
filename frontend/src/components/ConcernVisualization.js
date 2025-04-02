import * as React from 'react';
import { Speedometer } from './common/tools/Speedometer';
import '../App.css';
import '../styles/components/_primaryButton.scss';
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

export const BaseFontSize = 48;
export function getFontSize(value) {
  const digits = String(value).length;

  // Each additional digit reduces size by 4px
  const newSize = BaseFontSize - (digits - 1) * 4;
  return Math.max(newSize, 28);
}

export default function ConcernVisualizer({
  title,
  hoverText,
  value,
  scale,
  onReview,
  selected,
  speedometerHoverText,
  valueFontSize,
}) {
  //200,344
  return (
    <div style={{ width: '200px', height: '356px', position: 'relative' }}>
      <Speedometer
        scale={scale || [0, 10, 20, 30, 50, 100, 200]}
        value={value}
        speedometerHoverText={speedometerHoverText}
        valueFontSize={valueFontSize}
      />

      <div className="mt-3 text-dark" style={{ height: '50px', paddingInline: '8px', marginBottom: '34px' }}>
        {title}
        <Hover text={hoverText} />
      </div>
      {onReview && (
        <div style={{ marginLeft: '2rem' }}>
          <button
            className={`button button-3d`}
            style={{
              padding: '8px 48px',
            }}
            disabled={!value}
            onClick={onReview}
          >
            View
          </button>
        </div>
      )}
    </div>
  );
}
