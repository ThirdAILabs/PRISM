import json
import re
from typing import List

with open("./data/searchable_entities.json") as f:
    entities: List[str] = json.load(f)

print(entities[0])

print(len(entities))


pattern = re.compile(r"\[([A-Z ]+)\](.+?)(?=\[)", flags=re.DOTALL)


def process_entity(entity) -> dict:
    data = {}
    for match in re.findall(pattern, entity):
        key = match[0]
        if key == "NAMES START":
            key = "Names"
        else:
            key = key.lower()
            key = key[0].upper() + key[1:]
        value = match[1].strip()

        if key == "Address" and value == "nan":
            value = ""
        if value != "":
            data[key] = value
    return data


entities = [process_entity(entity) for entity in entities]

with open("./data/searchable_entities_processed.json", "w") as f:
    json.dump(entities, f, indent=4)
