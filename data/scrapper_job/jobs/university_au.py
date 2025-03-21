import os
import json
import subprocess


def crawl_university_data(config):
    intermediate_json_path = config["intermediate_json_path"]

    (
        os.remove(intermediate_json_path)
        if os.path.exists(intermediate_json_path)
        else None
    )
    os.makedirs(os.path.dirname(intermediate_json_path), exist_ok=True)

    cwd = os.path.join(os.path.dirname(__file__), "spider", "unitracker")

    cmd = [
        "scrapy",
        "runspider",
        "main.py",
        "-o",
        intermediate_json_path,
        "--nolog",
    ]
    subprocess.run(cmd, cwd=cwd, check=True)

    with open(intermediate_json_path, "r", encoding="utf-8") as f:
        data = json.load(f)
    return data


def process_university_data(raw_data, config):
    return raw_data


def update_university_store(processed_data, config):
    final_store_path = config["final_store_path"]
    added_store_path = config["added_store_path"]
    os.makedirs(os.path.dirname(final_store_path), exist_ok=True)
    if os.path.exists(final_store_path):
        with open(final_store_path, "r", encoding="utf-8") as f:
            try:
                existing = json.load(f)
            except json.JSONDecodeError:
                existing = []
    else:
        existing = []
    existing_keys = {
        entry.get("permalink") for entry in existing if entry.get("permalink")
    }
    new_entries = [
        item for item in processed_data if item.get("permalink") not in existing_keys
    ]
    if new_entries:
        existing.extend(new_entries)
        with open(final_store_path, "w", encoding="utf-8") as f:
            json.dump(existing, f, indent=4, ensure_ascii=False)
        print(
            f"[university_au] Appended {len(new_entries)} new entries to {final_store_path}"
        )

        with open(added_store_path, "w", encoding="utf-8") as f:
            json.dump(new_entries, f, indent=4, ensure_ascii=False)
        print(
            f"[university_au] Added {len(new_entries)} new entries to {added_store_path}"
        )
    else:
        print("[university_au] No new university entries found.")
