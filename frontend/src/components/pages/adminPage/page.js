import React from "react";
import { useState, useEffect } from "react";
import { useUser } from "../../../store/userContext";
import UserService from '../../../services/userService';
import { FiLogOut } from 'react-icons/fi';
import RandomAvatar from '../../../assets/images/RandomAvatar.jpg'
import '../../../styles/pages/_adminPage.scss';
import { Switch, Divider } from '@mui/material';
import { styled } from '@mui/material/styles';
import { useNavigate } from "react-router-dom";
import KeyIcon from '@mui/icons-material/Key';

const AntSwitch = styled(Switch)(({ theme }) => ({
    width: 42,
    height: 24,
    padding: 0,
    display: 'flex',
    '&:active': {
        '& .MuiSwitch-thumb': {
            width: 12,
        },
        '& .MuiSwitch-switchBase.Mui-checked': {
            transform: 'translateX(14px)',
        },
    },
    '& .MuiSwitch-switchBase': {
        padding: 3,
        '&.Mui-checked': {
            transform: 'translateX(18px)',
            color: '#fff',
            '& + .MuiSwitch-track': {
                opacity: 1,
                backgroundColor: '#1890ff',
                ...theme.applyStyles('dark', {
                    backgroundColor: '#177ddc',
                }),
            },
        },
    },
    '& .MuiSwitch-thumb': {
        boxShadow: '0 2px 4px 0 rgb(0 35 11 / 20%)',
        width: 18,
        height: 18,
        borderRadius: 9,
        transition: theme.transitions.create(['width'], {
            duration: 200,
        }),
    },
    '& .MuiSwitch-track': {
        borderRadius: 24 / 2,
        opacity: 1,
        backgroundColor: 'rgba(0,0,0,.25)',
        boxSizing: 'border-box',
        ...theme.applyStyles('dark', {
            backgroundColor: 'rgba(255,255,255,.35)',
        }),
    },
}));

const AdminPage = () => {
    const { userInfo } = useUser();
    console.log("User info.....", userInfo);
    const [users, setUsers] = useState([]);
    useEffect(() => {
        //Assuming some backend calls here, till the time feeding it with dummy data.
        setUsers([
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "SineAnand",
                state: true
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "SineAnand",
                state: false
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "SineAnand",
                state: true
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "SineAnand",
                state: false
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "SineAnand",
                state: true
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "SineAnand",
                state: false
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "SineAnand",
                state: true
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "SineAnand",
                state: false
            },
        ])
    }, [])
    const handleUserState = (selectedIndex) => {
        const tempUserList = [];
        for (let index = 0; index < users.length; index++) {
            const user = users[index];
            if (index === selectedIndex) {
                user.state = !user.state;
            }
            tempUserList.push(user);
        }
        setUsers(tempUserList);
    }
    const navigate = useNavigate();
    return (

        <div className="user-container">
            <div className="detailed-header">
                <div
                    style={{
                        flex: '1',
                        display: 'flex',
                        justifyContent: 'flex-start',
                        marginBottom: '-15px',
                    }}
                >
                    <button
                        onClick={() => navigate("/")}
                        className="btn text-dark mb-3"
                        style={{ display: 'flex', marginTop: '-10px' }}
                    >
                        <svg
                            width="24"
                            height="32"
                            viewBox="0 0 24 24"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                            style={{ marginRight: '8px' }}
                        >
                            <path
                                d="M10 19L3 12L10 5"
                                stroke="currentColor"
                                strokeWidth="2"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                            />
                            <path
                                d="M3 12H21"
                                stroke="currentColor"
                                strokeWidth="2"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                            />
                        </svg>
                    </button>
                    <span>My Profile</span>
                </div>
            </div>
            <div className="admin-card">
                <div className="admin-card-content">
                    <div style={{
                        display: "flex",
                        flexDirection: 'row',
                        gap: '20px'
                    }}>
                        <img src={users[0]?.avatar} alt="User" className="admin-card-content__avatar" />

                        <div className="admin-card-content__info">

                            <span className="admin-card-content__name">{users[0]?.username}
                                <span className={`admin-card-content__status`}>
                                    Admin
                                </span>
                            </span>
                            <span className="admin-card-content__email">{users[0]?.email}</span>
                        </div>
                    </div>
                    <button
                        className="button score-card-button generate-key-button"
                    >
                        Generate Key <KeyIcon />
                    </button>
                </div>
                {/* <Divider
                    sx={{
                        backgroundColor: 'black',
                        height: '1px',
                        width: '100%',
                        opacity: 0.1,
                    }}
                /> */}
                <div className="admin-card-">

                </div>
            </div>
            <span className="user-header">
                All Users
            </span>
            <div className="user-list">

                {users.map((user, index) => {
                    console.log("This user", index, user.username);
                    return (<div className="users-card">
                        <img src={user.avatar} alt="User" className="users-card__avatar" />
                        <div className="users-card__info">

                            <span className="users-card__name">{user.username}
                                <span className={`users-card__status ${!user.state ? 'users-card__status--inactive' : ''}`}>
                                    {user.state ? "Active" : "Inactive"}
                                </span>
                            </span>
                            <span className="users-card__email">{user.email}</span>
                        </div>
                        <div className="users-card__toggle">
                            <AntSwitch checked={user.state} onClick={() => {
                                handleUserState(index)
                            }} />
                        </div>
                    </div>
                    )
                })}
            </div>
        </div>
    );
}

export default AdminPage;



