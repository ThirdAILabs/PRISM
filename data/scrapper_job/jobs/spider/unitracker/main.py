import scrapy
import json
import html


class UnitrackerSpider(scrapy.Spider):
    name = "unitracker"
    allowed_domains = ["unitracker.aspi.org.au"]
    start_urls = ["https://unitracker.aspi.org.au/"]

    custom_settings = {
        "DOWNLOAD_HANDLERS": {
            "http": "scrapy_playwright.handler.ScrapyPlaywrightDownloadHandler",
            "https": "scrapy_playwright.handler.ScrapyPlaywrightDownloadHandler",
        },
        "TWISTED_REACTOR": "twisted.internet.asyncioreactor.AsyncioSelectorReactor",
        "PLAYWRIGHT_LAUNCH_OPTIONS": {"headless": True},
        "USER_AGENT": (
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
            "AppleWebKit/537.36 (KHTML, like Gecko) "
            "Chrome/95.0.4638.69 Safari/537.36"
        ),
    }

    def start_requests(self):
        for url in self.start_urls:
            yield scrapy.Request(url, meta={"playwright": True})

    def parse(self, response):
        rows = response.xpath(
            '//table[contains(@class, "data-table")]//tr[not(ancestor::thead)]'
        )
        if not rows:
            self.logger.error("No table rows found.")
            return

        for row in rows:
            institution_td = row.xpath("./td[1]")
            if institution_td:
                title = institution_td.xpath(
                    './/a[contains(@class,"data-table__university-title")]/text()'
                ).get()
                title = title.strip() if title else ""
                permalink = institution_td.xpath(
                    './/a[contains(@class,"data-table__university-title")]/@href'
                ).get()
                if permalink and not permalink.startswith("http"):
                    permalink = response.urljoin(permalink)
                aliases = institution_td.xpath(
                    './/a[contains(@class,"data-table__aliases")]//span/text()'
                ).getall()
                aliases = ", ".join(a.strip() for a in aliases if a.strip())
            else:
                title, permalink, aliases = "", "", ""

            # --- Second column: Kind ---
            kind = row.xpath("./td[2]//text()").getall()
            kind = " ".join(t.strip() for t in kind if t.strip())

            # --- Third column: Risk ---
            risk = row.xpath("./td[3]//text()").getall()
            risk = " ".join(t.strip() for t in risk if t.strip())

            # --- Fourth column: Security credentials ---
            security_credentials = row.xpath("./td[4]//text()").getall()
            security_credentials = " ".join(
                t.strip() for t in security_credentials if t.strip()
            )

            # --- Fifth column: End-user lists ---
            end_user_lists = row.xpath("./td[5]//text()").getall()
            end_user_lists = " ".join(t.strip() for t in end_user_lists if t.strip())

            # --- Sixth column: Espionage or misconduct ---
            espionage = row.xpath("./td[6]//text()").getall()
            espionage = " ".join(t.strip() for t in espionage if t.strip())

            yield {
                "title": title,
                "permalink": permalink,
                "also_known_as": aliases,
                "kind": kind,
                "risk": risk,
                "security_credentials": security_credentials,
                "end_user_lists": end_user_lists,
                "espionage_or_misconduct": espionage,
            }
