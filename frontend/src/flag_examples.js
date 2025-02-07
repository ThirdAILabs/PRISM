export const examples = [
    {
        "flagger_id": "oa_acknowledgement_eoc",
        "title": "Acknowledgements are entities of concern",
        "message": "This is a message",
        "work_title": "This is the title of a bad bad work",
        "work_url": "https://www.google.com",
        "metadata": {
            "entities": [
                {
                    "entity": "entity 1",
                    "lists": ["list 1", "list 2"],
                    "aliases": [
                        "alias 1",
                        "alias 2",
                        "alias 3"
                    ]
                },
                {
                    "entity": "entity A",
                    "lists": ["list 1", "list 2"],
                    "aliases": [
                        "alias A",
                        "alias B",
                        "alias C"
                    ]
                }
            ],
            "raw_acknowledgements": [
                "Raw ack flag 1",
                "Raw ack flag 2"
            ]
        }
    },
    {
        "flagger_id": "oa_multi_affil",
        "title": "Multiple Affiliations",
        "message": "This is a message",
        "work_title": "This is the title of a bad bad work",
        "work_url": "https://www.google.com",
        "metadata": {
            "author": "Author name",
            "affiliations": [
                "affiliation 1",
                "affiliation 2"
            ]
        }
    },
    {
        "flagger_id": "oa_funder_eoc",
        "title": "Funders are entities of concern",
        "message": "This is a message",
        "work_title": "This is the title of a bad bad work",
        "work_url": "https://www.google.com",
        "metadata": {
            "funders": [
                "Funder A",
                "Funder B",
                "Funder C"
            ]
        }
    },
    {
        "flagger_id": "oa_coauthor_eoc",
        "title": "Co-authors are entities of concern",
        "message": "This is a message",
        "work_title": "This is the title of a bad bad work",
        "work_url": "https://www.google.com",
        "metadata": {
            "coauthors": [
                "coauthor 1",
                "coauthor 2"
            ]
        }
    },
    {
        "flagger_id": "oa_affiliation_eoc",
        "title": "Affiliations are entities of concern",
        "message": "This is a message",
        "work_title": "This is the title of a bad bad work",
        "work_url": "https://www.google.com",
        "metadata": {
            "authors": [
                "author 1",
                "author 2",
                "author 3"
            ],
            "affiliations": [
                "affil 1",
                "affil 2",
                "affil 3"
            ]
        }
    },
    {
        "flagger_id": "oa_coauthor_affiliation_eoc",
        "title": "Coauthor affiliations are entities of concern",
        "message": "This is a message",
        "work_title": "This is the title of a bad bad work",
        "work_url": "https://www.google.com",
        "metadata": {
            "authors": [
                "author 1",
                "author 2",
                "author 3"
            ],
            "affiliations": [
                "affil 1",
                "affil 2",
                "affil 3"
            ]
        }
    },
    {
        "flagger_id": "oa_author_affiliation_eoc",
        "title": "author affiliations are entities of concern",
        "message": "This is a message",
        "work_title": "This is the title of a bad bad work",
        "work_url": "https://www.google.com",
        "metadata": {
            "affiliations": [
                "affil 1",
                "affil 2",
                "affil 3"
            ]
        }
    }
]