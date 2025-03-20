import json
from datetime import datetime
import os
import time
from openai import OpenAI
import time
import requests


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

            # Stop if an article is published before the start date
            if pub_date.date() < start_date.date():
                print(
                    "[fetch_articles] Encountered an article published before the start date; stopping further fetches."
                )
                found_old_entry = True
                break

            link = entry.get("url", "")
            combined_text = entry.get("body", "")

            # Apply your keyword filters
            if not any(
                country.lower() in combined_text.lower()
                for country in config["country_keywords"]
            ):
                continue

            if not any(
                keyword.lower() in combined_text.lower()
                for keyword in config["academic_keywords"]
            ):
                continue

            article = {
                "title": title,
                "link": link,
                "pubDate": pub_date.strftime("%Y-%m-%d %H:%M:%S"),
                "article_text": combined_text,
            }
            articles.append(article)

    process_entries(data)
    if found_old_entry:
        return articles

    for page in range(1, total_pages):
        time.sleep(0.5)  # Limit to 2 requests per second
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


def process_articles(fetched_articles, config):
    print(f"[process_articles] Processing {len(fetched_articles)} articles...")
    for i, article in enumerate(fetched_articles, 1):
        print(f"[process_articles] Processing article {i}/{len(fetched_articles)}")
        article["entities"] = get_entities(article["article_text"], config)
    print("[process_articles] Completed processing all articles")
    return fetched_articles


def update_articles(processed_articles, config):
    print("[update_articles] Starting to save processed articles...")
    output_file = config["output_file"]
    os.makedirs(os.path.dirname(output_file), exist_ok=True)
    print(f"[update_articles] Saving to file: {output_file}")
    with open(output_file, "w") as outfile:
        json.dump(processed_articles, outfile, indent=4)
    print(
        f"[update_articles] Successfully saved {len(processed_articles)} articles to {output_file}"
    )
