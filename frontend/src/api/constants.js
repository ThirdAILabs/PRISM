export const prismBaseUrl = 'http://localhost:8082';

export const API_ROUTES = {
  REPORTS: {
    LIST: '/api/v1/report/author/list',
    CREATE: '/api/v1/report/author/create',
    GET: (id) => `/api/v1/report/author/${id}`,
    ACTIVATE_LICENSE: '/api/v1/report/activate-license',
    CHECK_DISCLOSURE: (id) => `/api/v1/report/author/${id}/check-disclosure`,
    DOWNLOAD: (id) => `/api/v1/report/author/${id}/download`,
  },
  LICENSES: {
    LIST: '/api/v1/license/list',
    CREATE: '/api/v1/license/create',
    DEACTIVATE: (id) => `/api/v1/license/${id}`,
  },
  AUTOCOMPLETE: {
    AUTHOR: '/api/v1/autocomplete/author',
    INSTITUTION: '/api/v1/autocomplete/institution',
  },
  SEARCH: {
    REGULAR: '/api/v1/search/regular',
    ADVANCED: '/api/v1/search/advanced',
    MATCH_ENTITIES: '/api/v1/search/match-entities',
  },
};
