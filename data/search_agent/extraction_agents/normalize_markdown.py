from agents import Agent

from prompts.translate_to_english import translate_to_english_prompt

normalize_markdown_agent = Agent(
    name="Normalize Markdown",
    instructions=translate_to_english_prompt,
    model="gpt-4o",
)
