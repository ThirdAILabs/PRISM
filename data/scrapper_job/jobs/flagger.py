import json
import pandas as pd
import requests as rq
from urllib.parse import quote
from concurrent.futures import ThreadPoolExecutor, as_completed
from tqdm import tqdm
import time
import os
from difflib import SequenceMatcher


def fetch_flagger_source(config):
    input_csv_path = config["input_csv_path"]
    df = pd.read_csv(input_csv_path)
    if "type" in df.columns:
        df = df[df["type"].str.lower() != "individual"]
    entities = []
    for _, row in df.iterrows():
        entities.append(row.to_dict())
    return entities


def get_results(query, inst, entity_id, entity_type, email="pratik@thirdai.com"):
    try:
        url = f"https://api.openalex.org/autocomplete/{entity_type}?q={quote(query, safe='')}&mailto={email}"
        response = rq.get(url, timeout=3.0)
        response.raise_for_status()
        if len(response.json()["results"]) == 0:
            return {}

        def get_similarity(a, b):
            return SequenceMatcher(None, a.lower(), b.lower()).ratio()

        return {
            "entity_type": entity_type,
            "results": [
                {
                    "name": query,
                    "source": inst["source"],
                    "aliases": (
                        inst.get("alt_names", "").split("; ")
                        if not pd.isna(inst.get("alt_names"))
                        else []
                    ),
                    "match": {
                        "id": r["id"],
                        "name": r["display_name"],
                        "type": "entity_type",
                    },
                    "score": get_similarity(r["display_name"], query),
                }
                for r in response.json()["results"]
            ],
        }
    except Exception as e:
        print(f"Error for query '{query}' ({entity_type}): {e}")
        return {"id": entity_id, "entity_type": entity_type, "results": []}


def get_all(all_queries, email="pratik@thirdai.com"):
    AT_A_TIME = 1
    results = []
    with ThreadPoolExecutor(max_workers=AT_A_TIME) as executor:
        for i in range(0, len(all_queries), AT_A_TIME):
            futures = [
                executor.submit(get_results, query, inst, entity_id, entity_type, email)
                for query, inst, entity_id, entity_type in all_queries[
                    i : i + AT_A_TIME
                ]
            ]
            for future in as_completed(futures):
                results.append(future.result())
            time.sleep(0.2)
    return results


def process_flagger_source(entities, config):
    all_queries = []
    for entity_id, inst in enumerate(entities):
        if pd.isna(inst["name"]):
            continue
        queries = [inst["name"]] + (
            inst.get("alt_names", "").split("; ")
            if not pd.isna(inst.get("alt_names"))
            else []
        )
        if "" in queries:
            queries.remove("")
        for query in queries:
            for entity_type in ["institutions", "funders", "publishers"]:
                all_queries.append((query, inst, entity_id, entity_type))
    results = {"institutions": [], "funders": [], "publishers": []}
    for r in tqdm(get_all(all_queries), total=len(all_queries)):
        if r:
            etype = r["entity_type"]
            del r["entity_type"]
            results[etype].extend(r["results"])
    return results


def update_flagger_store(processed_data, config):
    out_inst = config["output_institutions_path"]
    out_funders = config["output_funders_path"]
    out_publishers = config["output_publishers_path"]
    os.makedirs(os.path.dirname(out_inst), exist_ok=True)

    def update_json_file(filepath, new_data):
        existing_data = []
        if os.path.exists(filepath):
            with open(filepath, "r") as f:
                existing_data = json.load(f)

        existing_names = {item["name"] for item in existing_data}

        new_items = [item for item in new_data if item["name"] not in existing_names]
        existing_data.extend(new_items)

        with open(filepath, "w") as f:
            json.dump(existing_data, f, indent=4)

    update_json_file(out_inst, processed_data.get("institutions", []))
    update_json_file(out_funders, processed_data.get("funders", []))
    update_json_file(out_publishers, processed_data.get("publishers", []))

    print(f"[flagger_data] Updated institutions, funders, and publishers JSON files.")
