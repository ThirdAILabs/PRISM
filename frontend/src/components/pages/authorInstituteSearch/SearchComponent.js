import React, { useState, useContext, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import AuthorInstitutionSearchComponent from './AuthorInstitutionSearch';
import OrcidSearchComponent from './OrcidSearch';
import { SearchContext } from '../../../store/searchContext';
import PaperTitleSearchComponent from './PaperTitleSearch';
import Logo from '../../../assets/images/prism-logo.png';
import RowRadioButtonsGroup from '../../common/tools/RadioButton';

const SearchComponent = () => {
  const { searchState, setSearchState } = useContext(SearchContext);

  const location = useLocation();
  const navigate = useNavigate();
  const params = new URLSearchParams(location.search);
  const defaultType = params.get('type') || 'author';
  const [radioButtonProps, setRadioButtonProps] = useState([]);
  const [selectedSearchType, setSelectedSearchType] = useState(defaultType);

  const handleSearchTypeChange = (e) => {
    const newType = e.target.value;
    if (!newType) {
      return;
    }

    setSelectedSearchType(newType);

    setSearchState((prev) => ({
      ...prev,
      // author/institution data
      author: null,
      institution: null,
      openAlexResults: [],
      hasSearched: false,
      loadMoreCount: 0,
      canLoadMore: true,
      isOALoading: false,

      // orcid data
      orcidResults: [],
      isOrcidLoading: false,
      hasSearchedOrcid: false,
      orcidQuery: '',

      // paper data
      paperResults: [],
      isPaperLoading: false,
      hasSearchedPaper: false,
      paperTitleQuery: '',
    }));

    navigate(`?type=${newType}`);
  };

  useEffect(() => {
    const newType = params.get('type') || 'author';
    setSelectedSearchType(newType);
  }, [location.search, params]);

  useEffect(() => {
    setRadioButtonProps([
      { value: 'author', label: 'Author & Institution' },
      { value: 'paper', label: 'Paper Title' },
      { value: 'orcid', label: 'ORCID ID' },
    ])
  }, []);

  return (
    <div className="basic-setup" style={{ color: 'black' }}>
      <div style={{ textAlign: 'center', marginTop: '3%', animation: 'fade-in 0.75s' }}>
        <img
          src={Logo}
          alt="Prism Logo"
          style={{
            width: '320px',
            marginTop: '1%',
            marginBottom: '1%',
            marginRight: '2%',
            animation: 'fade-in 0.5s',
          }}
        />
        <h1 style={{ fontWeight: 'bold', marginTop: 20, animation: 'fade-in 0.75s', fontFamily: 'serif' }}>
          Individual Assessment
        </h1>
        <div style={{ animation: 'fade-in 1s' }}>
          <div className="d-flex justify-content-center align-items-center">
            <div style={{ marginTop: 10, color: '#888888' }}>
              We help you comply with research security requirements by automating author
              assessments.
            </div>
          </div>
          <div className="d-flex justify-content-center align-items-center">
            <div style={{ marginTop: 10, marginBottom: '1%', color: '#888888' }}>
              Who would you like to conduct an assessment on?
            </div>
          </div>
        </div>
        <div style={{
          marginTop: '1rem',
          color: 'rgb(100, 100, 100)',
        }}>
          <RowRadioButtonsGroup
            selectedSearchType={selectedSearchType}
            formControlProps={radioButtonProps}
            handleSearchTypeChange={handleSearchTypeChange}
          />
        </div>
      </div>
      <div className="d-flex justify-content-center align-items-center">
        <div style={{ width: '80%', animation: 'fade-in 1.25s' }}>
          {selectedSearchType === 'author' && <AuthorInstitutionSearchComponent />}
          {selectedSearchType === 'orcid' && <OrcidSearchComponent />}
          {selectedSearchType === 'paper' && <PaperTitleSearchComponent />}
        </div>
      </div>
    </div>
  );
};

export default SearchComponent;
