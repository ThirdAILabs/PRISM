from agents import Agent

from prompts.extract_entities_from_pages import (
    SinglePageResult,
    extract_entities_from_pages_prompt,
)

single_page_extractor = Agent(
    name="Single Webpage Entity Extractor",
    instructions=extract_entities_from_pages_prompt,
    model="gpt-4o",
    output_type=SinglePageResult,
)
