doj_prompt = """
You are an FBI agent and you are an expert at finding out malicious actors. Given a piece of DOJ press release, you need to find out the malicious actors from the press release.
You are also an excellent investigative journalist and you need to generate some search queries given the press release that will enable you to find other malicious actors that are related to the ones in the press release.

In the following article, list all INSTITUTIONS responsible for the crime. Please DO NOT list the names of countries or cities. DO NOT use abbreviations. If there is not a crime, list nothing. Do not repeat the same entities. 

Note : Refrain from tagging against large institutions like Google or MIT. But if there is evidence that the institution is strongly associated with a bad actor, include them in the output. Note, a bad actor working for Google does not make Google malicious. On the other hand, if Huawei is found guilty of stealing trade secrets from T-Mobile, mark them as malicious. The key distinction is that the crime is committed by the institution itself, not by an individual working for the institution. Use your judgement to make this distinction.

That is the entity mentioned in the DOJ press release exercises strong influence over the bad actor. Or vice versa. We only want to find such entities. Returning empty list is a perfectly valid output if no risky entities are found in the webpage. Refrain from tagging against very large institutions of Western countries like Google, Microsoft, Apple, Harvard, DOJ, FBI, etc.

If you detect a malicious institution/individual, You also need to generate a search query that will be fired to the internet to find more information about the malicious actor. 

For example, if We Yan, a professor of Computer Science from MIT is found guilty of leaking IP to a company NuProbe in China.

Individual : 
    name : We Yan
    country : USA
    field_of_work : [Computer Science, Technology, Software, Engineering]
    affiliation : [Massachusetts Institute of Technology]

Institution : 
    name : NuProbe
    country : China
    field_of_work : [Computer Science, Technology, Software, Engineering]

Search Queries :
"We Yan Computer Science Professor"
"We Yan Nuprobe"
"NuProbe executives"
"MIT professors working for NuProbe"
"We Yan MIT lab"
"We Yan leaking IP Nuprobe Co-conspirators"
"Nuprobe Chinese American Professors"

Note : We did not add MIT in the search queries as it is a large institution and it is not the one committing the crime. 
Also we changed MIT to Massachusetts Institute of Technology in the affiliation field. 


We also generated search queries that will help us further our investigtion into the company NuProbe and professor We Yan. 
The idea is that these search queries will help us further locate other potential entities that are related to Wu Yan and NuProbe who are also risky. Note, it is not necessary to include Wu Yan in the search queries. You just want to investigate potential connections between other people and Wu Yan. You can exercise your judgement while generating the search queries. Try to generate diverse search queries that explore different aspects of the crime.


Article:
{article}
"""

from typing import List

from pydantic import BaseModel


class Individual(BaseModel):
    """
    Represents an individual person involved in a crime.

    Name is the name of the individual. Do not use abbreviations. For example if both J. Smith and John Smith are mentioned, use John Smith.

    Country is the country of the individual.

    Field of work is the field of work of the individual. Ensure to use the full name of the field of work. For example if both Computer Science and Technology are mentioned, use Computer Science and Technology. Do not use blanket words like Technology, Innovation, etc.

    Affiliation is the institution the individual is affiliated with. Ensure to use the full name of the institution.
    """

    name: str
    country: str
    field_of_work: List[str]
    affiliation: List[str]


class Institution(BaseModel):
    """
    Represents an institution involved in a crime.

    Name is the name of the institution. Do not use abbreviations. Example, use Microsoft and not MS.

    Country is the country of the institution.

    Field of work is the field of work of the institution. Ensure to use the full name of the field of work. For example if both Computer Science and Technology are mentioned, use Computer Science and Technology. Do not use blanket words like Technology, Innovation, etc.

    Refrain from tagging against very large institutions of Western countries like Google, Microsoft, Apple, etc.
    """

    name: str
    country: str
    field_of_work: List[str]


class DOJArticleExtractionOutput(BaseModel):
    """
    Represents a DOJ article.

    Individuals are the individuals involved in the crime.
    Institutions are the institutions involved in the crime.
    Search queries are the search queries that will help us further our investigation into the crime.
    """

    individuals: List[Individual]
    institutions: List[Institution]
    search_queries: List[str]


class EntityList(BaseModel):
    """
    Represents a list of entities of concern.
    """

    individuals: List[Individual]
    institutions: List[Institution]
