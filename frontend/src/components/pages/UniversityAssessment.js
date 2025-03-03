import React, { useState, useCallback } from 'react';
import Logo from '../../assets/images/prism-logo.png';
import '../common/searchBar/SearchBar.css';
import '../common/tools/button/button1.css';
import AutocompleteSearchBar from '../../utils/autocomplete';
import { autocompleteService } from '../../api/autocomplete';
import useCallOnPause from '../../hooks/useCallOnPause';
import { useNavigate } from 'react-router-dom';
import universityReportService from '../../api/universityReports';

function UniversityAssessment() {
  const navigate = useNavigate();
  const [institution, setInstitution] = useState();
  const [results, setResults] = useState([]);
  const [isLoading, setIsLoading] = useState(false);

  const debouncedSearch = useCallOnPause(300); // 300ms delay

  const autocompleteInstitution = useCallback(
    (query) => {
      return new Promise((resolve) => {
        debouncedSearch(async () => {
          try {
            const res = await autocompleteService.autocompleteInstitutions(query);
            resolve(res);
            return res;
          } catch (error) {
            console.error('Autocomplete error:', error);
            resolve([]);
          }
        });
      });
    },
    [debouncedSearch]
  );

  const handleSearch = async () => {
    if (!institution) {
      alert('Please select an institution');
      return;
    }
    const reportData = {
      UniversityId: institution.InstitutionId,
      UniversityName: institution.InstitutionName,
    };
    const reportId = await universityReportService.createReport(reportData);
    console.log('University report id: ', reportId);
    navigate(`report/${reportId.Id}`);
  };

  return (
    <div className="basic-setup" style={{ color: 'white' }}>
      <div style={{ textAlign: 'center', marginTop: '5%', animation: 'fade-in 0.75s' }}>
        <img
          src={Logo}
          style={{
            width: '320px',
            marginTop: '1%',
            marginBottom: '1%',
            marginRight: '2%',
            animation: 'fade-in 0.5s',
          }}
        />

        <div style={{ animation: 'fade-in 1s' }}>
          <div className="d-flex justify-content-center align-items-center">
            <div style={{ marginTop: 10, color: '#888888' }}>
              <h1
                style={{
                  fontWeight: 'bold',
                  color: 'black',
                  marginTop: 20,
                  animation: 'fade-in 0.75s',
                }}
              >
                University Assessment
              </h1>
              We help you comply with research security requirements by automating university
              assessments.
            </div>
          </div>
          <div className="d-flex justify-content-center align-items-center">
            <div style={{ marginTop: 10, marginBottom: '2%', color: '#888888' }}>
              Who would you like to conduct an assessment on?
            </div>
          </div>
        </div>
        <div style={{ animation: 'fade-in 1s' }}>
          <div className="d-flex justify-content-center align-items-center">
            <div style={{ width: '80%' }}>
              <div className="author-institution-search-bar">
                <div className="autocomplete-search-bar">
                  <AutocompleteSearchBar
                    autocomplete={autocompleteInstitution}
                    onSelect={setInstitution}
                    type={'institute'}
                    title={'University'}
                  />
                </div>
                <div style={{ width: '40px' }} />
                <div style={{ width: '200px', marginTop: '40px' }}>
                  <button className="button" onClick={handleSearch}>
                    {isLoading ? 'Searching...' : 'Search'}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default UniversityAssessment;
