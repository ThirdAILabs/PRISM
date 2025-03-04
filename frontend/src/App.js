import React, { useEffect, useState } from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import SearchComponent from './components/pages/authorInstituteSearch/SearchComponent';
import ItemDetails from './components/pages/itemDetails/page';
import EntityLookup from './components/pages/entityLookup/page';
import UserService from './services/userService';
import { useUser } from './store/userContext';
import { FaBars } from 'react-icons/fa';
import SidePanel from './components/sidebar/SidePanel';
import UniversityAssessment from './components/pages/UniversityAssessment';
import UniversityReport from './components/pages/UniversityReport';
import { useLocation } from 'react-router-dom';
import Error from './components/pages/Error';
import SearchProviderWrapper from './services/SearchProviderWrapper';
import UniversityProviderWrapper from './services/UniversityProviderWrapper';
//CSS
import 'bootstrap/dist/css/bootstrap.css';
import 'bootstrap/dist/js/bootstrap.bundle.js';
import './App.css';

function App() {
  return (
    <Router>
      <AppContent />
    </Router>
  );
}

function AppContent() {
  const { updateUserInfo } = useUser();
  const location = useLocation();
  const [isSidePanelOpen, setIsSidePanelOpen] = useState(false);

  useEffect(() => {
    const tokenParsed = UserService.getTokenParsed();
    const accessToken = UserService.getToken();
    if (tokenParsed) {
      const { preferred_username: username, email, name } = tokenParsed;
      updateUserInfo({ name: name || username, email, username, accessToken });
    }
  }, []);

  const showMenuIcon = !location.pathname.includes('report');

  return (
    <div className="App">
      {showMenuIcon && (
        <FaBars
          size={30}
          style={{
            cursor: 'pointer',
            position: 'fixed',
            left: '20px',
            top: '20px',
            zIndex: 1000,
          }}
          onClick={() => setIsSidePanelOpen(!isSidePanelOpen)}
          className="hover:bg-gray-200"
        />
      )}
      <SidePanel isOpen={isSidePanelOpen} onClose={() => setIsSidePanelOpen(false)} />
      <Routes>
        <Route element={<SearchProviderWrapper />}>
          <Route path="/" element={<SearchComponent />} />
          <Route path="/report/:report_id" element={<ItemDetails />} />
        </Route>
        <Route path="/entity-lookup" element={<EntityLookup />} />
        <Route element={<UniversityProviderWrapper />}>
          <Route path="/university" element={<UniversityAssessment />} />
          <Route path="/university/report/:report_id" element={<UniversityReport />} />
        </Route>
        <Route path="/error" element={<Error />} />
      </Routes>
    </div>
  );
}

export default App;
