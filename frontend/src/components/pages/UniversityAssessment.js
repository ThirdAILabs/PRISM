import React, { useState, useCallback, useContext } from 'react';
import Logo from '../../assets/images/prism-logo.png';
import '../common/searchBar/SearchBar.css';
import '../../styles/components/_primaryButton.scss';
import AutocompleteSearchBar from '../../utils/autocomplete';
import { autocompleteService } from '../../api/autocomplete';
import useCallOnPause from '../../hooks/useCallOnPause';
import { useNavigate } from 'react-router-dom';
import universityReportService from '../../api/universityReports';
import { UniversityContext } from '../../store/universityContext';

function UniversityAssessment() {
  const navigate = useNavigate();
  const { universityState, setUniversityState } = useContext(UniversityContext);

  const [institution, setInstitution] = useState(universityState.institution || null);
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

    setUniversityState((prev) => ({ ...prev, institution }));

    const reportData = {
      UniversityId: institution.Id,
      UniversityName: institution.Name,
      UniversityLocation: institution.Hint || '',
    };
    const reportId = await universityReportService.createReport(reportData);

    navigate(`report/${reportId.Id}`, {
      state: {
        canGoBack: true,
      },
    });
  };

  return (
    <div className="basic-setup" style={{ color: 'white' }}>
      <div style={{ textAlign: 'center', marginTop: '5%', animation: 'fade-in 0.75s' }}>
        <img
          src={Logo}
          alt="Prism Logo"
          style={{
            width: '240px',
            marginTop: '3%',
            marginBottom: '0.25%',
            marginRight: '2%',
            animation: 'fade-in 0.5s',
          }}
        />

        <div style={{ animation: 'fade-in 1s' }}>
          <div className="d-flex justify-content-center align-items-center">
            <div style={{ marginTop: 10, color: '#888888' }}>
              <h3
                style={{
                  fontWeight: 'bold',
                  color: 'black',
                  marginTop: 20,
                  animation: 'fade-in 0.75s',
                }}
              >
                University Assessment
              </h3>
            </div>
          </div>
          <div className="d-flex justify-content-center align-items-center">
            <div
              style={{
                marginTop: 10,
                marginBottom: '0.5%',
                color: '#888888',
                fontWeight: 'bold',
                fontSize: 'large',
              }}
            >
              Which university would you like to conduct an assessment on?
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
                    onSelect={(selected) => {
                      setInstitution(selected);
                      setUniversityState((prev) => ({ ...prev, institution: selected }));
                    }}
                    label={'Enter University Name'}
                    showHint={true}
                    initialValue={institution ? institution.InstitutionName : ''}
                  />
                </div>
                <div style={{ width: '200px', marginTop: '-2px' }}>
                  <button
                    className="button button-3d"
                    onClick={handleSearch}
                    disabled={!institution}
                  >
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
