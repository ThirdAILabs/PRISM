import { Gauge, GaugeValueArc, useGaugeState, gaugeClasses } from '@mui/x-charts/Gauge';
import * as React from 'react';

export function GradientValueGauge() {
  const { outerRadius } = useGaugeState();
  const x1 = `${-outerRadius}px`;
  const x2 = `${outerRadius}px`;
  return (
    <svg width="200" height="200">
      <defs>
        <linearGradient id="gauge-gradient" x1={x1} x2={x2} gradientUnits="userSpaceOnUse">
          <stop offset="0%" style={{ stopColor: 'blue', stopOpacity: 1 }} />
          <stop offset="100%" style={{ stopColor: 'red', stopOpacity: 1 }} />
        </linearGradient>
      </defs>
      <GaugeValueArc style={{ fill: 'url(#gauge-gradient)' }} />
    </svg>
  );
}

export function Ticks({ scale }) {
  const { innerRadius, cx, cy, startAngle, endAngle } = useGaugeState();
  const radius = innerRadius * 0.8;
  function angleAtStep(step) {
    return -Math.PI / 2 + startAngle + (step / (scale.length - 1)) * (endAngle - startAngle);
  }
  return (
    <g>
      {scale.map((val, step) => {
        const tickCx = cx + radius * Math.cos(angleAtStep(step));
        const tickCy = cy + radius * Math.sin(angleAtStep(step));
        return (
          <text
            key={step}
            x={tickCx}
            y={tickCy}
            style={{ fill: 'black' }}
            fontSize={0.2 * radius}
            textAnchor="middle"
            dominantBaseline="middle"
          >
            {val}
            {step === scale.length - 1 ? '+' : ''}
          </text>
        );
      })}
    </g>
  );
}

export function Value({ value, speedometerHoverText }) {
  const getFontSize = (value) => {
    // Base font size for single digit (largest size)
    const baseFontSize = 48;

    const digits = String(value).length;
    // Each additional digit reduces size by 4px
    const newSize = baseFontSize - (digits - 1) * 4;

    return `${Math.max(newSize, 28)}px`;
  };
  const { cx, cy } = useGaugeState();
  return (
    <g>
      <text
        x={cx}
        y={cy * 1.1}
        style={{ 
          fill: 'grey',
          fontSize: getFontSize(value),
          fontWeight: 'bold',
          textAnchor: 'middle',
          dominantBaseline: 'middle',
        }}
      >
        <title>{speedometerHoverText ? speedometerHoverText : null}</title>
        {value}
      </text>
    </g>
  );
}

export function Speedometer({ scale, value, speedometerHoverText }) {
  function transformValue(value) {
    if (value >= scale[scale.length - 1]) {
      return 100;
    }
    for (let i = 2; i <= scale.length; i++) {
      if (scale[scale.length - i] < value) {
        const segmentWidth = scale[scale.length - i + 1] - scale[scale.length - i];
        const delta = value - scale[scale.length - i];
        return (100 * (scale.length - i + delta / segmentWidth)) / (scale.length - 1);
      }
    }
    return 0;
  }
  return (
    <div className="chart-wrapper">
      <Gauge
        value={transformValue(value)}
        text={''}
        startAngle={-120}
        endAngle={120}
        cornerRadius="50%"
        innerRadius="80%"
        outerRadius="100%"
        sx={(theme) => ({
          [`& .${gaugeClasses.referenceArc}`]: {
            fill: 'rgb(245, 240, 240)',
          },
        })}
      >
        <Ticks fill={'white'} scale={scale} />
        <GradientValueGauge />
        <Value value={value} speedometerHoverText={speedometerHoverText} />
      </Gauge>
    </div>
  );
}
