import fitz
import io
import json
import os
import re
import requests

from bs4 import BeautifulSoup
from urllib.parse import urljoin


def contains_chinese(text):
    return bool(re.search(r"[\u4e00-\u9fff]", text))


def get_pdf_url_from_page(config):
    page_url = config["cset_url"]
    try:
        response = requests.get(page_url)
        response.raise_for_status()
    except requests.RequestException as e:
        print(f"Error fetching the page: {e}")
        return None

    soup = BeautifulSoup(response.text, "html.parser")
    link_tags = soup.find_all("a", href=True)

    for tag in link_tags:
        href = tag["href"]
        if href.endswith(".pdf"):
            print(
                f"[update_entities_with_cset] Found pdf with url: {urljoin(page_url, href)}"
            )
            return urljoin(page_url, href)

    print(f"[update_entities_with_cset] No PDF link found on the page {page_url}")
    return None


def fetch_source(config):
    pdf_url = get_pdf_url_from_page(config)
    response = requests.get(pdf_url)
    response.raise_for_status()

    return io.BytesIO(response.content)


def get_talent_contracts_from_pdf(pdf_data, config):
    doc = fitz.open(stream=pdf_data, filetype="pdf")
    talent_contracts = ""

    for page_num, page in enumerate(doc):
        blocks = page.get_text("dict")["blocks"]

        for block in blocks:
            if "lines" in block:
                for line in block["lines"]:
                    combined_text = ""
                    max_size = 0
                    is_bold = False

                    for span in line["spans"]:
                        text = span["text"]
                        font = span["font"]
                        size = span["size"]
                        bold = "Bold" in font or "bold" in font.lower()

                        combined_text += text
                        max_size = max(max_size, size)
                        is_bold = is_bold or bold

                    combined_text = combined_text.strip()

                    if (
                        is_bold
                        and max_size > 14
                        and len(combined_text) > 1
                        and combined_text != "Chinese Talent Program Tracker"
                    ):
                        if not contains_chinese(combined_text):
                            talent_contracts += f"{combined_text} "
                        else:
                            talent_contracts += f"({combined_text})\n"

    return talent_contracts


def update_json_file(talent_contracts, config):
    data = []
    for line in talent_contracts.splitlines():
        data.append(
            {
                "Names": line.strip(),
                "Country": "China",
                "Type": "Foreign Talent Program",
                "Resource": "Center for Security and Emerging Technologies (CSET) Chinese Talent Program Tracker https://chinatalenttracker.cset.tech",
            }
        )

    output_file_path = config["output_file_path"]
    os.makedirs(os.path.dirname(output_file_path), exist_ok=True)

    if os.path.exists(output_file_path):
        with open(output_file_path, "r", encoding="utf-8") as f:
            try:
                existing_data = json.load(f)
            except json.JSONDecodeError:
                existing_data = []
    else:
        existing_data = []

    existing_names = {entry["Names"] for entry in existing_data}

    unique_new_data = [entry for entry in data if entry["Names"] not in existing_names]

    existing_data.extend(unique_new_data)
    with open(output_file_path, "w", encoding="utf-8") as f:
        json.dump(existing_data, f, indent=4, ensure_ascii=False)

    print(
        f"[update_entities_with_cset] Appended {len(unique_new_data)} unique entries to {output_file_path}. Total entries now: {len(existing_data)}"
    )


# if __name__ == "__main__":
#     get_pdf_url_from_page("https://chinatalenttracker.cset.tech")
