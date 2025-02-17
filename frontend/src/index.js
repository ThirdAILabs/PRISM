import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import UserService from './services/userService';
import { UserProvider } from './store/userContext';
const root = ReactDOM.createRoot(document.getElementById('root'));

UserService.initKeycloak(() => {
  root.render(
    <React.StrictMode>
      <UserProvider>
        <App />
      </UserProvider>
    </React.StrictMode>
  );
});
