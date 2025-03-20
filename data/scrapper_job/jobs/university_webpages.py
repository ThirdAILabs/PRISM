import json
import os
import subprocess


def crawl_university_webpages(config):
    output_file = config["output_file"]
    os.makedirs(os.path.dirname(output_file), exist_ok=True)

    cwd = os.path.join(os.path.dirname(__file__), "spider", "unicrawler")

    cmd = [
        "scrapy",
        "runspider",
        "main.py",
        "-o",
        output_file,
        "-a",
        f"openai_api_key={config['openai_api_key']}",
        "-a",
        f"input_json={config['input_json']}",
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
    with open(output_file, "r", encoding="utf-8") as f:
        for line in f:
            data.append(json.loads(line))
    return data


def process_university_webpages(raw_data, config):
    return raw_data


def update_university_webpages(processed_data, config):
    output_file = config["output_file"]
    os.makedirs(os.path.dirname(output_file), exist_ok=True)

    existing_data = []
    if os.path.exists(output_file) and os.path.getsize(output_file) > 0:
        with open(output_file, "r", encoding="utf-8") as f:
            existing_data = json.load(f)

    existing_data.extend(processed_data)

    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(existing_data, f, indent=4, ensure_ascii=False)

    print(f"[university_webpages] Updated data in {output_file}")
