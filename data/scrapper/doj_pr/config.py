import json
from datetime import datetime

def load_config(path="config.json"):
    with open(path, "r") as f:
        return json.load(f)

def update_config(new_config, path="config.json"):
    with open(path, "w") as f:
        json.dump(new_config, f, indent=2)

def set_new_start_date(config, date_str):
    config["start_date"] = date_str
    return config

def parse_date(date_str, fmt="%Y-%m-%d"):
    return datetime.strptime(date_str, fmt)
