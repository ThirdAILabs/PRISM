import feedparser
from datetime import datetime
from scrapper import get_article_text


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
