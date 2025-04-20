import pandas as pd
import requests
import io
import os


def fetch_source(config):
    csv_url = config["csv_url"]
    response = requests.get(csv_url)
    response.raise_for_status()

    data = pd.read_csv(io.StringIO(response.text))

    os.makedirs(os.path.dirname(config["original_file_path"]), exist_ok=True)
    return data


def process_source(fetched_data, config):
    original_file_path = config["original_file_path"]

    try:
        original_data = pd.read_csv(original_file_path)
    except FileNotFoundError:
        original_data = pd.DataFrame(columns=fetched_data.columns)
    original_ids = (
        set(original_data["_id"]) if "_id" in original_data.columns else set()
    )
    new_rows = fetched_data[~fetched_data["_id"].isin(original_ids)]
    return new_rows


def update_local_store(new_data, config):
    output_file_path = config["output_file_path"]
    original_file_path = config["original_file_path"]

    os.makedirs(os.path.dirname(output_file_path), exist_ok=True)
    new_data.to_csv(output_file_path, index=False)
    print(
        f"[csl_data] New rows written to {output_file_path}, total new rows: {len(new_data)}"
    )

    try:
        original_data = pd.read_csv(original_file_path)
        combined_data = pd.concat([original_data, new_data], ignore_index=True)
    except FileNotFoundError:
        combined_data = new_data

    combined_data.to_csv(original_file_path, index=False)
    print(
        f"[csl_data] Updated original file at {original_file_path}, total rows: {len(combined_data)}"
    )
