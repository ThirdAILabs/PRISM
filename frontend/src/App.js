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
import Error from './components/pages/error/Error.js';
import { GetShowMenuIcon } from './utils/helper.js';
import SearchProviderWrapper from './services/SearchProviderWrapper';
import UniversityProviderWrapper from './services/UniversityProviderWrapper';
import { Tooltip } from '@mui/material';
import useOutsideClick from './hooks/useOutsideClick.js';

import { GoSidebarCollapse } from 'react-icons/go';
import { GoSidebarExpand } from 'react-icons/go';

//CSS
import 'bootstrap/dist/css/bootstrap.css';
import 'bootstrap/dist/js/bootstrap.bundle.js';
import './App.css';
import AdminPage from './components/pages/adminPage/page.js';

function App() {
  return (
    <Router>
      <AppContent />
    </Router>
  );
}

function AppContent() {
  const { updateUserInfo } = useUser();
  const [isSidePanelOpen, setIsSidePanelOpen] = useState(false);

  useEffect(() => {
    const tokenParsed = UserService.getTokenParsed();
    const accessToken = UserService.getToken();
    if (tokenParsed) {
      const { preferred_username: username, email, name } = tokenParsed;
      updateUserInfo({ name: name || username, email, username, accessToken });
    }
  }, []);

  const showMenuIcon = GetShowMenuIcon();
  const sidepanelRef = useOutsideClick(() => {
    setIsSidePanelOpen(false);
  });

  return (
    <div className="App">
      {isSidePanelOpen && showMenuIcon && (
        <div className="overlay" onClick={() => setIsSidePanelOpen(false)} />
      )}

      {showMenuIcon && (
        <div
          className={`sidebar-toggle ${isSidePanelOpen ? '' : 'closed'}`}
          onClick={() => setIsSidePanelOpen(!isSidePanelOpen)}
        >
          {<GoSidebarCollapse size={30} />}
        </div>
      )}

      <div ref={sidepanelRef} className="sidepanel-container">
        <SidePanel
          isOpen={isSidePanelOpen && showMenuIcon}
          onClose={() => setIsSidePanelOpen(false)}
        />
      </div>

      <div className="content-container">
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
          <Route path="/admin-page" element={<AdminPage />} />
          <Route path="*" element={<Error />} />
        </Routes>
      </div>
    </div>
  );
}

export default App;
