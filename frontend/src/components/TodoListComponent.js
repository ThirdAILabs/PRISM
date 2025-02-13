// src/TodoListComponent.js
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { reportService } from '../api/reports';

const TodoListComponent = ({ results, canLoadMore, loadMore }) => {

  const navigate = useNavigate();

  const handleItemClick = async (result) => {
    const response = await reportService.createReport({
      "AuthorId": result.AuthorId, "AuthorName": result.AuthorName, "Source": result.Source,
      "StartYear": -1, "EndYear": 3000,
    });
    console.log("Response", response);
    // navigate(`/item`, { state: { response } });
    return;
  };

  return (
    <div className='d-flex flex-column align-items-center w-100 '>
      <>
        <ul className='result-wrapper d-flex flex-wrap mt-3'>
          {results.map((result, index) => (
            <li key={index} onClick={() => handleItemClick(result)} className='result-item'>
              <div className='text-start px-5'>
                <div className='d-flex align-items-center mb-2'>
                  <h5 className='m-0'>{result.AuthorName}</h5>
                  {/* <p className='m-0 ms-2' style={{ fontSize: 'small' }}>{result.email}</p> */}
                </div>
                <b className='m-0 p-0' style={{ fontSize: 'small' }}>{result.Institutions.join(', ')}</b>
              </div>
            </li>
          ))}
        </ul>

      </>
    </div>
  );
};

export default TodoListComponent;
