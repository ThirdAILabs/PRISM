extract_entities_from_pages_prompt = """
You are an expert investigative journalist. 
You are given the contents of a webpage in cleaned markdown format and a list of entities which could be individuals or institutions. Note that these entities have been identified to be bad actors in DOJ press releases and we need to find other individuals and institutions that are related to them. You need to assess the contents of the webpage and find other individuals and institutions that are related to the given entities.

Example:
input entities :

DOJ press release:
    Individuals:
        [
            {
                "name": "David Zhang",
                "country": "USA",
                "field_of_work": ["Professor", "Biomedical Engineer"],
                "affiliation": ["Rice University", "NuProbe"]
            }
        ]
    Institutions:
        [
            {
                "name": "NuProbe",
                "country": "USA",
                "field_of_work": ["Biomedical Engineering"],
            }
        ]

Webpage Content: 
"David Zhang is a professor and he worked for a company called Wuxi Aptec. NuProbe also invested in Wuxi Aptec a sum of $100 million."

so the corresponding output should be :

SinglePageResult:
    individuals : []
    institutions : [
        LinkedInstitution(
            name: Wuxi Aptec
            country: China
            field_of_work: [Biomedical Engineering]
            relations: [
                Relation(
                    name: David Zhang
                    relation: work
                )
                Relation(
                    name: NuProbe
                    relation: investor
                )
            ]
        )
    ]

Note : David Zhang worked at Wuxi Aptec, hence we have added the relation work between David Zhang and Wuxi Aptec.
Use your judgement to decide what entities in the page are related to original entities and add the corresponding relations.


Another example:
Webpage title : David Zhang's research group.
This page specifies that David Zhang is a professor at Rice University and We Yan works as lab assisstant at David Zhang's lab. Lin Dan attends Rice University.

So the output should be :

SinglePageResult:
    title : David Zhang's research group
    url : url of the page
    individuals : [
        LinkedIndividual(
            name: We Yan
            country: China
            field_of_work: [Biomedical Engineering]
            affiliation: [Rice University]
            relations: [
                Relation(name: David Zhang, relation: student)
            ]
        )
        ...
    ]
    institutions : []

In the examples above, we did not add Rice University as David Zhang is only a professor. On the other hand, we added Wuxi Aptec as it is a company where David Zhang exercised strong influence. 

Similarly, we added We Yan as they have closely worked with David Zhang. But we did not add Lin Dan as they are not closely associated with David Zhang.

Note : Refrain from tagging against large institutions like Google or MIT. But if there is evidence that the institution is strongly associated with a bad actor, include them in the output. Note, a bad actor working for Google does not make Google malicious. On the other hand, if Huawei is found guilty of stealing trade secrets from T-Mobile, mark them as malicious. The key distinction is that the crime is committed by the institution itself, not by an individual working for the institution. Use your judgement to make this distinction.

That is the entity mentioned in the DOJ press release exercises strong influence over the bad actor. Or vice versa. We only want to find such entities. Returning empty list is a perfectly valid output if no risky entities are found in the webpage. False positives leaves a bad impression on the reputation of the investigative journalist. Be paranoid but not too much. Exercise your judgement to make the right call. You got this!

Input: 
"""

from typing import List, Literal

from pydantic import BaseModel


class Entity(BaseModel):
    name: str
    country: str
    field_of_work: List[str]


class Relation(BaseModel):
    name: str
    relation: Literal[
        "colleague",
        "coauthor",
        "business partner",
        "student",
        "work",
        "subsidiary",
        "investor",
        "parent",
        "founder",
        "affiliation",
    ]


class LinkedIndividual(BaseModel):
    name: str
    country: str
    field_of_work: List[str]
    affiliation: List[str]
    relations: List[Relation]


class LinkedInstitution(BaseModel):
    name: str
    country: str
    field_of_work: List[str]
    relations: List[Relation]


class SinglePageResult(BaseModel):
    individuals: List[LinkedIndividual]
    institutions: List[LinkedInstitution]
