# csl.py
import pandas as pd
import requests
import io
from dvc_utils import dvc_read_csv, dvc_write_csv


def fetch_source(config):
    csv_url = config["csv_url"]
    response = requests.get(csv_url)
    response.raise_for_status()
    data = pd.read_csv(io.StringIO(response.text))
    return data


def process_source(fetched_data, config):
    original_file_path = config["original_file_path"]
    try:
        original_data = dvc_read_csv(original_file_path)
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

    dvc_write_csv(new_data, output_file_path)
    print(
        f"[csl_data] New rows written to {output_file_path}, total new rows: {len(new_data)}"
    )

    try:
        original_data = dvc_read_csv(original_file_path)
        combined_data = pd.concat([original_data, new_data], ignore_index=True)
    except FileNotFoundError:
        combined_data = new_data

    dvc_write_csv(combined_data, original_file_path)
    print(
        f"[csl_data] Updated original file at {original_file_path}, total rows: {len(combined_data)}"
    )
