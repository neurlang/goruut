# TRAINING HOMOGRAPHs

## Dictionary stage

1. copy latest lexicon.tsv and multi.tsv into your language's folder
2. create dummy (empty) weights5_reverse.json.zlib and weights5.json.zlib:
  - `touch ../../dicts/english/weights5_reverse.json.zlib `
  - `touch ../../dicts/english/weights5.json.zlib`
3. in goruut compile cmd/backtest using go build
4. run `./dict.sh <your_language_dir_name>`

## Verifying the dataset (lexicon.tsv, multi.tsv, multi_eval.tsv)

1. Make sure relevant words in lexicon.tsv are tagged. **Without any words tagged, the model won't learn anything.**
Example tagged:
```
subordinate	sə'bɔːɹdənət	["adjective","noun"]
subordinate	sə'bɔːɹdəˌneɪt	["verb"]
```
2. Make sure that all words that are in the homograph/homophone sentences in multi.tsv have their words in lexicon.
3. Make sure that there are not any tag collisions in lexicon. A tag collision is a word with identical writing and tags.
Example collsion:
```
remained	ɹɪmˈeɪnd	["verb"]
remained	ɹɪˈmeɪnd	["verb"]
```
As you can see it has identical tags.
You will be warned about these scenarios in the next stage.

## Homograph stage

1. In the backtest directory, execute:
  - `./train_homograph.sh english`
2. You can also specify `--overwrite` to delete your existing model

## Backtest stage

- Compile homograph test:
   `cmd/build.sh`
- Test WordErrorRate ignoring stress
   `./homotest --langname english -nostress`
- Or test WordErrorRate taking stress into account
   `./homotest --langname english`
