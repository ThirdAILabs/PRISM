import json
import pandas as pd
import requests as rq
from urllib.parse import quote
from concurrent.futures import ThreadPoolExecutor, as_completed
from tqdm import tqdm
import time

# --- Step 1: Load and Filter CSV Data ---

# Read the new CSV file (change the filename if needed)
df = pd.read_csv(
    "/Users/pratikqpranav/ThirdAI/PRISM/data/scrapper/csl_data/new_data.csv"
)

# If there's a "type" column, filter out entries labeled as individuals.
if "type" in df.columns:
    df = df[df["type"].str.lower() != "individual"]

# Build a list of entities.
# We expect at least a "name" column.
# If an "aliases" column exists (with comma-separated aliases), split it; otherwise use just the name.
entities = []
for _, row in df.iterrows():
    name = row["name"]
    if "aliases" in row and pd.notna(row["aliases"]):
        aliases = [alias.strip() for alias in str(row["aliases"]).split(",")]
    else:
        aliases = [name]
    entities.append({"name": name, "aliases": aliases})

# --- Step 2: Define OpenAlex Query Functions ---


def get_results(query, entity_id, entity_type):
    """
    Calls the OpenAlex autocomplete endpoint for the given entity type.
    Returns a dict with the local entity id, entity_type, and list of results.
    """
    try:
        url = f"https://api.openalex.org/autocomplete/{entity_type}?q={quote(query, safe='')}&mailto=pratik@thirdai.com"
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


def get_all(all_queries):
    """
    Uses a ThreadPoolExecutor to concurrently execute queries.
    Yields each result.
    """
    AT_A_TIME = 9
    with ThreadPoolExecutor(max_workers=AT_A_TIME) as executor:
        for i in range(0, len(all_queries), AT_A_TIME):
            futures = [
                executor.submit(get_results, query, entity_id, entity_type)
                for query, entity_id, entity_type in all_queries[i : i + AT_A_TIME]
            ]
            for future in as_completed(futures):
                yield future.result()
            time.sleep(1.0)


def search(entities, part_size=100):
    results_institutions = {}
    results_funders = {}
    results_publishers = {}

    for i in range(0, len(entities), part_size):
        part = entities[i : i + part_size]
        print(f"Processing chunk starting at index {i} of {len(entities)}")
        all_queries = []
        for entity_id, inst in enumerate(part):
            queries = [inst["name"]] + inst.get("aliases", [])
            for query in queries:
                for entity_type in ["institutions", "funders", "publishers"]:
                    all_queries.append((query, entity_id, entity_type))

        for r in tqdm(get_all(all_queries), total=len(all_queries)):
            etype = r["entity_type"]
            if etype == "institutions":
                if r["id"] not in results_institutions:
                    results_institutions[r["id"]] = r
                else:
                    results_institutions[r["id"]]["results"].extend(r["results"])
            elif etype == "funders":
                if r["id"] not in results_funders:
                    results_funders[r["id"]] = r
                else:
                    results_funders[r["id"]]["results"].extend(r["results"])
            elif etype == "publishers":
                if r["id"] not in results_publishers:
                    results_publishers[r["id"]] = r
                else:
                    results_publishers[r["id"]]["results"].extend(r["results"])

    # Write aggregated results into separate JSON files.
    json.dump(results_institutions, open("institutions.json", "w"), indent=4)
    json.dump(results_funders, open("funders.json", "w"), indent=4)
    json.dump(results_publishers, open("publishers.json", "w"), indent=4)
    print(
        "Search complete. Results saved to institutions.json, funders.json, and publishers.json."
    )


if __name__ == "__main__":
    search(entities)
