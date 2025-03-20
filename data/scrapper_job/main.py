import json
import importlib
import traceback


def load_config(path="config.json"):
    with open(path, "r") as f:
        return json.load(f)


def run_job(job):
    try:
        module_name = job["module"]
        fetcher_name = job["fetcher"]
        processor_name = job["processor"]
        updater_name = job["updater"]
        job_config = job.get("config", {})

        module = importlib.import_module(module_name)
        print(f"\n=== Running job: {job['name']} ===")

        raw_data = getattr(module, fetcher_name)(job_config)
        print(f"[{job['name']}] Fetched raw data.")

        processed_data = getattr(module, processor_name)(raw_data, job_config)
        print(
            f"[{job['name']}] Processed data; items: {len(processed_data) if processed_data is not None else 'N/A'}."
        )

        getattr(module, updater_name)(processed_data, job_config)
        print(f"[{job['name']}] Updated local store.")
    except Exception as e:
        print(f"Error running job {job['name']}: {e}")
        traceback.print_exc()


def resolve_dependencies(jobs):
    job_dict = {job["name"]: job for job in jobs}

    indegree = {job["name"]: 0 for job in jobs}
    graph = {job["name"]: [] for job in jobs}

    for job in jobs:
        for dep in job.get("depends_on", []):
            if dep in job_dict:
                graph[dep].append(job["name"])
                indegree[job["name"]] += 1
            else:
                print(f"Warning: Job '{job['name']}' depends on unknown job '{dep}'")

    queue = [name for name, deg in indegree.items() if deg == 0]
    sorted_names = []

    while queue:
        current = queue.pop(0)
        sorted_names.append(current)
        for neighbor in graph[current]:
            indegree[neighbor] -= 1
            if indegree[neighbor] == 0:
                queue.append(neighbor)

    if len(sorted_names) != len(jobs):
        raise Exception("Cycle detected or missing dependencies among jobs!")

    return [job_dict[name] for name in sorted_names]


def main():
    config = load_config()
    jobs = config.get("jobs", [])

    for key, val in config.get("global", {}).items():
        for job in jobs:
            job.setdefault("config", {})[key] = val

    sorted_jobs = resolve_dependencies(jobs)
    print("Jobs will run in the following order:")
    for job in sorted_jobs:
        print(" ->", job["name"])

    for job in sorted_jobs:
        run_job(job)


if __name__ == "__main__":
    main()
