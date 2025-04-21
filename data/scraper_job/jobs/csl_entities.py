import pandas as pd
import os
import json


def read_csv_source(config):
    input_csv_path = config["input_csv_path"]
    if not os.path.exists(input_csv_path):
        raise FileNotFoundError(
            f"CSV file not found: {input_csv_path}. CSL Job should be run first."
        )
    data = pd.read_csv(input_csv_path)
    return data


def convert_csv_to_json(fetched_data, config):
    new_json_entries = []
    for index, row in fetched_data.iterrows():
        name = row.get("name", "")
        resource = ""

        if pd.isna(name):
            continue

        alt_names = row.get("alt_names", "")
        if pd.notna(alt_names) and alt_names:
            combined_names = f"{name}\n{alt_names}"
        else:
            combined_names = name

        address = row.get("addresses", "")

        source = row.get("source", "")
        source_list_url = row.get("source_list_url", "")

        if pd.isna(source) and pd.isna(source_list_url):
            continue

        if pd.notna(source_list_url) and source_list_url:
            resource = f"{source} {source_list_url}"
        else:
            resource = source

        json_entry = {"Names": combined_names, "Resource": resource}
        if pd.notna(address) and address:
            json_entry["Address"] = address

        new_json_entries.append(json_entry)
    return new_json_entries


def update_json_file(new_data, config):
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

    unique_new_data = [
        entry for entry in new_data if entry["Names"] not in existing_names
    ]

    existing_data.extend(unique_new_data)
    with open(output_file_path, "w", encoding="utf-8") as f:
        json.dump(existing_data, f, indent=4)

    print(
        f"[update_entities_with_csl] Appended {len(unique_new_data)} unique entries to {output_file_path}. Total entries now: {len(existing_data)}"
    )
