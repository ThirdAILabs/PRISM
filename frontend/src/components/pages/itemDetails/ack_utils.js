import React from 'react';

export const createHighlights = (fundCodeTriangulation, authorName) => {
  if (!fundCodeTriangulation || typeof fundCodeTriangulation !== 'object') {
    return [];
  }

  return Object.entries(fundCodeTriangulation).flatMap(([funderName, funderMap]) =>
    Object.entries(funderMap)
      .map(([grantCode, isRecipient]) => ({
        regex: new RegExp(grantCode, 'i'),
        color: isRecipient ? 'red' : 'green',
        tooltip: `${authorName} is ${isRecipient ? 'likely' : 'likely NOT'} a primary recipient of this fund.`,
      }))
  );
};

export const applyHighlighting = (text, highlights) => {
  if (!highlights.length) {
    return text;
  }

  let parts = [text];

  highlights.forEach(({ regex, color, tooltip }) => {
    parts = parts.flatMap((part) => {
      if (typeof part !== 'string') return part;

      const match = part.match(regex);
      if (!match) return part;

      const splitParts = part.split(regex);
      return splitParts.flatMap((splitPart, j, arr) => {
        if (j >= arr.length - 1) return splitPart;

        return [
          splitPart,
          <span key={`highlight-${j}`} style={{ color }} title={tooltip}>
            <strong>{match[0]}</strong>
          </span>,
        ];
      });
    });
  });

  return parts;
};

export const hasValidTriangulationData = (triangulationData) => {
  return (
    triangulationData &&
    typeof triangulationData === 'object' &&
    Object.keys(triangulationData).length > 0
  );
};
