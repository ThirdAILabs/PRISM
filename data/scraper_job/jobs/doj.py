import json
from datetime import datetime
import os
import time
from openai import OpenAI
import time
import requests
from bs4 import BeautifulSoup


def ask_gpt(message, config):
    client = OpenAI(api_key=config["openai_api_key"])
    retries = 0
    max_retries = 3
    while retries < max_retries:
        try:
            response = client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[{"role": "user", "content": message}],
                temperature=0.1,
            )
            return response.choices[0].message.content
        except Exception as e:
            retries += 1
            if retries == max_retries:
                raise Exception(f"[ask_gpt] Error after {max_retries} retries: {e}")
            time.sleep(5)


def get_entities(content, config):

    prompt = f"In the following article, list all NON-US INSTITUTIONS responsible for the crime. Please DO NOT list the names of countries or cities. DO NOT use abbreviations. If there is not a crime, list nothing. Please return a list of all of these CRIMINAL institutions in a newline separated list such that I can parse the response with response.split('\\n') in python. Include no special characters like '`'. Here is the article for you to examine: {content}"
    institutions = ask_gpt(prompt, config)

    prompt = f"In the following article, list the NAMES of the individual person or persons who are RESPONSIBLE for the crime. Who are the ones at fault, not the ones after them or innocent. If there is not a crime, list nothing. Please return a list of all of these names in a newline separated list such that I can parse the response with response.split('\\n') in python. Include no special characters like '`'. Here is the article for you to examine: {content}"
    persons = ask_gpt(prompt, config)

    entities = institutions.split("\n") + persons.split("\n")
    return entities


def fetch_articles(config):
    start_date = datetime.strptime(config["start_date"], "%Y-%m-%d")

    url_template = (
        "https://www.justice.gov/api/v1/press_releases.json?"
        "sort=created&direction=DESC&pagesize=50&page={page}"
    )

    articles = []
    page = 0
    feed_url = url_template.format(page=page)
    try:
        response = requests.get(feed_url, headers={"Accept": "application/json"})
        data = response.json()
    except Exception as e:
        print(f"[fetch_articles] Error fetching or parsing data: {e}")
        return articles

    metadata = data.get("metadata", {})
    resultset = metadata.get("resultset", {})
    total_results = int(resultset.get("count", 0))
    current_pagesize = int(resultset.get("pagesize", 50))
    total_pages = (total_results + current_pagesize - 1) // current_pagesize

    found_old_entry = False

    def process_entries(data):
        nonlocal found_old_entry
        for entry in data.get("results", []):
            try:
                pub_timestamp = int(entry.get("date", 0))
            except Exception as e:
                print("[fetch_articles] Invalid date in entry, skipping")
                continue

            pub_date = datetime.fromtimestamp(pub_timestamp)
            title = entry.get("title", "")

            if pub_date.date() < start_date.date():
                print(
                    "[fetch_articles] Encountered an article published before the start date; stopping further fetches."
                )
                found_old_entry = True
                break

            link = entry.get("url", "")
            try:
                content = entry.get("body", "")
                soup = BeautifulSoup(content, "html.parser")
                combined_text = soup.get_text(separator=" ", strip=True)
            except Exception as e:
                print(f"[fetch_articles] Error cleaning HTML: {e}")
                combined_text = content

            # Get matching country keywords
            country_matches = set()
            text_lower = combined_text.lower()

            for keyword_dict in config["country_keywords"]:
                for keyword, country in keyword_dict.items():
                    if keyword.lower() in text_lower:
                        country_matches.add(country)

            # Skip if no country matches
            if not country_matches:
                continue

            if not any(
                keyword.lower() in text_lower for keyword in config["academic_keywords"]
            ):
                continue

            article = {
                "title": title,
                "link": link,
                "pubDate": pub_date.strftime("%Y-%m-%d %H:%M:%S"),
                "article_text": combined_text,
                "country": list(country_matches)[0],
            }
            articles.append(article)

    process_entries(data)
    if found_old_entry:
        return articles

    for page in range(1, total_pages):
        time.sleep(
            0.5
        )  # Limit to 2 requests per second, as there is a limit of 10 requests per second per machine
        feed_url = url_template.format(page=page)
        try:
            response = requests.get(feed_url, headers={"Accept": "application/json"})
            data = response.json()
        except Exception as e:
            print(f"[fetch_articles] Error on page {page}: {e}")
            break

        process_entries(data)
        if found_old_entry:
            break

    print(
        f"[fetch_articles] Completed. Found {len(articles)} unique articles published on or after {start_date.date()}."
    )
    return articles


# TODO(pratik): Add a way to get relevant webpages here
def process_articles(fetched_articles, config):
    for i, article in enumerate(fetched_articles, 1):
        print(f"[process_articles] Processing article {i}/{len(fetched_articles)}")
        article["entities"] = get_entities(article["article_text"], config)
        article["entities_as_text"] = " ; ".join(article["entities"])
        article["relevant_webpages"] = []
    return fetched_articles


def update_articles(processed_articles, config):
    output_file_path = config["output_file_path"]
    os.makedirs(os.path.dirname(output_file_path), exist_ok=True)

    country_articles = {}
    if os.path.exists(output_file_path):
        with open(output_file_path, "r") as infile:
            try:
                country_articles = json.load(infile)
            except json.JSONDecodeError:
                country_articles = {}

    for article in processed_articles:
        country = article["country"].lower()

        if country not in country_articles:
            country_articles[country] = []

        existing_urls = {existing["link"] for existing in country_articles[country]}
        if article["link"] not in existing_urls:
            country_articles[country].append(article)

    with open(output_file_path, "w") as outfile:
        json.dump(country_articles, outfile, indent=4)

    total_articles = sum(len(articles) for articles in country_articles.values())
    print(
        f"[update_articles] Successfully saved {total_articles} articles to {output_file_path}"
    )
