import React from 'react';
import { DataGrid, GridToolbar } from '@mui/x-data-grid';
import { TbReportSearch } from 'react-icons/tb';
import { reportService } from '../../../api/reports';

const AuthorCard = ({ authors }) => {
  const handleClick = async (authorId, authorName, Source) => {
    const report = await reportService.createReport({
      AuthorId: authorId,
      AuthorName: authorName,
      Source: Source,
      StartYear: 1990,
    });

    window.open('/report/' + report.Id, '_blank');
  };

  const columns = [
    {
      field: 'authorName',
      headerName: 'Author Name',
      width: 450,
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
        <span
          onClick={() => {
            handleClick(params.row.authorId, params.row.authorName, params.row.source);
          }}
          style={{
            color: 'inherit',
            textDecoration: 'none',
            cursor: 'pointer',
          }}
        >
          <TbReportSearch />
          <span style={{ marginLeft: '10px' }}>View Report</span>
        </span>
      ),
    },
  ];

  const rows = authors.map((author, index) => ({
    id: index,
    authorName: author.AuthorName,
    flagCount: author.FlagCount,
    authorId: author.AuthorId,
    source: author.Source,
  }));
  console.log('Rows in AuthorCard', rows, authors);

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
          // '& .MuiButton-root:nth-of-type(2)': {
          //   // Filter button
          //   display: 'inline-flex',
          // },
          // '& .MuiButton-root:nth-of-type(4)': { // Export button
          //   display: 'inline-flex'
          // }
        }}
      />
    );
  };

  return (
    // <div style={{ maxHeight: 800, width: 'auto' }}>
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
      sx={{
        '& .MuiDataGrid-footerContainer': {
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0px',
          minHeight: '52px',
        },
        '& .MuiTablePagination-toolbar': {
          display: 'flex',
          alignItems: 'center',
          padding: 0,
          minHeight: 'auto',
          marginTop: 0,
          marginBottom: 0,
        },
        '& .MuiTablePagination-actions': {
          display: 'flex',
          alignItems: 'center',
          margin: 0,
          padding: 0,
          height: 'auto',
          position: 'relative',
          top: '-0px',
        },
        '& .MuiTablePagination-displayedRows': {
          margin: 0,
          padding: 0,
        },
        '& .MuiTablePagination-selectLabel': {
          margin: 0,
          padding: 0,
        },
        '& .MuiIconButton-root': {
          padding: '4px',
          margin: 0,
        },
      }}
    />
    // </div>
  );
};

export default AuthorCard;
