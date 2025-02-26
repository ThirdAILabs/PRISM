import React, { createContext, useState, useContext } from 'react';

const UserContext = createContext(null);

export const useUser = () => useContext(UserContext);

export const UserProvider = ({ children }) => {
  const [userInfo, setUserInfo] = useState({
    name: '',
    email: '',
    username: '',
    accessToken: '',
  });

  const updateUserInfo = (newInfo) => {
    setUserInfo((prev) => ({ ...prev, ...newInfo }));
  };

  return (
    <UserContext.Provider value={{ userInfo, updateUserInfo }}>{children}</UserContext.Provider>
  );
};
