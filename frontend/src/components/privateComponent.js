import { useEffect } from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import UserService from '../services/userService';

const PrivateComponent = () => {
  useEffect(() => {
    if (!UserService.isLoggedIn()) {
      UserService.doLogin(); // Force immediate redirect
    }
  }, []);

  return UserService.isLoggedIn() ? <Outlet /> : <Navigate to="/login" />;
};
