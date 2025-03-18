import requests
from scrapy.selector import Selector


def get_article_text(url):
    r = requests.get(url)
    sel = Selector(text=r.text)
    paragraphs = sel.xpath("//p//text()").getall()
    return " ".join(paragraphs)
