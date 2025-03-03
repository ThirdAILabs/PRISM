import React from 'react';
import { DataGrid, GridToolbar } from '@mui/x-data-grid';
import { TbReportSearch } from 'react-icons/tb';

const AuthorCard = ({ authors }) => {
  const columns = [
    {
      field: 'authorName',
      headerName: 'Author Name',
      flex: 1,
    },
    {
      field: 'source',
      headerName: 'Source',
      flex: 1,
    },
    {
      field: 'flagCount',
      headerName: 'Flag Count',
      width: 230,
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 230,
      renderCell: (params) => (
        <a
          href={`/report/${params.row.reportId}`}
          target="_blank"
          rel="noopener noreferrer"
          style={{ color: 'inherit', textDecoration: 'none' }}
        >
          <TbReportSearch />
          <span style={{ marginLeft: '8px' }}>View Report</span>
        </a>
      ),
    },
  ];

  const rows = authors.map((author, index) => ({
    id: index,
    authorName: author.AuthorName,
    source: author.Source,
    flagCount: author.FlagCount,
    reportId: author.reportId,
  }));

  const handlePaginationList = () => {
    const pageSizeOptionsList = [5];
    if (authors.length > 5) {
      pageSizeOptionsList.push(10);
    }
    if (authors.length > 10) {
      pageSizeOptionsList.push(25);
    }
    return pageSizeOptionsList;
  };

  const CustomToolbar = () => {
    return (
      <GridToolbar
        showQuickFilter={true}
        quickFilterProps={{ debounceMs: 500 }}
        sx={{
          '& .MuiButton-root': {
            display: 'none',
          },
          '& .MuiButton-root:nth-of-type(2)': {
            // Filter button
            display: 'inline-flex',
          },
          // '& .MuiButton-root:nth-of-type(4)': { // Export button
          //   display: 'inline-flex'
          // }
        }}
      />
    );
  };

  return (
    <div style={{ maxHeight: 800, width: '1000px' }}>
      <DataGrid
        rows={rows}
        columns={columns}
        slots={{ toolbar: CustomToolbar }}
        density="comfortable"
        initialState={{
          pagination: {
            paginationModel: { pageSize: 10 },
          },
        }}
        pageSizeOptions={handlePaginationList()}
        disableRowSelectionOnClick
      />
    </div>
  );
};

export default AuthorCard;
