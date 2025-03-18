import os
import json


# Define the pipeline using JSON instead of CSV.
class DiffPipeline:
    def open_spider(self, spider):
        self.items = []

    def process_item(self, item, spider):
        self.items.append(dict(item))
        return item

    def close_spider(self, spider):
        json_filename = "unitracker_au.json"
        new_json_filename = "unitracker_au_new.json"

        # Load existing data from the JSON file if it exists.
        if os.path.exists(json_filename):
            with open(json_filename, "r", encoding="utf-8") as f:
                try:
                    existing_data = json.load(f)
                except json.JSONDecodeError:
                    existing_data = []
        else:
            existing_data = []

        # Create a set of unique keys from the existing data (using 'permalink').
        existing_keys = {
            entry.get("permalink") for entry in existing_data if entry.get("permalink")
        }

        # Filter out items that are already present.
        new_entries = [
            item for item in self.items if item.get("permalink") not in existing_keys
        ]

        if not new_entries:
            spider.logger.info("No new university entries found.")
        else:
            spider.logger.info(
                f"Found {len(new_entries)} new entries. Saving to {new_json_filename} ..."
            )
            with open(new_json_filename, "w", encoding="utf-8") as f:
                json.dump(new_entries, f, indent=2, ensure_ascii=False)
