cd "$(dirname "$0")"

# Denormalize resources.json
python3 denormalize.py resources.json -p query_entity_id -s name -f entities.json

# Denormalize entity-appearances.json
python3 denormalize.py entity-appearances.json -p entity_id resource_id -s name url -f entities.json resources.json

# Denormalize entity-connections.json
python3 denormalize.py entity-connections.json -p entity1_id entity2_id resource_id -s name name url -f entities.json entities.json resources.json
