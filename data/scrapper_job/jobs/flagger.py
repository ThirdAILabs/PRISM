import json
import pandas as pd
import requests as rq
from urllib.parse import quote
from concurrent.futures import ThreadPoolExecutor, as_completed
from tqdm import tqdm
import time
import os


def fetch_flagger_source(config):
    input_csv = config["input_csv"]
    df = pd.read_csv(input_csv)
    if "type" in df.columns:
        df = df[df["type"].str.lower() != "individual"]
    entities = []
    for _, row in df.iterrows():
        name = row["name"]
        if "aliases" in row and pd.notna(row["aliases"]):
            aliases = [alias.strip() for alias in str(row["aliases"]).split(",")]
        else:
            aliases = [name]
        entities.append({"name": name, "aliases": aliases})
    return entities


def get_results(query, entity_id, entity_type, email="pratik@thirdai.com"):
    try:
        url = f"https://api.openalex.org/autocomplete/{entity_type}?q={quote(query, safe='')}&mailto={email}"
        response = rq.get(url, timeout=1.0)
        response.raise_for_status()
        return {
            "id": entity_id,
            "entity_type": entity_type,
            "results": [
                {"id": r["id"], "name": r["display_name"], "type": entity_type}
                for r in response.json()["results"]
            ],
        }
    except Exception as e:
        print(f"Error for query '{query}' ({entity_type}): {e}")
        return {"id": entity_id, "entity_type": entity_type, "results": []}


def get_all(all_queries, email="pratik@thirdai.com"):
    AT_A_TIME = 9
    results = []
    with ThreadPoolExecutor(max_workers=AT_A_TIME) as executor:
        for i in range(0, len(all_queries), AT_A_TIME):
            futures = [
                executor.submit(get_results, query, entity_id, entity_type, email)
                for query, entity_id, entity_type in all_queries[i : i + AT_A_TIME]
            ]
            for future in as_completed(futures):
                results.append(future.result())
            time.sleep(1.0)
    return results


def process_flagger_source(entities, config):
    all_queries = []
    for entity_id, inst in enumerate(entities):
        queries = [inst["name"]] + inst.get("aliases", [])
        for query in queries:
            for entity_type in ["institutions", "funders", "publishers"]:
                all_queries.append((query, entity_id, entity_type))
    results = {"institutions": {}, "funders": {}, "publishers": {}}
    for r in tqdm(get_all(all_queries), total=len(all_queries)):
        etype = r["entity_type"]
        if r["id"] not in results[etype]:
            results[etype][r["id"]] = r
        else:
            results[etype][r["id"]]["results"].extend(r["results"])
    return results


def update_flagger_store(processed_data, config):
    out_inst = config["output_institutions"]
    out_funders = config["output_funders"]
    out_publishers = config["output_publishers"]
    os.makedirs(os.path.dirname(out_inst), exist_ok=True)
    with open(out_inst, "w") as f:
        json.dump(processed_data.get("institutions", {}), f, indent=4)
    with open(out_funders, "w") as f:
        json.dump(processed_data.get("funders", {}), f, indent=4)
    with open(out_publishers, "w") as f:
        json.dump(processed_data.get("publishers", {}), f, indent=4)
    print(f"[flagger_data] Updated institutions, funders, and publishers JSON files.")
