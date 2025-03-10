# General Rules for Endpoints

Unless stated otherwise, the following apply to all endpoints.
- Access token should be passed via the `"Authorization": "Bearer <token>"` header.
- All endpoints will return status code 200 on success. 
- On errors the response body will be the text of the error message. 
- If the status code is >= 400 and < 500 then it means that the error was the result of something in the user's request. For example an expired license, report id that doesn't exist, etc. These messages may need to be relayed to the user so they use can resolve the issue. 
- If the status code is >= 500 then the error was due to a server error the user cannot control. This is an issue that we would need to look into on the backend side. We should still tell the user that the error occurred, but the message may not provide many details since the user cannot resolve the issue themselves and we don't want to leak implementation details.

# Report Endpoints

## List Author Reports

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/report/author/list` | Yes | Token for Keycloak User Realm |

Lists all author reports created by the user. The user id is determined from the provided access token.

__Example Request__: 
```
No request body
```
__Example Response__:

Note: that the field `Status` will be one of `queued`, `in-progress`, `failed`, or `complete`.
```json
[
    {
        "Id": "d8a56efd-5e92-4272-ad4a-cb8ac186539e",
        "CreatedAt": "2025-02-11T20:21:49.387032Z",
        "AuthorId": "author id",
        "AuthorName": "author name",
        "Source": "openalex",
        "Status": "queued"
    },
    {
        "Id": "e42ba4dd-f56b-4916-835b-034679df2d4b",
        "CreatedAt": "2025-02-11T20:21:49.387547Z",
        "AuthorId": "author id",
        "AuthorName": "author name",
        "Source": "openalex",
        "Status": "in-progress"
    }
]
```

## Create an Author Report

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v1/report/author/create` | Yes | Token for Keycloak User Realm |

Create a new author report. The user id is determined from the provided access token.

__Example Request__: 
```json
{
    "AuthorId": "author name",
    "AuthorName": "author id",
    "Source": "openalex"
}
```
__Example Response__:
```json
{
    "Id": "f9589b57-4b73-409a-98b8-a97b0ca5d936"
}
```

## Get an Author Report

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/report/author/{report_id}` | Yes | Token for Keycloak User Realm |

Gets an author report. The user must be the same one who create the report. 

__Example Request__: 
```
No request body
```
__Example Response__:

Note: See the `report_format.md` for a description of the format of the report content.

```json
{
    "Id": "e42ba4dd-f56b-4916-835b-034679df2d4b",
    "CreatedAt": "2025-02-11T20:21:49.387547Z",
    "AuthorId": "author id",
    "AuthorName": "author name",
    "Source": "openalex",
    "Status": "in-progress",
    "Content": {

    }
}
```

## Delete an Author Report

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v1/report/author/{report_id}` | Yes | Token for Keycloak User Realm |

Deletes an author report. The user must be the same one who create the report. 

__Example Request__: 
```
No request body
```
__Example Response__:
```
No response body
```

## Check Disclosure

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v1/report/author/{report_id}/check-disclosure` | Yes | Token for Keycloak User Realm |

Checks for the disclosure of flagged details within a report. The process involves scanning one or more uploaded files for text that matches tokens extracted from the report’s flag details. If a token is found in any file’s text, the corresponding flag will be marked as disclosed.

__Example Request__: 
```
Content-Type: multipart/form-data; boundary=----WebKitFormBoundaryXYZ

------WebKitFormBoundaryXYZ
Content-Disposition: form-data; name="files"; filename="document.txt"
Content-Type: text/plain

This is the content of the document...
------WebKitFormBoundaryXYZ--
```

__Example Response__:
```
{
    "Id": "e42ba4dd-f56b-4916-835b-034679df2d4b",
    "CreatedAt": "2025-02-11T20:21:49.387547Z",
    "AuthorId": "author id",
    "AuthorName": "author name",
    "Source": "openalex",
    "Status": "complete",
    "Content": {
        "name": "test",
        "risk_score": 10,
        "connections": [],
        "type_to_flag": {
            "doj_press_release_eoc": [ /* flags with updated disclosure statuses */ ],
            "oa_coauthor_affiliation_eoc": [ /* flags */ ],
            "uni_faculty_eoc": [ /* flags */ ],
            "oa_acknowledgement_eoc": [ /* flags */ ],
            "oa_author_affiliation_eoc": [ /* flags */ ],
            "oa_funder_eoc": [ /* flags */ ]
        }
    }
}

```

## Download Report

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v1/report/author/{report_id}/download` | Yes | Token for Keycloak User Realm |

Downloads a completed report in the requested format. The report status must be complete in order for the report to be downloadable.

