import json
import scrapy
import requests
from bs4 import BeautifulSoup
from openai import OpenAI
from urllib.parse import quote


def extract_url_from_content(content, api_key):
    client = OpenAI(api_key=api_key)
    message = (
        "Extract the official homepage URL from the following HTML content. "
        "Return only the URL as a single output without any additional text or explanation:\n\n"
        f"{content}"
    )
    try:
        response = client.chat.completions.create(
            model="gpt-4o-mini",
            messages=[{"role": "user", "content": message}],
            temperature=0.1,
        )
        url = response.choices[0].message.content.strip()
        return url
    except Exception as e:
        return None


def search_homepage(query, api_key):
    query_encoded = quote(query + " official website")
    search_url = f"https://duckduckgo.com/html/?q={query_encoded}"
    response = requests.get(
        search_url, headers={"User-Agent": "Mozilla/5.0"}, timeout=10
    )
    if response.status_code == 200:
        return extract_url_from_content(response.text, api_key)
    return None


class UniversitySpider(scrapy.Spider):
    name = "university"
    custom_settings = {"DEPTH_LIMIT": 1, "LOG_ENABLED": False}

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        if not self.openai_api_key:
            self.logger.error("OPENAI_API_KEY setting not provided!")
            raise ValueError("OPENAI_API_KEY setting is required")

        if not self.input_json_path:
            self.logger.error("input_json_path setting not provided!")
            raise ValueError("input_json_path setting is required")

    def start_requests(self):
        self.logger.info("Starting to process university data...")
        with open(
            self.input_json_path,
            "r",
            encoding="utf-8",
        ) as f:
            data = json.load(f)
        records = data if isinstance(data, list) else [data]
        self.logger.info(f"Found {len(records)} records to process")

        for record in records:
            entity = record.get("title", "")
            if not entity:
                self.logger.info("Skipping record with no title")
                continue
            query = entity + " official website"
            url = search_homepage(query, self.openai_api_key)
            if url:
                self.logger.info(f"Processing entity: {entity} at URL: {url}")
                meta = {"entity": entity, "depth": 0, "download_timeout": 10}
                yield scrapy.Request(url=url, callback=self.parse_page, meta=meta)

    def parse_page(self, response):
        entity = response.meta.get("entity", "")
        self.logger.info(f"Parsing page for entity: {entity} at URL: {response.url}")

        raw_html = response.text
        soup = BeautifulSoup(raw_html, "html.parser")
        text_content = soup.get_text(separator=" ", strip=True)

        item = {
            "url": response.url,
            "title": response.xpath("//title/text()").get(default="").strip(),
            "content": text_content,
            "entity": entity,
        }
        yield item

        if response.meta.get("depth", 0) < 1:
            links = response.css("a::attr(href)").getall()
            self.logger.info(f"Found {len(links)} links on page {response.url}")
            for link in links:
                abs_link = response.urljoin(link)
                if abs_link and abs_link != response.url:
                    meta = {
                        "entity": entity,
                        "depth": response.meta.get("depth", 0) + 1,
                    }
                    yield scrapy.Request(
                        url=abs_link, callback=self.parse_page, meta=meta
                    )
