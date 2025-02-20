# General Rules for Endpoints

Unless stated otherwise, the following apply to all endpoints.
- Access token should be passed via the `"Authorization": "Bearer <token>"` header.
- All endpoints will return status code 200 on success. 
- On errors the response body will be the text of the error message. 
- If the status code is >= 400 and < 500 then it means that the error was the result of something in the user's request. For example an expired license, report id that doesn't exist, etc. These messages may need to be relayed to the user so they use can resolve the issue. 
- If the status code is >= 500 then the error was due to a server error the user cannot control. This is an issue that we would need to look into on the backend side. We should still tell the user that the error occurred, but the message may not provide many details since the user cannot resolve the issue themselves and we don't want to leak implementation details.

# Report Endpoints

## List Reports

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/report/list` | Yes | Token for Keycloak User Realm |

Lists all reports created by the user. The user id is determined from the provided access token.

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
        "StartYear": 10,
        "EndYear": 12,
        "Status": "queued"
    },
    {
        "Id": "e42ba4dd-f56b-4916-835b-034679df2d4b",
        "CreatedAt": "2025-02-11T20:21:49.387547Z",
        "AuthorId": "author id",
        "AuthorName": "author name",
        "Source": "openalex",
        "StartYear": 3,
        "EndYear": 8,
        "Status": "in-progress"
    }
]
```

## Create a Report

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v1/report/create` | Yes | Token for Keycloak User Realm |

Create a new report. The user id is determined from the provided access token.

__Example Request__: 
```json
{
    "AuthorId": "author name",
    "AuthorName": "author id",
    "Source": "openalex",
    "StartYear": 10,
    "EndYear": 12
}
```
__Example Response__:
```json
{
    "Id": "f9589b57-4b73-409a-98b8-a97b0ca5d936"
}
```

## Get a Report

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/report/{report_id}` | Yes | Token for Keycloak User Realm |

Gets a report. The user must be the same one who create the report. 

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
    "StartYear": 3,
    "EndYear": 8,
    "Status": "in-progress",
    "Content": {

    }
}
```

## Activate a License 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v1/report/activate-license` | Yes | Token for Keycloak User Realm |

Activates a license for the given user. This will be stored so that the user can create reports in the future. If a user attempts to create a report before activating a license an error will be returned. The user id is determined from the provided access token.

__Example Request__: 

The license key should be passed in the license field of the request body.
```json
{
    "License": "V1-Ln8DAQEOTGljZW5zZVBheWxvYWQB_4AAAQIBAklkAf-CAAEGU2VjcmV0AQoAAAAQ_4EGAQEEVVVJRAH_ggAAAEf_gAEQpAie_oymRTqsRgHBKV4PZQEwwkiCShClSaNJNZM1CVazo9lzqq9Opzulu9SCfkTksIsbftR0EpK8-P4PdeVa_xbeAA=="
}
```
__Example Response__:
```
No response body
```

## Check Disclosure

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v1/report/{report_id}/check-disclosure` | Yes | Token for Keycloak User Realm |

Checks for the disclosure of flagged details within a report. The process involves scanning one or more uploaded files for text that matches tokens extracted from the report’s flag details. If a token is found in any file’s text, the corresponding flag will be marked as disclosed.

__Example Request__: 
```
url = f"http://localhost:8082/api/v1/report/{report_id}/check-disclosure"
headers = {"Authorization": f"Bearer {token}"}

files = []
for file_path in file_paths:
    files.append(('files', (file_path, open(file_path, 'rb'), 'text/plain')))

response = requests.post(url, headers=headers, files=files)
```

__Example Response__:
```
{
    "Id": "e42ba4dd-f56b-4916-835b-034679df2d4b",
    "CreatedAt": "2025-02-11T20:21:49.387547Z",
    "AuthorId": "author id",
    "AuthorName": "author name",
    "Source": "openalex",
    "StartYear": 3,
    "EndYear": 8,
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


# License Endpoints

## List licenses

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/license/list` | Yes | Token for Keycloak Admin Realm |

Lists all created licenses. 

__Example Request__: 
```
No request body
```
__Example Response__:
```json
[
    {
        "Id": "a131b6ae-503c-4792-8755-2dd713b390ba",
        "Name": "xyz",
        "Expiration": "2025-02-11T20:43:25.798785Z",
        "Deactivated": false
    },
    {
        "Id": "5b34088d-cb2d-46ac-85a8-85e6e8a325ae",
        "Name": "abc",
        "Expiration": "2025-02-11T20:43:25.900165Z",
        "Deactivated": false
    }
]
```

## Create a License

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v1/license/create` | Yes | Token for Keycloak Admin Realm |

Create a new license. The timezone of the expiration is treated as UTC. 

__Example Request__: 
```json
{
    "Name": "test-license",
    "Expiration": "2025-02-11T20:37:49.004638Z"
}
```
__Example Response__:
```json
{
    "Id": "a4089efe-8ca6-453a-ac46-01c1295e0f65",
    "License": "V1-Ln8DAQEOTGljZW5zZVBheWxvYWQB_4AAAQIBAklkAf-CAAEGU2VjcmV0AQoAAAAQ_4EGAQEEVVVJRAH_ggAAAEf_gAEQpAie_oymRTqsRgHBKV4PZQEwwkiCShClSaNJNZM1CVazo9lzqq9Opzulu9SCfkTksIsbftR0EpK8-P4PdeVa_xbeAA=="
}
```

## Deactivate a License

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v1/license/{license_id}` | Yes | Token for Keycloak Admin Realm |

Deactivates a license. This is a soft delete, the license and all associated data is preserved, but the license is marked as deactivated and cannot be used.

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
        "AuthorId": "https://openalex.org/A5108903505",
        "AuthorName": "Anshumali Shrivastava",
        "Institutions": [
            ""
        ],
        "Source": "openalex"
    },
    {
        "AuthorId": "https://openalex.org/A5024993683",
        "AuthorName": "Anshumali Shrivastava",
        "Institutions": [
            "Rice University, USA"
        ],
        "Source": "openalex"
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
        "InstitutionId": "https://openalex.org/I74775410",
        "InstitutionName": "Rice University",
        "Location": "Houston, USA"
    }
]
```

# Search Endpoints

## Search Openalex Authors

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v1/search/regular?author_name=<author name>&institution_id=<institution_id>` | Yes | Token for Keycloak User Realm |

Searches for authors on openalex matching the given name and institution id. The `institution_id` and `author_name` are passed as url query parameters. The institution id can come from the `InstitutionId` field returned from the institution autocompletion endpoint.

__Example Request__: 
```
GET http://example.com/search/regular?author_name=anshumali+shrivastava&institution_id=https%3A%2F%2Fopenalex.org%2FI74775410
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