format (optional): Specifies the file format for the report.
Allowed values:
- csv (default): CSV (Comma-Separated Values) format.
- pdf: PDF (Portable Document Format).
- excel or xlsx: Excel file in XLSX format.


__Example Request__: 
**Status Code:** 200

**Response Headers:**

- **Content-Type:**
  - `text/csv` for CSV.
  - `application/pdf` for PDF.
  - `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` for Excel.
- **Content-Disposition:**
  - `attachment; filename="report.<ext>"` where `<ext>` is `csv`, `pdf`, or `xlsx` depending on the format.
- **Cache-Control:**
  - `no-store`

**Response Body:**

The body contains the raw bytes of the generated report file.

## List University Reports

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/report/university/list` | Yes | Token for Keycloak User Realm |

Lists all university reports created by the user. The user id is determined from the provided access token.

__Example Request__: 
```
No request body
```
__Example Response__:

Note: that the field `Status` will be one of `queued`, `in-progress`, `failed`, or `complete`.
```json
[
    {
        "Id": "e42ba4dd-f56b-4916-835b-034679df2d4b",
        "CreatedAt": "2025-02-11T20:21:49.387547Z",
        "UniversityId": "university id",
        "UniversityName": "university name",
        "Status": "complete",
    },
    {
        "Id": "e42ba4dd-f56b-4916-835b-034679df2d4b",
        "CreatedAt": "2025-02-11T20:21:49.387547Z",
        "UniversityId": "university id",
        "UniversityName": "university name",
        "Status": "complete",
    }
]
```

## Create a University Report

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v1/report/university/create` | Yes | Token for Keycloak User Realm |

Create a new university report. The user id is determined from the provided access token.

__Example Request__: 
```json
{
    "UniversityId": "university name",
    "UniversityName": "university id",
}
```
__Example Response__:
```json
{
    "Id": "f9589b57-4b73-409a-98b8-a97b0ca5d936"
}
```

## Get a University Report

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/report/university/{report_id}` | Yes | Token for Keycloak User Realm |

Gets an university report. The user must be the same one who create the report. 

__Example Request__: 
```
No request body
```
__Example Response__:

Notes: 
- The report flags will update as the author reports are completed (if not already complete).
- The `TotalAuthors` and `AuthorsReviewed` fields in the content give the total number of authors detected at the university that need review and the number that have been reviewed so far.
- The `FlagCount` for each author entry in each flag gives the number of flags of that type for the given author. 
- To get an author report from the results you can simply pass the `AuthorId`, `AuthorName`, and `Source` fields to the `/report/author/create` endpoint. The author report will be cached, so it will be displayed immediately after creation. 
```json
{
    "Id": "e42ba4dd-f56b-4916-835b-034679df2d4b",
    "CreatedAt": "2025-02-11T20:21:49.387547Z",
    "UniversityId": "university id",
    "UniversityName": "university name",
    "Status": "complete",
    "Content": {
        "TotalAuthors": 4,
        "AuthorsReviewed": 3,
        "Flags": {
            "AssociationsWithDeniedEntities": [
                {
                    "AuthorId": "author id",
                    "AuthorName": "author name",
                    "Source": "source",
                    "FlagCount": 2
                }
            ],
            "AuthorAffiliations": [
                {
                    "AuthorId": "author id",
                    "AuthorName": "author name",
                    "Source": "source",
                    "FlagCount": 2
                }
            ],
            "CoauthorAffiliations": [
                {
                    "AuthorId": "author id",
                    "AuthorName": "author name",
                    "Source": "source",
                    "FlagCount": 2
                }
            ],
            "HighRiskFunders": [
                {
                    "AuthorId": "author id",
                    "AuthorName": "author name",
                    "Source": "source",
                    "FlagCount": 2
                }
            ],
            "MiscHighRiskAssociations": [
                {
                    "AuthorId": "author id",
                    "AuthorName": "author name",
                    "Source": "source",
                    "FlagCount": 2
                }
            ],
            "PotentialAuthorAffiliations": [
                {
                    "AuthorId": "author id",
                    "AuthorName": "author name",
                    "Source": "source",
                    "FlagCount": 2
                }
            ],
            "TalentContracts": [
                {
                    "AuthorId": "author id",
                    "AuthorName": "author name",
                    "Source": "source",
                    "FlagCount": 2
                }
            ]
        }
    }
}
```

## Delete a University Report

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v1/report/university/{report_id}` | Yes | Token for Keycloak User Realm |

Deletes a univeristy report. The user must be the same one who create the report. 

__Example Request__: 
```
No request body
```
__Example Response__:
```
No response body
```

# Autocomplete Endpoints

