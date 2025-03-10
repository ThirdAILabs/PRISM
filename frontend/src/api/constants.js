const runtimeConfig = window._env_ || {};

export const prismBaseUrl =
  runtimeConfig.REACT_APP_BACKEND_ORIGIN || process.env.REACT_APP_BACKEND_ORIGIN;

export const API_ROUTES = {
  REPORTS: {
    LIST: '/api/v1/report/author/list',
    CREATE: '/api/v1/report/author/create',
    GET: (id) => `/api/v1/report/author/${id}`,
    ACTIVATE_LICENSE: '/api/v1/report/activate-license',
    CHECK_DISCLOSURE: (id) => `/api/v1/report/author/${id}/check-disclosure`,
    DOWNLOAD: (id) => `/api/v1/report/author/${id}/download`,
  },
  UNIVERSITY_REPORTS: {
    LIST: '/api/v1/report/university/list',
    CREATE: '/api/v1/report/university/create',
    GET: (id) => `/api/v1/report/university/${id}`,
  },
  LICENSES: {
    LIST: '/api/v1/license/list',
    CREATE: '/api/v1/license/create',
    DEACTIVATE: (id) => `/api/v1/license/${id}`,
  },
  AUTOCOMPLETE: {
    AUTHOR: '/api/v1/autocomplete/author',
    INSTITUTION: '/api/v1/autocomplete/institution',
    PAPER_TITLE: '/api/v1/autocomplete/paper',
  },
  SEARCH: {
    AUTHOR: '/api/v1/search/authors',
    AUTHOR_ADVANCED: '/api/v1/search/authors-advanced',
    MATCH_ENTITIES: '/api/v1/search/match-entities',
  },
};
