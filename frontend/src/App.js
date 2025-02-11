import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import SearchComponent from './components/pages/SearchComponent';
import ItemDetails from './components/pages/itemDetails/page';
import EntityLookup from './components/pages/entityLookup/page';
// import PrivateComponent from '../';

import "bootstrap/dist/css/bootstrap.css";
import "bootstrap/dist/js/bootstrap.bundle.js";
import './App.css';


function App() {
  return (
    <div className="App">
      <Router>
        <Routes>
          {/* <Route element={<PrivateComponent />}> */}
          <Route path="/" element={<SearchComponent />} />
          <Route path="/item" element={<ItemDetails />} />
          <Route path="/entity-lookup" element={<EntityLookup />} />
          {/* </Route> */}
          <Route
            path="/login"
            element={<h1>HUE HUE HUE</h1>} />
        </Routes>
      </Router>
    </div>
  );
}

export default App;