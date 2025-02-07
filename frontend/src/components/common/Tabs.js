// Tabs.js
import React from 'react';
import { Tabs, Tab, Box, Divider } from '@mui/material';
import { makeStyles } from '@mui/styles';

const useStyles = makeStyles({
    tabs: {
        '& .MuiTabs-indicator': {
            backgroundColor: 'white',
            height: '3px', // Makes the active tab indicator bolder
        },
    },
    tab: {
        color: 'white !important',
        '&.Mui-selected': {
            color: 'white !important',
        },
        '&:hover': {
            backgroundColor: 'rgba(255, 255, 255, 0.1) !important', // Light gray hover effect
            transition: 'background-color 0.3s',
        },
    },
});

const CustomTabs = ({ activeTab, handleTabChange }) => {
    const classes = useStyles();

    return (
        <Box>
            <Box sx={{ display: 'flex', 
                    justifyContent: 'center', // Center the Tabs horizontally
                    borderBottom: 0  }}>
                <Tabs
                    value={activeTab}
                    onChange={handleTabChange}
                    className={classes.tabs}
                    TabIndicatorProps={{
                        style: {
                            backgroundColor: 'white'
                        }
                    }}
                    sx={{
                        '& .MuiTabs-flexContainer': {
                            gap: '2rem'
                        },
                        minHeight: '48px'
                    }}
                >
                    <Tab
                        label="Dashboard"
                        className={classes.tab}
                        sx={{
                            textTransform: 'none',
                            minHeight: '48px',
                            padding: '12px 16px'
                        }}
                    />
                    <Tab
                        label="Graph Visualization"
                        className={classes.tab}
                        sx={{
                            textTransform: 'none',
                            minHeight: '48px',
                            padding: '12px 16px'
                        }}
                    />
                </Tabs>
            </Box>
            {/* Added white divider below tabs */}
            <Divider sx={{
                backgroundColor: 'white',
                height: '1px',
                width: '100%',
                opacity: 0.1
            }} />
        </Box>
    );
};

export default CustomTabs;