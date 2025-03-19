import pandas as pd
import requests
import io
import os


def fetch_source(config):
    csv_url = config["csv_url"]
    response = requests.get(csv_url)
    response.raise_for_status()

    data = pd.read_csv(io.StringIO(response.text))

    os.makedirs(os.path.dirname(config["original_file"]), exist_ok=True)
    return data


def process_source(fetched_data, config):
    original_file = config["original_file"]

    try:
        original_data = pd.read_csv(original_file)
    except FileNotFoundError:
        original_data = pd.DataFrame(columns=fetched_data.columns)
    original_ids = (
        set(original_data["_id"]) if "_id" in original_data.columns else set()
    )
    new_rows = fetched_data[~fetched_data["_id"].isin(original_ids)]
    return new_rows


def update_local_store(new_data, config):
    output_file = config["output_file"]
    os.makedirs(os.path.dirname(output_file), exist_ok=True)
    new_data.to_csv(output_file, index=False)
    print(
        f"[csl_data] New rows written to {output_file}, total new rows: {len(new_data)}"
    )
