import json
import os
import subprocess


def crawl_university_webpages(config):
    # Run your Scrapy spider for university webpages.
    output_file = config["output_file"]
    os.makedirs(os.path.dirname(output_file), exist_ok=True)
    cmd = ["scrapy", "crawl", "university", "-o", output_file, "--nolog"]
    subprocess.run(cmd, check=True)
    # Load the resulting JSON.
    with open(output_file, "r", encoding="utf-8") as f:
        data = json.load(f)
    return data


def process_university_webpages(raw_data, config):
    # Here you might perform additional processing on the scraped webpages.
    return raw_data


def update_university_webpages(processed_data, config):
    output_file = config["output_file"]
    os.makedirs(os.path.dirname(output_file), exist_ok=True)
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(processed_data, f, indent=2, ensure_ascii=False)
    print(f"[university_webpages] Updated data written to {output_file}")
