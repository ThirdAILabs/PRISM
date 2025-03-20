import os
import json
import subprocess


def crawl_university_data(config):
    intermediate_json = config["intermediate_json"]

    os.remove(intermediate_json) if os.path.exists(intermediate_json) else None
    os.makedirs(os.path.dirname(intermediate_json), exist_ok=True)

    cwd = os.path.join(os.path.dirname(__file__), "spider", "unitracker")

    cmd = [
        "scrapy",
        "runspider",
        "main.py",
        "-o",
        intermediate_json,
        "--nolog",
    ]
    subprocess.run(cmd, cwd=cwd, check=True)

    with open(intermediate_json, "r", encoding="utf-8") as f:
        data = json.load(f)
    return data


def process_university_data(raw_data, config):
    return raw_data


def update_university_store(processed_data, config):
    final_store = config["final_store"]
    added_store = config["added_store"]
    os.makedirs(os.path.dirname(final_store), exist_ok=True)
    if os.path.exists(final_store):
        with open(final_store, "r", encoding="utf-8") as f:
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
        with open(final_store, "w", encoding="utf-8") as f:
            json.dump(existing, f, indent=2, ensure_ascii=False)
        print(
            f"[university_au] Appended {len(new_entries)} new entries to {final_store}"
        )

        with open(added_store, "w", encoding="utf-8") as f:
            json.dump(new_entries, f, indent=2, ensure_ascii=False)
        print(f"[university_au] Added {len(new_entries)} new entries to {added_store}")
    else:
        print("[university_au] No new university entries found.")
