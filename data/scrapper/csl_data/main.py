import pandas as pd
import requests
import io

# URL for the consolidated CSV
CSV_URL = "https://data.trade.gov/downloadable_consolidated_screening_list/v1/consolidated.csv"

# Fetch the CSV data from the URL
response = requests.get(CSV_URL)
response.raise_for_status()  # Raise an error if the request failed

# Read the fetched CSV data into a DataFrame
fetched_data = pd.read_csv(io.StringIO(response.text))

# Read the locally stored CSV data
original_data = pd.read_csv("original.csv")

# Determine new rows by comparing the _id column values
original_ids = set(original_data["_id"])
new_rows = fetched_data[~fetched_data["_id"].isin(original_ids)]

# Write the new rows to a CSV file
new_rows.to_csv("new_data.csv", index=False)

print(f"New rows written to new_data.csv, total new rows: {len(new_rows)}")
