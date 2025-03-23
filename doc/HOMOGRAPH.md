# TRAINING HOMOGRAPHs

## Dictionary stage

copy latest dirty.tsv and multi.tsv into your language's folder
create dummy (empty) weights5_reverse.json.zlib and weights5.json.zlib:
- `touch ../../dicts/english/weights5_reverse.json.zlib `
- `touch ../../dicts/english/weights5.json.zlib`
in goruut compile cmd/backtest using go build
run ./dict.sh <your_language_dir_name>

## Homograph stage

In the backtest directory, execute:
`./train_homograph.sh english`
You can also specify `--overwrite` to delete your existing model

