import os
import json
import subprocess
import json
from dvc_utils import dvc_read_json, dvc_write_json
import os


def crawl_university_data(config):
    intermediate_json_path = config["intermediate_json_path"]

    (
        os.remove(intermediate_json_path)
        if os.path.exists(intermediate_json_path)
        else None
    )
    os.makedirs(os.path.dirname(intermediate_json_path), exist_ok=True)

    cwd = os.path.join(os.path.dirname(__file__), "spider", "unitracker")

    cmd = ["scrapy", "runspider", "main.py", "-o", intermediate_json_path]
    subprocess.run(cmd, cwd=cwd, check=True)

    with open(intermediate_json_path, "r", encoding="utf-8") as f:
        data = json.load(f)
    return data


def process_university_data(raw_data, config):
    return raw_data


def update_university_store(processed_data, config):
    final_store_path = config["final_store_path"]
    added_store_path = config["added_store_path"]

    try:
        existing = dvc_read_json(final_store_path)
    except (FileNotFoundError, json.JSONDecodeError):
        existing = []

    existing_keys = {
        entry.get("permalink") for entry in existing if entry.get("permalink")
    }
    new_entries = [
        item for item in processed_data if item.get("permalink") not in existing_keys
    ]

    if new_entries:
        existing.extend(new_entries)
        dvc_write_json(existing, final_store_path)
        print(
            f"[university_au] Appended {len(new_entries)} new entries to {final_store_path}"
        )

        dvc_write_json(new_entries, added_store_path)
        print(
            f"[university_au] Added {len(new_entries)} new entries to {added_store_path}"
        )
    else:
        print("[university_au] No new university entries found.")
