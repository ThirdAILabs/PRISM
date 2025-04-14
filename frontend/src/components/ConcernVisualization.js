import * as React from 'react';
import { Speedometer } from './common/tools/Speedometer';
import '../App.css';
import '../styles/components/_primaryButton.scss';
import Tooltip from './common/tools/Tooltip';

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
        width: '100%',
        height: '320px',
        position: 'relative',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
      }}
    >
      <div style={{ width: '100%', display: 'flex', justifyContent: 'center', marginTop: '-20px' }}>
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
          width: '96%',
          fontSize: '14px',
          lineHeight: '20px',
          letterSpacing: '0.1px',
          fontWeight: 'bold',
          color: value !== 0 ? '#333f54' : '#6a798f',
          paddingInline: '16px',
        }}
      >
        {title} <Tooltip text={hoverText} />
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
              backgroundColor: value !== 0 && '#64b6f7',
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
