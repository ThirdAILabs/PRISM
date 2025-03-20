import json
from datetime import datetime
import os
import time
from openai import OpenAI
import time
import requests


def ask_gpt(message, config):
    print("[ask_gpt] Initiating GPT request...")
    client = OpenAI(api_key=config["openai_api_key"])
    while True:
        try:
            print("[ask_gpt] Sending request to GPT-4...")
            response = client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[{"role": "user", "content": message}],
                temperature=0.1,
            )
            print("[ask_gpt] Successfully received response from GPT")
            return response.choices[0].message.content
        except Exception as e:
            print(f"[ask_gpt] Error occurred: {e}")
            print("[ask_gpt] Retrying in 5 seconds...")
            time.sleep(5)


def get_entities(content, config):
    print("[get_entities] Starting entity extraction...")

    print("[get_entities] Querying GPT for institutions...")
    prompt = f"In the following article, list all NON-US INSTITUTIONS responsible for the crime. Please DO NOT list the names of countries or cities. DO NOT use abbreviations. If there is not a crime, list nothing. Please return a list of all of these CRIMINAL institutions in a newline separated list such that I can parse the response with response.split('\\n') in python. Include no special characters like '`'. Here is the article for you to examine: {content}"
    institutions = ask_gpt(prompt, config)

    print("[get_entities] Querying GPT for persons...")
    prompt = f"In the following article, list the NAMES of the individual person or persons who are RESPONSIBLE for the crime. Who are the ones at fault, not the ones after them or innocent. If there is not a crime, list nothing. Please return a list of all of these names in a newline separated list such that I can parse the response with response.split('\\n') in python. Include no special characters like '`'. Here is the article for you to examine: {content}"
    persons = ask_gpt(prompt, config)

    entities = institutions.split("\n") + persons.split("\n")
    print(f"[get_entities] Found {len(entities)} total entities")
    return entities


def fetch_articles(config):
    print("[fetch_articles] Starting article fetch process...")

    print("[fetch_articles] Starting article fetch process...")
    start_date = datetime.strptime(config["start_date"], "%Y-%m-%d")
    print(f"[fetch_articles] Start date: {start_date}")

    # API URL template without a date filter; pagesize set to 50.
    url_template = (
        "https://www.justice.gov/api/v1/press_releases.json?"
        "sort=created&direction=DESC&pagesize=50&page={page}"
    )

    articles = []

    # Fetch the first page to obtain metadata
    page = 0
    feed_url = url_template.format(page=page)
    print(f"[fetch_articles] Fetching API page {page}: {feed_url}")
    try:
        response = requests.get(feed_url, headers={"Accept": "application/json"})
        data = response.json()
    except Exception as e:
        print(f"[fetch_articles] Error fetching or parsing data: {e}")
        return articles

    # Retrieve metadata to calculate total pages
    metadata = data.get("metadata", {})
    resultset = metadata.get("resultset", {})
    total_results = int(resultset.get("count", 0))
    current_pagesize = int(resultset.get("pagesize", 50))
    total_pages = (total_results + current_pagesize - 1) // current_pagesize
    print(
        f"[fetch_articles] Total results: {total_results}, pagesize: {current_pagesize}, total_pages: {total_pages}"
    )

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
            print(f"[fetch_articles] Processing article:[{title}] from {pub_date}")

            # If the article was published before the current date, stop processing further pages
            if pub_date.date() < start_date.date():
                print(
                    "[fetch_articles] Encountered an article published before the current date; stopping further fetches."
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
                print(
                    "[fetch_articles] Skipping article - no matching country keywords"
                )
                continue

            if not any(
                keyword.lower() in combined_text.lower()
                for keyword in config["academic_keywords"]
            ):
                print(
                    "[fetch_articles] Skipping article - no matching academic keywords"
                )
                continue

            print("[fetch_articles] Article matched criteria - adding to collection")
            article = {
                "title": title,
                "link": link,
                "pubDate": pub_date.strftime("%Y-%m-%d %H:%M:%S"),
                "article_text": combined_text,
            }
            articles.append(article)

    # Process the first page
    process_entries(data)
    if found_old_entry:
        return articles

    # Loop over remaining pages, stopping if an old article is encountered.
    for page in range(1, total_pages):
        time.sleep(0.5)  # Respect rate limit: max 2 requests per second
        feed_url = url_template.format(page=page)
        print(f"[fetch_articles] Fetching API page {page}: {feed_url}")
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
        f"[fetch_articles] Completed. Found {len(articles)} articles published on the current date."
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
