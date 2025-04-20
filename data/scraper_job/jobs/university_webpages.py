import json
import os
import subprocess


def crawl_university_webpages(config):
    output_file_path = config["intermediate_jsonl_path"]
    os.makedirs(os.path.dirname(output_file_path), exist_ok=True)

    cwd = os.path.join(os.path.dirname(__file__), "spider", "unicrawler")

    cmd = [
        "scrapy",
        "runspider",
        "main.py",
        "-o",
        output_file_path,
        "-a",
        f"openai_api_key={config['openai_api_key']}",
        "-a",
        f"input_json_path={config['input_json_path']}",
    ]
    try:
        process = subprocess.run(cmd, cwd=cwd, capture_output=False, text=True)
        print(f"[university_webpages] Command output:\n{process.stdout}")
        if process.stderr:
            print(f"[university_webpages] Errors/Warnings:\n{process.stderr}")
        if process.returncode != 0:
            raise subprocess.CalledProcessError(
                process.returncode, cmd, process.stdout, process.stderr
            )
    except subprocess.CalledProcessError as e:
        print(f"Error executing scrapy command. Return code: {e.returncode}")
        print(f"Error output:\n{e.stderr}")
        raise

    data = []
    with open(output_file_path, "r", encoding="utf-8") as f:
        for line in f:
            try:
                data.append(json.loads(line))
            except json.JSONDecodeError as e:
                print(f"[university_webpages] Error parsing JSON line: {e}")
                continue
    return data


def process_university_webpages(raw_data, config):
    return raw_data


def update_university_webpages(processed_data, config):
    output_file_path = config["output_json_path"]
    os.makedirs(os.path.dirname(output_file_path), exist_ok=True)

    existing_data = []
    if os.path.exists(output_file_path) and os.path.getsize(output_file_path) > 0:
        with open(output_file_path, "r", encoding="utf-8") as f:
            existing_data = json.load(f)

    existing_data.extend(processed_data)

    with open(output_file_path, "w", encoding="utf-8") as f:
        json.dump(existing_data, f, indent=4, ensure_ascii=False)

    print(f"[university_webpages] Updated data in {output_file_path}")
