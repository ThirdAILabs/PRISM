import json
import feedparser
from datetime import datetime
import os
import requests
from scrapy.selector import Selector
import time
from openai import OpenAI


def get_article_text(url):
    r = requests.get(url)
    sel = Selector(text=r.text)
    paragraphs = sel.xpath("//p//text()").getall()
    return " ".join(paragraphs)


def ask_gpt(message, config):
    client = OpenAI(api_key=config["openai_api_key"])
    while True:
        try:
            response = client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[{"role": "user", "content": message}],
                temperature=0.1,
            )
            return response.choices[0].message.content
        except Exception as e:
            time.sleep(5)


def get_entities(content, config):
    prompt = f"In the following article, list all NON-US INSTITUTIONS responsible for the crime. Please DO NOT list the names of countries or cities. DO NOT use abbreviations. If there is not a crime, list nothing. Please return a list of all of these CRIMINAL institutions in a newline separated list such that I can parse the response with response.split('\\n') in python. Include no special characters like '`'. Here is the article for you to examine: {content}"
    institutions = ask_gpt(prompt)
    prompt = f"In the following article, list the NAMES of the individual person or persons who are RESPONSIBLE for the crime. Who are the ones at fault, not the ones after them or innocent. If there is not a crime, list nothing. Please return a list of all of these names in a newline separated list such that I can parse the response with response.split('\\n') in python. Include no special characters like '`'. Here is the article for you to examine: {content}"
    persons = ask_gpt(prompt, config)
    return institutions.split("\n") + persons.split("\n")


def fetch_articles(config):
    start_date = datetime.strptime(config["start_date"], "%Y-%m-%d")
    url_template = "https://www.justice.gov/news/rss?type=press_release&m={}"
    articles = []
    m = 1
    found_old_entry = False
    while True:
        feed_url = url_template.format(m)
        feed = feedparser.parse(feed_url)
        if not feed.entries:
            break
        for entry in feed.entries:
            if not hasattr(entry, "published_parsed"):
                continue
            pub_date = datetime(*entry.published_parsed[:6])
            if pub_date < start_date:
                found_old_entry = True
                break
            full_text = get_article_text(entry.get("link", ""))
            combined_text = full_text.lower()
            if not any(
                country.lower() in combined_text
                for country in config["country_keywords"]
            ):
                continue
            if not any(
                keyword.lower() in combined_text
                for keyword in config["academic_keywords"]
            ):
                continue
            article = {
                "title": entry.get("title", ""),
                "link": entry.get("link", ""),
                "pubDate": pub_date.strftime("%Y-%m-%d %H:%M:%S"),
                "article_text": full_text,
            }
            articles.append(article)
        if found_old_entry:
            break
        m += 1
    return articles


def process_articles(fetched_articles, config):
    for article in fetched_articles:
        article["entities"] = get_entities(article["article_text"], config)
    return fetched_articles


def update_articles(processed_articles, config):
    output_file = config["output_file"]
    os.makedirs(os.path.dirname(output_file), exist_ok=True)
    with open(output_file, "w") as outfile:
        json.dump(processed_articles, outfile, indent=4)
    print(f"[doj_pr] Articles with entities saved to {output_file}")
