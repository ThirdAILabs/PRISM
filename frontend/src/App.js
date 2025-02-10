import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import SearchComponent from './components/pages/SearchComponent';
import ItemDetails from './components/pages/itemDetails/page';
import EntityLookup from './components/pages/entityLookup/page';

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