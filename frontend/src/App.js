import React, { useEffect } from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import SearchComponent from './components/pages/SearchComponent';
import ItemDetails from './components/pages/itemDetails/page';
import EntityLookup from './components/pages/entityLookup/page';
import UserService from './services/userService';
import { useUser } from './store/userContext';

import "bootstrap/dist/css/bootstrap.css";
import "bootstrap/dist/js/bootstrap.bundle.js";
import './App.css';
import AuthorSearch from './components/pages/author';



function App() {
  const { updateUserInfo } = useUser();

  useEffect(() => {
    const tokenParsed = UserService.getTokenParsed();
    const accessToken = UserService.getToken();
    if (tokenParsed) {
      const { preferred_username: username, email, name } = tokenParsed;
      updateUserInfo({ name: name || username, email, username, accessToken });
    }
  }, []);

  return (
    <div className="App">

      <Router>
        <Routes>
          <Route path="/" element={<SearchComponent />} />
          <Route path="/item" element={<ItemDetails />} />
          <Route path="/auto" element={<AuthorSearch />} />
          <Route path="/entity-lookup" element={<EntityLookup />} />
        </Routes>
      </Router>
    </div>
  );
}

export default App;