import React, { createContext, useState } from 'react';

export const UniversityContext = createContext();

export const UniversityProvider = ({ children }) => {
  const [universityState, setUniversityState] = useState({
    institution: null,
  });

  return (
    <UniversityContext.Provider value={{ universityState, setUniversityState }}>
      {children}
    </UniversityContext.Provider>
  );
};
