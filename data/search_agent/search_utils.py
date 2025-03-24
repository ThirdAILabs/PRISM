import mimetypes
import os
from concurrent.futures import ThreadPoolExecutor, as_completed
from typing import List, Optional
from urllib.parse import urlparse

import requests
from pydantic import BaseModel

from utils import clean_html, convert_html_to_markdown


class WebPage(BaseModel):
    url: str
    title: str
    Markdown: str


class WebSearchResult(BaseModel):
    results: List[WebPage]


def clean_webpage_content(url):
    """
    Cleans the html of the webpage and return the cleaned data in markdown format
    """
    try:
        cleanhtml = clean_html(url)
        readable = convert_html_to_markdown(cleanhtml)
        return readable
    except Exception as e:
        print(f"Error cleaning/fetching webpage content: {e}")
        return None


def google_search(query, k=5):
    """
    Performs a google search and returns the results the top k urls along with the title and snippet
    """
    api_key = os.environ["GOOGLE_API_KEY"]
    cx = os.environ["GOOGLE_CX_CODE"]
    url = "https://www.googleapis.com/customsearch/v1"
    params = {"key": api_key, "cx": cx, "q": query, "num": k}
    response = requests.get(url, params=params)
    return response.json()


def is_acceptable_url(url: str) -> bool:
    """Check if URL is likely to be a simple webpage"""
    # Check file extension
    parsed = urlparse(url)
    ext = mimetypes.guess_extension(parsed.path)

    # List of blocked extensions
    blocked_extensions = {
        ".pdf",
        ".doc",
        ".docx",
        ".ppt",
        ".pptx",
        ".xls",
        ".xlsx",
        ".zip",
        ".rar",
        ".mp4",
        ".mp3",
        ".avi",
        ".mov",
        ".jpg",
        ".jpeg",
        ".png",
        ".gif",
    }

    if ext in blocked_extensions:
        print(f"Skipping {url}: blocked file type {ext}")
        return False

    return True


def fetch_and_clean_webpage(result: dict) -> Optional[WebPage]:
    """Helper function to fetch and clean a single webpage"""

    url = result["link"]
    if not is_acceptable_url(url):
        return None

    try:
        head_response = requests.head(url, timeout=5, allow_redirects=True)
        content_type = head_response.headers.get("content-type", "").lower()
        if not ("text/html" in content_type or "application/xhtml+xml" in content_type):
            print(f"Skipping {url}: non-HTML content type {content_type}")
            return None

        markdown = clean_webpage_content(url)

        if len(markdown) > 200_000:
            print(f"Skipping webpage {url} because it is too long")
            return None

        if markdown is not None:
            return WebPage(url=url, title=result["title"], Markdown=markdown)
    except Exception as e:
        print(f"Error cleaning/fetching webpage content: {e}")
    return None


def perform_web_search(query: str, max_workers: int = 5) -> Optional[WebSearchResult]:
    """
    Performs a web search and returns the results in parallel
    """
    try:
        results = google_search(query)
    except Exception as e:
        print(f"Error performing web search: {e}")
        return None

    websearch_results = WebSearchResult(results=[])

    # Process webpage fetching in parallel
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        # Create futures for all webpage fetches
        try:
            future_to_result = {
                executor.submit(fetch_and_clean_webpage, result): result
                for result in results["items"]
            }
        except Exception as e:
            print(f"Error creating futures: {e}")
            return None

        # Process completed futures as they finish
        for future in as_completed(future_to_result):
            webpage = future.result()
            if webpage is not None:
                websearch_results.results.append(webpage)

    if len(websearch_results.results) == 0:
        return None

    return websearch_results
