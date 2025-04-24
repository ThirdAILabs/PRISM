import React from "react";
import { useState, useEffect } from "react";
import RandomAvatar from '../../../assets/images/RandomAvatar.jpg'
import OpenAI from '../../../assets/images/OpenAI.png'
import Google from '../../../assets/images/Google.png'

import '../../../styles/pages/_adminPage.scss';
import { Switch } from '@mui/material';
import { styled } from '@mui/material/styles';
import { useNavigate } from "react-router-dom";
import KeyIcon from '@mui/icons-material/Key';
import SearchIcon from '@mui/icons-material/Search';
import MonetizationOnIcon from '@mui/icons-material/MonetizationOn';
import { CiSearch } from 'react-icons/ci';
// import '../../../styles/pages/_authorCard.scss';

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


const CircleProgress = ({ percentage, variant = 'primary' }) => {
    const radius = 21;
    const circumference = 2 * Math.PI * radius;
    const offset = circumference - (percentage / 100) * circumference;

    return (
        <svg className={`circle-progress circle-progress--${variant}`} viewBox="0 0 48 48">
            <circle
                className="circle-progress__circle circle-progress__circle--bg"
                cx="24"
                cy="24"
                r={radius}
            />
            <circle
                className="circle-progress__circle circle-progress__circle--progress"
                cx="24"
                cy="24"
                r={radius}
                strokeDasharray={circumference}
                strokeDashoffset={offset}
            />
        </svg>
    );
};

const AdminPage = () => {
    const [users, setUsers] = useState([]);
    const [initialUsers, setInitialUsers] = useState([]);
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
                username: "gautam",
                state: false
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "pratik",
                state: true
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "siddharth",
                state: false
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "Tharun",
                state: true
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "Nicholas",
                state: false
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "Benito",
                state: true
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "Shubh",
                state: false
            },
        ])
        setInitialUsers([
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
                username: "gautam",
                state: false
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "pratik",
                state: true
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "siddharth",
                state: false
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "Tharun",
                state: true
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "Nicholas",
                state: false
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "Benito",
                state: true
            },
            {
                avatar: RandomAvatar,
                name: "Anand Kumar",
                email: "anand@thirdai.com",
                username: "Shubh",
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
    const [selectedUserTab, setSelectedUserTab] = useState("All");
    const handleTabClick = (tabName) => {
        setSelectedUserTab(tabName);
    }

    const isShowUserCard = (status) => {
        if (selectedUserTab === "All")
            return true;
        if (status && selectedUserTab === "Active")
            return true;
        if (!status && selectedUserTab === "Inactive")
            return true;
        return false;
    }

    const [searchTerm, setSearchTerm] = useState('');
    useEffect(() => {
        if (searchTerm === '') {
            setUsers(initialUsers);
        }
        setUsers(initialUsers.filter((user) => {
            if (user.username.toLocaleLowerCase().includes(searchTerm.toLocaleLowerCase()))
                return true;
            return false;
        }))
    }, [searchTerm, initialUsers])

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
                        className="button button-3d generate-key-button"
                    >
                        Generate Key <KeyIcon />
                    </button>
                </div>

                <div className="stats">
                    <div className="stats-card-1">
                        <div style={{
                            display: "flex",
                            flexDirection: 'row',
                            gap: '36px'
                        }}>
                            <div className="stats__avatar">
                                <img src={Google} alt="OpenAI" className="users-card__avatar" />

                            </div>
                            <div className="stats__info">

                                <span className="stats__name">
                                    {"Total Searches"}
                                </span>
                                <span className="stats__email">{"30,000"}</span>
                            </div>
                        </div>
                    </div>
                    <div className="stats-card-2">
                        <div style={{
                            display: "flex",
                            flexDirection: 'row',
                            gap: '36px'
                        }}>
                            <div className="stats__avatar">
                                <CircleProgress percentage={70} />
                            </div>
                            <div className="stats__info">

                                <span className="stats__name">
                                    {"Plan Search Left"}
                                </span>
                                <span className="stats__email">{"14,000"}</span>
                            </div>
                        </div>
                    </div>
                    <div className="stats-card-3">
                        <div style={{
                            display: "flex",
                            flexDirection: 'row',
                            gap: '36px'
                        }}>
                            <div className="stats__avatar">
                                <CircleProgress percentage={50} />
                            </div>
                            <div className="stats__info">

                                <span className="stats__name">
                                    {"Total Searches Left"}
                                </span>
                                <span className="stats__email">{"18,000"}</span>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="stats">
                    <div className="stats-card-1">
                        <div style={{
                            display: "flex",
                            flexDirection: 'row',
                            gap: '36px'
                        }}>
                            <div className="stats__avatar">
                                <img src={OpenAI} alt="OpenAI" className="users-card__avatar" />
                            </div>
                            <div className="stats__info">

                                <span className="stats__name">
                                    {"Total Cost"}
                                </span>
                                <span className="stats__email">{"30,000"}</span>
                            </div>
                        </div>
                    </div>
                    <div className="stats-card-2">
                        <div style={{
                            display: "flex",
                            flexDirection: 'row',
                            gap: '36px'
                        }}>
                            <div className="stats__avatar">
                                <CircleProgress percentage={42} variant="warning" />
                            </div>
                            <div className="stats__info">

                                <span className="stats__name">
                                    {"Input Token"}
                                </span>
                                <span className="stats__email">{"14,000"}</span>
                            </div>
                        </div>
                    </div>
                    <div className="stats-card-3">
                        <div style={{
                            display: "flex",
                            flexDirection: 'row',
                            gap: '36px'
                        }}>
                            <div className="stats__avatar">
                                <CircleProgress percentage={42} variant="warning" />
                            </div>
                            <div className="stats__info">

                                <span className="stats__name">
                                    {"Output Token"}
                                </span>
                                <span className="stats__email">{"18,000"}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div className="user-header">
                <div className="user-header-tab">
                    <span className={`user-header-tabs ${selectedUserTab === "All" ? 'active' : ''}`} onClick={() => {
                        handleTabClick("All")
                    }}>
                        All Users
                    </span>
                    <span className={`user-header-tabs ${selectedUserTab === "Active" ? 'active' : ''}`} onClick={() => {
                        handleTabClick("Active")
                    }}>
                        Active Users
                    </span>
                    <span className={`user-header-tabs ${selectedUserTab === "Inactive" ? 'active' : ''}`} onClick={() => {
                        handleTabClick("Inactive")
                    }}>
                        Inactive Users
                    </span>
                </div>
                <div className="search-container">
                    <div className="search-icon">
                        <CiSearch />
                    </div>
                    <input
                        type="text"
                        placeholder="Search by name..."
                        className="search-input"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
            </div>
            <div className="user-list">

                {users.map((user, index) => {
                    return (
                        isShowUserCard(user.state) && <div className="users-card">
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