## Autocomplete Authors

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/autocomplete/author?query=<start of author name>` | Yes | Token for Keycloak User Realm |

Generates autocompletion suggestions for the given author name. The author name is specified in the `query` url query parameter.

__Example Request__: 
```
GET http://example.com/autocomplete/author?query=anshumali+shriva
```
__Example Response__:
```json
[
    {
        "Id": "https://openalex.org/A5108903505",
        "Name": "Anshumali Shrivastava",
        "Hint": "",
    },
    {
        "Id": "https://openalex.org/A5024993683",
        "Name": "Anshumali Shrivastava",
        "Hint": "Rice University, USA"
    }
]
```

## Autocomplete Institutions

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/autocomplete/institution?query=<start of institution name>` | Yes | Token for Keycloak User Realm |

Generates autocompletion suggestions for the given institution name. The institution name is specified in the `query` url query parameter.

__Example Request__: 
```
GET http://example.com/autocomplete/institution?query=rice+univer
```
__Example Response__:
```json
[
    {
        "Id": "https://openalex.org/I74775410",
        "Name": "Rice University",
        "Hint": "Houston, USA"
    }
]
```

## Autocomplete Papers

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/autocomplete/paper?query=<start of paper title>` | Yes | Token for Keycloak User Realm |

Generates autocompletion suggestions for the given paper title. The tile query is specified in the `query` url query parameter.

__Example Request__: 
```
GET http://example.com/autocomplete/paper?query=From+Research+to+Production%3A+Towards+Scalable+and+Sustainable+Neural+Recommendation
```
__Example Response__:
```json
[
    {
        "Id": "https://openalex.org/W4386729297",
        "Name": "From Research to Production: Towards Scalable and Sustainable Neural Recommendation Models on Commodity CPU Hardware",
        "Hint": "Anshumali Shrivastava, Vihan Lakshman, Tharun Medini, et al."
    }
]
```

# Search Endpoints

## Search Authors

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/search/authors` | Yes | Token for Keycloak User Realm |

Searches for authors on openalex using one of the following filters. 
1. Author Name: Must specify the only following query parameters `author_name`, `institution_id`, and `institution_name`. The institution id can come from the `InstitutionId` field returned from the institution autocompletion endpoint.
2. ORCID: Must specify only the `orcid` query parameter. Will return a single author for the given orcid, or 404 if no author is found.
3. Paper Title: Must specify only the `paper_title` query parameter. Will return the authors for the given paper, or 404 if no paper matching the title is found.

__Example Request__: 
```
GET http://example.com/search/authors?author_name=anshumali+shrivastava&institution_id=https%3A%2F%2Fopenalex.org%2FI74775410&institution_name=Rice+University

GET http://example.com/search/authors?orcid=0000-0002-5042-2856

GET http://example.com/search/authors?paper_title=From+Research+to+Production%3A+Towards+Scalable+and+Sustainable+Neural+Recommendation+Models+on+Commodity+CPU+Hardware
```
__Example Response__:
```json
[
    {
        "AuthorId": "https://openalex.org/A5024993683",
        "AuthorName": "Anshumali Shrivastava",
        "Institutions": [
            "Rice University",
            "Third Way",
            "Amazon (United States)",
            "Search",
            "University of Houston",
            "Duke University",
            "Cornell University"
        ],
        "Source": "openalex"
    }
]
```

## Search Google Scholar Authors

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/search/advanced?query=<query>&cursor=<cursor>` | Yes | Token for Keycloak User Realm |

Searches for authors matching the given query on google scholar. The cursor url query parameter is optional. If provided it allows for the query to return the next page of results after a first page. The reponse object contains a cursor that can be passed to the next query. 

__Example Request__: 
```
GET http://example.com/search/advanced?query=anshumali+shrivastava
```
__Example Response__:
```json
{
    "Authors": [
        {
            "AuthorId": "SGT23RAAAAAJ",
            "AuthorName": "Anshumali Shrivastava",
            "Institutions": [
                "Rice University",
                "ThirdAI Corp."
            ],
            "Source": "google-scholar"
        }
    ],
    "Cursor": "eyJWMUN1cnNvciI6bnVsbCwiVjJDdXJzb3IiOjIwLCJTZWVuIjpbIlNHVDIzUkFBQUFBSiJdfQ=="
}
```

## Match Entities

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/search/match-entities?query=<searched entity name>` | Yes | Token for Keycloak User Realm |

Attempts to match the given query against known entities of concern. Returns a list of possible matches. The query is specified in the `query` url query parameter.

__Example Request__: 
```
GET http://example.com/search/match-entities?query=xyz
```
__Example Response__:
```json
{
    "Entities": [
        "institute of xyz"
    ]
}
```