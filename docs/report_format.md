# Format of Report Content

The report content is provided as a json object containing 7 fields, one for each type of 
flag we display to the user. Here is what the object looks like (examples of the objects in each list is described next): 
```json
{
    "TalentContracts": [],

    "AssociationsWithDeniedEntities": [],

    "HighRiskFunders": [],

    "AuthorAffiliations": [],

    "PotentialAuthorAffiliations": [],

    "MiscHighRiskAssociations": [],

    "CoauthorAffiliations": []
}
```

### General Notes: 
- All flags have a message field.
- All flags that are related to information from a particular work (for example acknowledgements or listed affiliations) will have a field called `Work` that provides some information about that work. The schema for this work field is consistent accross all flags which contain it. 
- All flags have a field called `Disclosed` which indicates if that flag was disclosed by an uploaded disclosure. If no disclosure has been uploaded, this will be false.

## TalentContracts
Notes: 
- The list of entities could be empty if we detect an entity of concern but it doesn't occur on one of the lists of high risk entities.
```json
{
    "Message": "Description of flag",
    "Work": {
        "WorkId": "id for work",
        "DisplayName": "name of work",
        "WorkUrl": "url to work (if found)",
        "OaUrl": "open alex work url (if found)",
        "PublicationYear": 2018
    },
    "Entities": {
        "Entity": "Entity Name",
        "Sources": ["Name of high risk entity list the entity is found on"],
        "Aliases": ["Alias of the known entity that is matched"]
    },
    "RawAcknowledements": ["The raw text of the acknowledgement section of the work"],
    "Disclosed": false
}
```

## AssociationsWithDeniedEntities
Notes: 
- The list of entities could be empty if we detect an entity of concern but it doesn't occur on one of the lists of high risk entities. 
```json
{
    "Message": "Description of flag",
    "Work": {
        "WorkId": "id for work",
        "DisplayName": "name of work",
        "WorkUrl": "url to work (if found)",
        "OaUrl": "open alex work url (if found)",
        "PublicationYear": 2018
    },
    "Entities": {
        "Entity": "Entity Name",
        "Sources": ["Name of high risk entity list the entity is found on"],
        "Aliases": ["Alias of the known entity that is matched"]
    },
    "RawAcknowledements": ["The raw text of the acknowledgement section of the work"],
    "Disclosed": false
}
```

## HighRiskFunders
Notes:
- There are two flaggers that can produce this flag.
- The first flagger checks the funders/grants listed by openalex and compares them to known entities of concern. If this flag is produced by this flagger then the `Funders` field will be the list of concerning funders and the `RawAcknowledements` field will be empty. 
- The second flagger is the acknowledgements flagger. If the flag is produced by this flagger then the `Funders` field will contain the text of any of the matched entities that are found in government watch lists, and the `RawAcknowledements` field will contain the full text of the acknowledgements. The `Funders` list of could be empty while the `RawAcknowledements` list isn't if we detect an entity of concern but it doesn't occur on one of the lists of high risk entities. 
```json
{
    "Message": "Description of flag",
    "Work": {
        "WorkId": "id for work",
        "DisplayName": "name of work",
        "WorkUrl": "url to work (if found)",
        "OaUrl": "open alex work url (if found)",
        "PublicationYear": 2018
    },
    "Funders": ["Funder name"],
    "RawAcknowledements": ["acknowledgement text"],
    "Disclosed": false
}
```

## AuthorAffiliations
```json
{
    "Message": "Description of flag",
    "Work": {
        "WorkId": "id for work",
        "DisplayName": "name of work",
        "WorkUrl": "url to work (if found)",
        "OaUrl": "open alex work url (if found)",
        "PublicationYear": 2018
    },
    "Affiliations": ["name of affiliated university/institution"],
    "Disclosed": false
}
```

## PotentialAuthorAffiliations
```json
{
    "Message": "Description of flag",
    "University": "Name of potential university appointment",
    "UniversityUrl": "Url of university",
    "Disclosed": false
}
```

## MiscHighRiskAssociations
Notes: 
- This example is for a DOJ press release, but in theory it could be any document that implicates an individual of some wrongdoing.
- The `DocTitle` and `DocUrl` fields in the root of the object are always for the incriminating document (i.e. DOJ press release).
- The `DocTitle` and `DocUrl` fields in the Connections list are for the docs that establish links between the author and the individiual implicated by the main document.
- If the author themselves is implicated by the main document, then the connections will be empty. 
- The first connection will always be the one that links to the author, subsequent connections will link to the previous connection. The last connection will link to the main incriminating document.
- If the author and the incriminated entity are linked by coauthorship, then the `FrequentCoauthor` field will have the name of the incriminated coauthor. It will be null otherwise.
```json
{
    "Message": "Description of flag",
    "DocTitle": "Name of DOJ press release",
    "DocUrl": "Url of DOJ press release",
    "DocEntities": ["Entities in DOJ press release"],
    "EntityMentioned": "Entity mentioned in press release with link to author",
    "Connections": [
        {
            "DocTitle": "Name of doc",
            "DocUrl": "Url of doc"
        }
    ],
    "FrequentCoauthor": "Optional: name of frequent coauthor if that is the link",
    "Disclosed": false
}
```

## CoauthorAffiliations
```json
{
    "Message": "Description of flag",
    "Work": {
        "WorkId": "id for work",
        "DisplayName": "name of work",
        "WorkUrl": "url to work (if found)",
        "OaUrl": "open alex work url (if found)",
        "PublicationYear": 2018
    },
    "Coauthors": ["Name of coauthor"],
    "Affiliations": ["name of affiliated university/institution"],
    "Disclosed": false
}
```