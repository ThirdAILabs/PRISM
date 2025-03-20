import json
import importlib
import traceback


def load_config(path="config_test.json"):
    with open(path, "r") as f:
        return json.load(f)


def test_job(job):
    module_name = job["module"]
    fetcher_name = job["fetcher"]
    processor_name = job["processor"]
    updater_name = job["updater"]
    job_config = job.get("config", {})

    print(f"\n--- Testing job: {job['name']} ---")
    try:
        module = importlib.import_module(module_name)
    except Exception as e:
        print(f"Error importing module {module_name}: {e}")
        return

    try:
        raw_data = getattr(module, fetcher_name)(job_config)
        print(
            f"[{job['name']}] Fetcher succeeded: returned data of type {type(raw_data)}"
        )
    except Exception as e:
        print(f"[{job['name']}] Fetcher failed: {e}")
        traceback.print_exc()
        return

    # Test processor
    try:
        processed_data = getattr(module, processor_name)(raw_data, job_config)
        print(
            f"[{job['name']}] Processor succeeded: returned data of type {type(processed_data)}"
        )
    except Exception as e:
        print(f"[{job['name']}] Processor failed: {e}")
        traceback.print_exc()
        return

    # Test updater
    try:
        getattr(module, updater_name)(processed_data, job_config)
        print(f"[{job['name']}] Updater succeeded.")
    except Exception as e:
        print(f"[{job['name']}] Updater failed: {e}")
        traceback.print_exc()


def main():
    config = load_config()
    jobs = config.get("jobs", [])

    for key, val in config.get("global", {}).items():
        for job in jobs:
            job.setdefault("config", {})[key] = val

    if not jobs:
        print("No jobs defined in config.")
        return

    for job in jobs:
        test_job(job)


if __name__ == "__main__":
    main()
