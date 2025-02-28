import React, { useEffect } from 'react';
import styled from 'styled-components';
import { TbReportSearch } from "react-icons/tb";
import { reportService } from '../../../api/reports';

const Card = ({ authorId, authorName, Source, flagCount }) => {
    const [reportId, setReportId] = React.useState("");

    const handleGenerateReport = async () => {
        const response = await reportService.createReport({
            AuthorId: authorId,
            AuthorName: authorName,
            Source: Source,
            StartYear: 1990,
        });
        setReportId(response.Id);
    };


    return (
        <StyledWrapper>
            <div className="card-client">
                <div className="user-picture">
                    <svg viewBox="0 0 448 512" xmlns="http://www.w3.org/2000/svg">
                        <path d="M224 256c70.7 0 128-57.31 128-128s-57.3-128-128-128C153.3 0 96 57.31 96 128S153.3 256 224 256zM274.7 304H173.3C77.61 304 0 381.6 0 477.3c0 19.14 15.52 34.67 34.66 34.67h378.7C432.5 512 448 496.5 448 477.3C448 381.6 370.4 304 274.7 304z" />
                    </svg>
                </div>
                <p className="name-client">{authorName}
                    <span>Flag Count: {flagCount}
                    </span>
                </p>
                <div className="social-media">

                    {!reportId ? <button onClick={handleGenerateReport}
                        cursor="pointer">
                        <TbReportSearch />
                        <span>Generate report</span>
                    </button>
                        :
                        <a href={`/report/${reportId}`}
                            target="_blank"
                            rel="noopener noreferrer">
                            <button>
                                <TbReportSearch />
                                <span>See full report</span>
                            </button>
                        </a>
                    }
                </div>
            </div>
        </StyledWrapper>
    );
}

const StyledWrapper = styled.div`
  .card-client {
    background:rgb(255, 255, 255);
    width: 13rem;
    padding-top: 25px;
    padding-bottom: 25px;
    padding-left: 20px;
    padding-right: 20px;
    border: 4px solidrgb(255, 255, 255);
    box-shadow: 0 6px 10px rgba(207, 212, 222, 1);
    border-radius: 10px;
    text-align: center;
    color:rgb(54, 54, 54);
    font-family: "Poppins", sans-serif;
    transition: all 0.3s ease;
  }

  .card-client:hover {
    transform: translateY(-10px);
  }

  .user-picture {
    overflow: hidden;
    object-fit: cover;
    width: 5rem;
    height: 5rem;
    border: 4px solidrgb(255, 255, 255);
    border-radius: 999px;
    display: flex;
    justify-content: center;
    align-items: center;
    margin: auto;
  }

  .user-picture svg {
    width: 2.5rem;
    fill: currentColor;
  }

  .name-client {
    margin: 0;
    margin-top: 20px;
    font-weight: 600;
    font-size: 18px;
  }

  .name-client span {
    display: block;
    font-weight: 200;
    font-size: 16px;
  }

  .social-media:before {
    content: " ";
    display: block;
    width: 100%;
    height: 2px;
    margin: 20px 0;
    background:rgb(114, 103, 201);
  }

  .social-media a {
    position: relative;
    margin-right: 15px;
    text-decoration: none;
    color: rgb(39, 0, 197);
  }

  .social-media a:last-child {
    margin-right: 0;
  }

  .social-media a svg {
    width: 1.1rem;
    fill: rgb(39, 0, 197);
  }`;

export default Card;
