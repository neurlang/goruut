# TRAINING HOMOGRAPHs

## Dictionary stage

1. copy latest lexicon.tsv and multi.tsv into your language's folder
2. create dummy (empty) weights5_reverse.json.zlib and weights5.json.zlib:
  - `touch ../../dicts/english/weights5_reverse.json.zlib `
  - `touch ../../dicts/english/weights5.json.zlib`
3. in goruut compile cmd/backtest using go build
4. run ./dict.sh <your_language_dir_name>

## Homograph stage

1. In the backtest directory, execute:
  - `./train_homograph.sh english`
2. You can also specify `--overwrite` to delete your existing model

