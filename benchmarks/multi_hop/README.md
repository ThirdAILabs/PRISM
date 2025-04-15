# Multi-hop Connection Benchmark Dataset
This dataset is curated to benchmark the quality of PRISM's multi-hop connection scraping. To create this dataset, we started with a DOJ article, manually identified entities in it, and then recursively searched the web for resources about the indicted entity and entities connected to it. 

## Tables
It contains 4 tables, all stored in JSON.

1. `entities.json`: A list of entities with their known risk levels. Indicted individuals or high-risk foreign entities are marked as high-risk. These entities are extracted from 'true positive' resources (see description of `resources.json`).
2. `resources.json`: A list of webpages that are found by searching the web for entity names. Each resources has a corresponding query entity and a "true positive" status. Webpages about homonymous unique individuals are marked as false positives. All query entities in `resources.json` are high-risk entities or entities connected to a high-risk entity.
3. `entity-appearances.json`: Lists entity-resource pairs, marking which resources feature which entities. This file does not feature every resource.
4. `entity-connections.json`: Lists connections between entities and the resources through which these connections can be inferred. This file exhaustively lists the connections between entities in `entity-appearances.json` in their respective files. For example, if `entity-appearances.json` states that entities A and B appear in some file X, then any professional connection between A and B that can be inferred from X would be included in `entity-connections.json`. However, if a connection between A and B can only be inferred from a file Y but `entity-appearances.json` does not recognize A's appearance in Y, then the A -> B connection will not be featured in `entity-connections.json`.

## Viewing and Denormalization
To prevent typos and to save space, all JSONs are normalized; instead of directly referring to entities and resources by name and URL, the JSONs refer to their IDs. We provide a script called `denormalize.sh`, which you can run to `resources.json`, `entity-appearances.json`, and `entity-connections.json`. `entities.json` does not need to be denormalized because it does not refer to other tables.

## Suggested use
Here are some examples of how you can utilize this dataset to assess the quality of PRISM's multi-hop connection scraping:

### Measuring the ability to find correct resources about a particular entity.
1. Choose an entity that is featured in `resources.json`.
2. Use the scraping system to collect resources about this entity.
3. Assert that the system found all "true positive" resources listed for this entity in `resources.json`.
3. Assert that the system discarded all "false positive" resources listed for this entity in `resources.json`.

### Measuring the ability to only extract concerning entities from a resource.
1. Choose an entity that is featured in `resources.json`.
2. Choose a resource for that entity from `resources.json` that also appears in `entity-connections.json`.
3. Extract entities from the chosen resource.
3. Assert that all extracted entities have a connection to the entity from step 1 according to `entity-connections.json`.

### Measuring end-to-end graph completeness
1. Use the dataset to build a graph of entities.
2. Discard nodes that are not high-risk entities and are not connected to one.
3. Assert that PRISM can recover these connections when the user searches for any entity from the graph in step 2.
