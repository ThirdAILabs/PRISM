from __future__ import annotations

import asyncio
import os
from typing import Dict, List, Set, Tuple

from agents import Runner, gen_trace_id, trace
from pydantic import BaseModel
from rich.console import Console

from extraction_agents import (
    doj_extractor,
    normalize_markdown_agent,
    single_page_extractor,
)
from printer import Printer
from prompts import DOJArticleExtractionOutput
from prompts.doj_extraction import EntityList
from prompts.extract_entities_from_pages import SinglePageResult
from search_utils import WebPage, WebSearchResult, perform_web_search

assert os.environ["OPENAI_API_KEY"] is not None


class FinalAuxiliaryWebpage(BaseModel):
    entities: SinglePageResult
    page: WebPage


class SearchManager:
    def __init__(self):
        self.console = Console()
        self.printer = Printer(self.console)

    async def run(self, doj_press_release: str):
        trace_id = gen_trace_id()

        with trace("Runner Trace", trace_id=trace_id):
            self.printer.update_item(
                "trace_id",
                f"View trace: https://platform.openai.com/traces/{trace_id}",
                is_done=True,
                hide_checkmark=True,
            )

            self.printer.update_item(
                "starting",
                "Starting processing DOJ Press Release...",
                is_done=True,
                hide_checkmark=True,
            )

            doj_extraction_output = await self._process_doj_press_release(
                doj_press_release
            )
            url_to_webpage_and_queries: Dict[str, Tuple[WebPage, Set[str]]] = {}
            for query in doj_extraction_output.search_queries:
                web_search_result = self._scrape_webpages(query)
                if web_search_result is None:
                    continue

                for webpage in web_search_result.results:
                    if webpage.url in url_to_webpage_and_queries:
                        url_to_webpage_and_queries[webpage.url][1].add(query)
                    else:
                        url_to_webpage_and_queries[webpage.url] = (webpage, {query})
            unique_webpages = WebSearchResult(
                results=[page for page, _ in url_to_webpage_and_queries.values()]
            )

            self.printer.update_item(
                "planning",
                f"Processing {len(unique_webpages.results)} unique webpages...",
            )

            cleaned_webpages = await self._preprocess_webpages(unique_webpages)
            if cleaned_webpages is None:
                return None

            self.printer.update_item(
                "planning",
                f"Extracting auxiliary information from {len(cleaned_webpages.results)} webpages...",
            )

            auxiliary_webpages = await self._extract_auxiliary_webpages(
                cleaned_webpages, doj_extraction_output
            )

            self.printer.update_item(
                "planning",
                "Finished extracting auxiliary webpages. Completed all tasks.",
                is_done=True,
            )
            return {
                "auxiliary_webpages": [
                    auxiliary_webpage.model_dump()
                    for auxiliary_webpage in auxiliary_webpages
                ],
                "url_to_queries": {
                    url: list(queries)
                    for url, (_, queries) in url_to_webpage_and_queries.items()
                },
                "doj_extraction_output": doj_extraction_output.model_dump(),
            }

    async def _process_doj_press_release(
        self, doj_press_release: str
    ) -> DOJArticleExtractionOutput:
        self.printer.update_item(
            "starting",
            "Starting extraction of entities from DOJ Press Release...",
        )

        result = await Runner.run(doj_extractor, doj_press_release)

        self.printer.update_item(
            "completed",
            f"Completed extraction of entities from DOJ Press Release",
            is_done=True,
        )
        return result.final_output_as(DOJArticleExtractionOutput)

    def _scrape_webpages(self, search_query: str) -> WebSearchResult:
        """
        Scrapes webpages for the given search queries
        """
        self.printer.update_item(
            "planning",
            f"Scraping webpages for {search_query}...",
        )
        markdown_pages = perform_web_search(search_query)
        if markdown_pages is None:
            self.printer.update_item(
                "planning",
                f"No webpages found for {search_query}",
                is_done=True,
            )
            return None
        self.printer.update_item(
            "planning",
            f"Scraped {len(markdown_pages.results)} webpages",
            is_done=True,
        )
        return markdown_pages

    async def _preprocess_webpages(
        self, web_search_result: WebSearchResult
    ) -> WebSearchResult:
        """
        Preprocesses the webpages for the given search queries
        """
        self.printer.update_item(
            "planning",
            f"Normalizing markdown of {len(web_search_result.results)} webpages...",
        )

        if len(web_search_result.results) == 0:
            self.printer.update_item(
                "planning",
                "No webpages found",
                is_done=True,
            )
            return None

        cleaning_tasks = [
            Runner.run(normalize_markdown_agent, page.Markdown)
            for page in web_search_result.results
        ]
        cleaned_results = await asyncio.gather(*cleaning_tasks)
        cleaned_markdowns = [result.final_output_as(str) for result in cleaned_results]

        final_results = WebSearchResult(
            results=[
                WebPage(url=page.url, title=page.title, Markdown=cleaned_markdown)
                for page, cleaned_markdown in zip(
                    web_search_result.results, cleaned_markdowns
                )
            ]
        )

        self.printer.update_item(
            "planning",
            f"Normalized markdown of {len(web_search_result.results)} webpages",
            is_done=True,
        )

        return final_results

    async def _extract_auxiliary_webpages(
        self,
        web_search_result: WebSearchResult,
        doj_article_extraction_output: DOJArticleExtractionOutput,
    ) -> List[FinalAuxiliaryWebpage]:
        """
        Extracts auxiliary webpages for the given search queries
        """
        self.printer.update_item(
            "planning",
            f"Extracting entities from {len(web_search_result.results)} webpages...",
        )

        if len(web_search_result.results) == 0:
            self.printer.update_item(
                "planning",
                "No webpages found",
                is_done=True,
            )
            return []

        entity_list_json = EntityList(
            individuals=doj_article_extraction_output.individuals,
            institutions=doj_article_extraction_output.institutions,
        ).model_dump_json(indent=2)

        extraction_tasks = [
            Runner.run(
                single_page_extractor,
                f"""DOJ Press Release : 
{entity_list_json}

Webpage content : 
{page.Markdown}""",
            )
            for page in web_search_result.results
        ]
        extraction_results = await asyncio.gather(*extraction_tasks)

        final_auxiliary_webpages = [
            FinalAuxiliaryWebpage(
                entities=result.final_output_as(SinglePageResult), page=page
            )
            for result, page in zip(extraction_results, web_search_result.results)
        ]
        return final_auxiliary_webpages
