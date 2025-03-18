// Tabs.js
import React from 'react';
import { Tabs, Tab, Box, Divider, Tooltip } from '@mui/material';
import { makeStyles } from '@mui/styles';

const useStyles = makeStyles({
  tabs: {
    '& .MuiTabs-indicator': {
      backgroundColor: 'black',
      height: '3px', // Makes the active tab indicator bolder
    },
  },
  tab: {
    color: 'black !important',
    '&.Mui-selected': {
      color: 'black !important',
    },
    '&:hover': {
      backgroundColor: 'rgb(245,240,240) !important', // Light gray hover effect
      transition: 'background-color 0.3s',
    },
  },
});

const CustomTabs = ({ activeTab, handleTabChange, disabled }) => {
  const classes = useStyles();

  return (
    <Box>
      <Box
        sx={{
          position: 'relative',
          display: 'flex',
          justifyContent: 'center', // Center the Tabs horizontally
          alignItems: 'center',
          borderBottom: 0,
          minHeight: '48px',
        }}
      >
        <Tabs
          value={activeTab}
          onChange={handleTabChange}
          className={classes.tabs}
          TabIndicatorProps={{
            style: {
              backgroundColor: 'black',
            },
          }}
          sx={{
            '& .MuiTabs-flexContainer': {
              gap: '2rem',
            },
            minHeight: '48px',
          }}
        >
          <span onClick={(e) => handleTabChange(e, 0)} className={classes.tab}>
            <Tab
              label="Dashboard"
              className={classes.tab}
              sx={{
                textTransform: 'none',
                minHeight: '48px',
                padding: '12px 16px',
              }}
            />
          </span>
          <Tooltip
            title={disabled ? 'Please wait while the report is being generated.' : ''}
            arrow
            componentsProps={{
              tooltip: {
                sx: {
                  bgcolor: 'rgba(60,60,60, 0.87)',
                  '& .MuiTooltip-arrow': {
                    color: 'rgba(60, 60, 60, 0.87)',
                  },
                  padding: '8px 12px',
                  fontSize: '14px',
                },
              },
            }}
          >
            <span
              onClick={(e) => !disabled && handleTabChange(e, 1)}
              style={{
                cursor: disabled ? 'not-allowed' : 'pointer',
              }}
              className={classes.tab}
            >
              <Tab
                label="Graph Visualization"
                className={classes.tab}
                sx={{
                  textTransform: 'none',
                  minHeight: '48px',
                  padding: '12px 16px',
                  pointerEvents: 'none',
                }}
                disabled={disabled}
              />
            </span>
          </Tooltip>
        </Tabs>
      </Box>
      {/* Added black divider below tabs */}
      <Divider
        sx={{
          backgroundColor: 'black',
          height: '1px',
          width: '100%',
          opacity: 0.1,
        }}
      />
    </Box>
  );
};

export default CustomTabs;
