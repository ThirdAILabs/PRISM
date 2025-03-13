# Benchmark Results

These are synthetic data benchmarks that take the entities we are searching over in different parts of PRISM and generate alternative forms of the entity names with openai. They then compare the new entity lookup index with NDB and Flash. 

## Benchmark 1: Multi-hop dataset
This dataset is from taking a 1000 random entities from the entities found in the auxiliary webpages and doj press release pages. There are about 6000 entities originally. Synethic queries are generated from these entities as described above.

### Results
| Method            | p@1  | p@10 |
|-------------------|------|------|
| New Entity Lookup | __0.822__ | __0.930__ |
| NDB               | 0.701| 0.866|
| Flash             | 0.681| 0.755|

## Benchmark 2: Watchlist entities
This dataset is from the aliases from the list of aliases on government watchlists that we use in the acknowledgement flagger. Synethic queries are generated from these entities as described above.

### Results
| Method            | p@1  | p@10 |
|-------------------|------|------|
| New Entity Lookup | __0.556__ | __0.745__ |
| NDB               | 0.420| 0.560|
| Flash             | 0.476| 0.602|