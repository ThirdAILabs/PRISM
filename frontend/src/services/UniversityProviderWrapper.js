import React from 'react';
import { Outlet } from 'react-router-dom';
import { UniversityProvider } from '../store/universityContext';

const UniversityProviderWrapper = () => {
  return (
    <UniversityProvider>
      <Outlet />
    </UniversityProvider>
  );
};

export default UniversityProviderWrapper;
