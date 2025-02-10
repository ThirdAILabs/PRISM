// src/TodoListComponent.js
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiService from '../services/apiService';

const TodoListComponent = ({ results, canLoadMore, loadMore }) => {

  const navigate = useNavigate();
  const [titles, setTitles] = useState([]);

  const handleItemClick = (result) => {
    setTitles([]);
    navigate(`/item`, { state: { result } });
    return;
  };

  return (
    <div className='d-flex flex-column align-items-center w-100 '>
      <>
        <ul className='result-wrapper d-flex flex-wrap mt-3'>
          {results.filter((result) => (result.works_count === null) || (result.works_count > 0)).map((result, index) => (
            <li key={index} onClick={() => handleItemClick(result)} className='result-item'>
              <div className='text-start px-5'>
                <div className='d-flex align-items-center mb-2'>
                  <h5 className='m-0'>{result.display_name}</h5>
                  <p className='m-0 ms-2' style={{ fontSize: 'small' }}>{result.email}</p>
                </div>
                <b className='m-0 p-0' style={{ fontSize: 'small' }}>{result.institutions.join(', ')}</b>
                {result.works_count && <>
                  <br />
                  <b className='m-0 p-0' style={{ fontSize: 'small' }}>{`${result.works_count} works`}</b>
                </>}

              </div>
            </li>
          ))}
        </ul>

        {/* {
              canDeepSearch && !triedDp && !isDpLoading &&
              <button onClick={()=>{loadMore(null)}} 
                      className='btn bg-opacity-25 text-light rounded-5 border-primary p-4 d-flex justify-content-center align-items-center bg-primary mx-2 mt-1' 
              style={{maxHeight: '45px'}}>Deep Search</button>
            } */}

        {
          canLoadMore
            ?
            <button onClick={loadMore}
              className='btn bg-opacity-25 text-light rounded-5 border-primary p-4 d-flex justify-content-center align-items-center bg-primary mx-2 mt-1'
              style={{ maxHeight: '45px' }}>Load more</button>
            :
            <button className='btn bg-opacity-25 text-light rounded-5 border-primary p-4 d-flex justify-content-center align-items-center bg-primary mx-2 mt-1' disabled style={{ maxHeight: '45px' }}>Search Exhausted</button>
        }
        {/* {
              isLoading && <div className="spinner-border text-primary" style={{width: '3rem', height: '3rem'}} role="status"></div>
            } */}
        {/* {
              titles.length > 0 && (
                <>
                  <h3>Found {` ${titles.length} `} papers</h3>
                  {
                    titles.map((val, idx) => (<p key={idx}>{val}</p>))
                  }
                </>
              )
            } */}
      </>
    </div>
  );
};

export default TodoListComponent;
