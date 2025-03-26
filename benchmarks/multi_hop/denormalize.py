import argparse
from pathlib import Path
import json
from typing import Any


def args() -> tuple[Path, list[str], list[str], list[Path]]:
    parser = argparse.ArgumentParser()
    _ = parser.add_argument("to_denormalize", type=Path)
    _ = parser.add_argument("-p", "--primary_keys", type=str, nargs="+", required=True)
    _ = parser.add_argument("-s", "--secondary_keys", type=str, nargs="+", required=True)
    _ = parser.add_argument("-f", "--secondary_files", type=Path, nargs="+", required=True)
    
    args = parser.parse_args()
    
    if len(args.primary_keys) != len(args.secondary_keys) or len(args.secondary_keys) != len(args.secondary_files):
        raise ValueError("Number of primary keys, secondary keys, and secondary files must match")
    
    return args.to_denormalize, args.primary_keys, args.secondary_keys, args.secondary_files


def build_index_from_secondary_table(secondary_table: dict[str,list[dict[str, Any]]], secondary_key: str) -> dict[str, str]:
    # There should only be one key in the secondary table
    entries = list(secondary_table.values())[0]
    return {
        entry['id']: entry[secondary_key]
        for entry in entries
    }


def denormalize_entry(entry: dict[str, Any], primary_keys: list[str], indices: list[dict[str, str]]) -> dict[str, Any]:
    for key, index in zip(primary_keys, indices):
        # key is always XYZ_id. Remove the "_id" suffix.
        entry[key[:-3]] = index[entry[key]]
        entry.pop(key)
    return entry


def denormalize(table: dict[str, list[dict[str, Any]]], primary_keys: list[str], secondary_keys: list[str], secondary_files: list[Path]):
    indices: list[dict[str, str]] = [
        build_index_from_secondary_table(
            json.load(open(sfile, "r")),
            skey)
        for skey, sfile in zip(secondary_keys, secondary_files)
    ]

    main_key = list(table.keys())[0]
    table[main_key] = [
        denormalize_entry(entry, primary_keys, indices)
        for entry in table[main_key]
    ]

    return table

def main():
    to_denormalize, primary_keys, secondary_keys, secondary_files = args()
    with open(to_denormalize, "r") as f:
        table = json.load(f)

    denormalized_table = denormalize(table, primary_keys, secondary_keys, secondary_files)

    output = to_denormalize.parent / to_denormalize.name.replace(".json", "_denormalized.json")

    with open(output, "w") as f:
        json.dump(denormalized_table, f, indent=2)


if __name__ == "__main__":
    main()
