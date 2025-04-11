import React, { useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useRef } from 'react';
import AuthorInstitutionSearchComponent from './AuthorInstitutionSearch';
import OrcidSearchComponent from './OrcidSearch';
import PaperTitleSearchComponent from './PaperTitleSearch';
import Logo from '../../../assets/images/prism-logo.png';
import RowRadioButtonsGroup from '../../common/tools/RadioButton';
import { CSSTransition, SwitchTransition } from 'react-transition-group';
import './SearchComponent.css';

const SearchComponent = () => {
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
    navigate(`?type=${newType}`);
  };

  useEffect(() => {
    setSelectedSearchType(defaultType);
  }, [defaultType]);

  useEffect(() => {
    setRadioButtonProps([
      { value: 'author', label: 'Author & Institution' },
      { value: 'paper', label: 'Paper Title' },
      { value: 'orcid', label: 'ORCID ID' },
    ]);
  }, []);

  const nodeRef = useRef(null);
  return (
    <div className="basic-setup" style={{ color: 'black' }}>
      <div style={{ textAlign: 'center', marginTop: '3%', animation: 'fade-in 0.75s' }}>
        <img
          src={Logo}
          alt="Prism Logo"
          style={{
            width: '240px',
            marginTop: '5%',
            marginBottom: '0.25%',
            marginRight: '2%',
            animation: 'fade-in 0.5s',
          }}
        />
        <h4
          style={{
            fontWeight: 'bold',
            marginTop: 20,
            animation: 'fade-in 0.75s',
          }}
        >
          We Simplify Research Security Compliance
        </h4>
        <div style={{ animation: 'fade-in 1s' }}>
          <div className="d-flex justify-content-center align-items-center">
            <div
              style={{
                marginTop: 10,
                marginBottom: '1%',
                color: '#888888',
                fontWeight: 'bold',
                fontSize: 'large',
              }}
            >
              Who would you like to assess?
            </div>
          </div>
        </div>
        <div
          style={{
            marginTop: '1rem',
          }}
        >
          <RowRadioButtonsGroup
            selectedSearchType={selectedSearchType}
            formControlProps={radioButtonProps}
            handleSearchTypeChange={handleSearchTypeChange}
          />
        </div>
      </div>
      <div className="d-flex justify-content-center align-items-center">
        <div style={{ width: '80%' }}>
          <SwitchTransition mode="out-in">
            <CSSTransition
              key={selectedSearchType}
              timeout={300}
              classNames="fade"
              nodeRef={nodeRef}
            >
              <div ref={nodeRef}>
                {selectedSearchType === 'author' && <AuthorInstitutionSearchComponent />}
                {selectedSearchType === 'orcid' && <OrcidSearchComponent />}
                {selectedSearchType === 'paper' && <PaperTitleSearchComponent />}
              </div>
            </CSSTransition>
          </SwitchTransition>
        </div>
      </div>
    </div>
  );
};

export default SearchComponent;
