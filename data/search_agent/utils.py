import re

import html2text
import requests
from bs4 import BeautifulSoup
from readability import Document


def clean_html(url):
    headers = {
        "User-Agent": (
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
        ),
        "Accept": (
            "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"
        ),
        "Accept-Language": "en-US,en;q=0.5",
        "Connection": "keep-alive",
    }
    response = requests.get(url, headers=headers, timeout=15)
    response.raise_for_status()

    soup = BeautifulSoup(response.text, "html.parser")

    # Remove common non-content elements
    for element in soup.find_all(
        [
            "style",
            "script",
            "noscript",
            "iframe",
            "head",
            "meta",
            "footer",
            "header",
            "nav",
            "button",
            "aside",
            "form",
            "select",
            "option",
        ]
    ):
        element.decompose()

    # Remove cookie notices and privacy policy elements (common patterns)
    for element in soup.find_all(
        lambda tag: tag.string
        and re.compile(r"cookie|privacy policy|terms of use", re.I).search(
            tag.get_text()
        )
    ):
        element.decompose()

    # Remove elements with common navigation/menu related text
    for element in soup.find_all(
        lambda tag: tag.string
        and re.compile(
            r"back to|main menu|navigation|corporate|worldwide", re.I
        ).search(tag.get_text())
    ):
        element.decompose()

    # Remove empty or near-empty nested structures
    for element in soup.find_all(["div", "span"]):
        text_content = element.get_text(strip=True)
        if not text_content or len(text_content) < 5:
            element.decompose()

    # Unwrap unnecessary nested tags
    for tag in soup.find_all(["div", "span"]):
        if len(tag.contents) == 1:
            tag.unwrap()

    # Keep only essential attributes
    for tag in soup.find_all(True):
        allowed_attrs = {"a": ["href"], "img": ["src", "alt"]}
        if tag.name in allowed_attrs:
            attrs = dict(tag.attrs)
            for attr in attrs:
                if attr not in allowed_attrs[tag.name]:
                    del tag.attrs[attr]
        else:
            tag.attrs = {}

    # Remove javascript: links
    for a in soup.find_all("a", href=re.compile(r"^javascript:")):
        a.decompose()

    # Pretty print with minimal indentation
    cleaned_html = soup.prettify(formatter="minimal")

    # Remove excessive blank lines and spaces
    cleaned_html = re.sub(r"\n\s*\n", "\n", cleaned_html)
    cleaned_html = re.sub(r">\s+<", ">\n<", cleaned_html)

    return cleaned_html


def convert_html_to_markdown(html):
    h = html2text.HTML2Text()
    h.ignore_links = True
    h.ignore_images = True
    return h.handle(html)
