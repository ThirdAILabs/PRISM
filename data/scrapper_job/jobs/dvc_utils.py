# dvc_utils.py
import os
import json
import subprocess
import pandas as pd


def dvc_open(file_path, mode="r"):
    if file_path.startswith("s3://"):
        import dvc.api

        # Note: adjust repo and rev if needed.
        return dvc.api.open(file_path, repo=".", mode=mode)
    else:
        return open(file_path, mode)


def dvc_read_csv(file_path):
    if file_path.startswith("s3://"):
        import dvc.api

        with dvc.api.open(file_path, repo=".", mode="r") as f:
            return pd.read_csv(f)
    else:
        return pd.read_csv(file_path)


def dvc_write_csv(df, file_path):
    os.makedirs(os.path.dirname(file_path), exist_ok=True)
    df.to_csv(file_path, index=False)
    subprocess.run(["dvc", "add", file_path], check=True)


def dvc_read_json(file_path):
    if file_path.startswith("s3://"):
        import dvc.api

        with dvc.api.open(file_path, repo=".", mode="r") as f:
            return json.load(f)
    else:
        with open(file_path, "r", encoding="utf-8") as f:
            return json.load(f)


def dvc_write_json(data, file_path):
    os.makedirs(os.path.dirname(file_path), exist_ok=True)
    with open(file_path, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=4)
    subprocess.run(["dvc", "add", file_path], check=True)
