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
  const newSize = BaseFontSize - (digits - 1) * 12;
  return 28;
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
  return (
    <div
      style={{
        width: '180px',
        height: '340px',
        position: 'relative',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
      }}
    >
      <div style={{ width: '100%', display: 'flex', justifyContent: 'center', marginTop: '-12px' }}>
        <Speedometer
          scale={scale || [0, 10, 20, 30, 50, 100, 200]}
          value={value}
          speedometerHoverText={speedometerHoverText}
          valueFontSize={valueFontSize}
        />
      </div>

      <div
        style={{
          height: '50px',
          padding: '8px',
          marginBottom: '48px',
          marginTop: '-16px',
          textAlign: 'center',
          width: '100%',
          fontSize: '14px',
          lineHeight: '20px',
          letterSpacing: '0.1px',
          fontWeight: 'bold',
          color: value !== 0 ? '#333f54' : '#6a798f',
          paddingInline: '16px',
        }}
      >
        {title}
      </div>

      {onReview && (
        <div
          style={{
            width: '100%',
            display: 'flex',
            justifyContent: 'center',
          }}
        >
          <button
            className="button button-3d"
            style={{
              padding: '8px 48px',
              backgroundColor: value && '#64b6f7',
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
