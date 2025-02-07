import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import SearchComponent from './SearchComponent';
import ItemDetails from './features/dashboard/page';
import EntityLookup from './features/dashboard/component/entityLookup/EntityLookup';

import "bootstrap/dist/css/bootstrap.css";
import "bootstrap/dist/js/bootstrap.bundle.js";
import './App.css';


function App() {
  return (
    <div className="App">
      <Router>
        <Routes>
          <Route path="/" element={<SearchComponent />} />
          <Route path="/item" element={<ItemDetails />} />
          <Route path="/entity-lookup" element={<EntityLookup />} />
        </Routes>
      </Router>
    </div>
  );
}

export default App;