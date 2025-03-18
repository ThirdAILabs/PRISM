import json
from datetime import datetime
import tqdm
from config import load_config, update_config, set_new_start_date
from rss_fetcher import fetch_articles
from entity_extractor import get_entities


def main():
    config = load_config()
    articles = fetch_articles(config)
    for article in tqdm.tqdm(articles):
        entities = get_entities(article["article_text"])
        article["entities"] = entities
    with open("doj_articles_with_content_and_entities.json", "w") as outfile:
        json.dump(articles, outfile, indent=4)
    new_start_date = datetime.now().strftime("%Y-%m-%d")
    config = set_new_start_date(config, new_start_date)
    update_config(config)


if __name__ == "__main__":
    main()
