from agents import Agent

from prompts.doj_extraction import DOJArticleExtractionOutput, doj_prompt

doj_extractor = Agent(
    name="Department of Justice Entities of Concern Extractor",
    instructions=doj_prompt,
    model="gpt-4o",
    output_type=DOJArticleExtractionOutput,
)
