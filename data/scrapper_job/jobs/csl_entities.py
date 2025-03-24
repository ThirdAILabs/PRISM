import pandas as pd
import json
from dvc_utils import dvc_read_csv, dvc_read_json, dvc_write_json


def read_csv_source(config):
    input_csv_path = config["input_csv_path"]
    try:
        data = dvc_read_csv(input_csv_path)
    except FileNotFoundError:
        raise FileNotFoundError(
            f"CSV file not found: {input_csv_path}. CSL Job should be run first."
        )
    return data


def convert_csv_to_json(fetched_data, config):
    new_json_entries = []
    for index, row in fetched_data.iterrows():
        name = row.get("name", "")
        resource = ""

        if pd.isna(name):
            continue

        alt_names = row.get("alt_names", "")
        combined_names = (
            f"{name}\n{alt_names}" if pd.notna(alt_names) and alt_names else name
        )

        address = row.get("addresses", "")
        source = row.get("source", "")
        source_list_url = row.get("source_list_url", "")

        if pd.isna(source) and pd.isna(source_list_url):
            continue

        resource = (
            f"{source} {source_list_url}"
            if pd.notna(source_list_url) and source_list_url
            else source
        )

        json_entry = {"Names": combined_names, "Resource": resource}
        if pd.notna(address) and address:
            json_entry["Address"] = address

        new_json_entries.append(json_entry)
    return new_json_entries


def update_json_file(new_data, config):
    output_file_path = config["output_file_path"]
    try:
        existing_data = dvc_read_json(output_file_path)
    except (FileNotFoundError, json.JSONDecodeError):
        existing_data = []

    existing_names = {entry["Names"] for entry in existing_data}
    unique_new_data = [
        entry for entry in new_data if entry["Names"] not in existing_names
    ]

    existing_data.extend(unique_new_data)
    dvc_write_json(existing_data, output_file_path)

    print(
        f"[update_entities_with_csl] Appended {len(unique_new_data)} unique entries to {output_file_path}. Total entries now: {len(existing_data)}"
    )
