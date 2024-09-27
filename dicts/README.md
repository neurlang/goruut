# Adding a language to Goruut

## Roadmap

![Adding a langauge to Goruut](adding_a_language.drawio.svg)

## Creating the language folder

Create a folder in `dicts/` according to the name of your language. This folder will be referred to as the language folder.

## dirty.tsv

Create a `dirty.tsv` text file in the language folder. This is a TSV file that serves as the dictionary, mapping the language words (or sentences, if your language doesn't separate words by spaces) to their IPA pronunciations.

Example `dirty.tsv` content (Romanian language):

```
frumos     fruˈmos
mâncare    mɨnˈkare
apă        ˈapə
om         om
femeie     feˈmeje
dragoste   ˈdraɡoste
copil      koˈpil
floare     ˈfloare
pădure     pəˈdure
soare      ˈsoare
```

In case there are multiple possible IPA pronunciations for a specific language word, use multiple rows in `dirty.tsv` (Romanian language):

```
înțelege   ɨntseˈledʒe
înțelege   ɨntseˈleʒe
```

## language.json

Here’s a default `language.json` that you can use as a starting point for the algorithm. You need to input at least two letters of the written language (we chose "a" and "j" for Romanian) and provide their IPA pronunciations. In this case, "a" is pronounced as "a" and "j" is pronounced as "ʒ".

```json
{
  "Map": {
    "a": ["a"],
    "j": ["ʒ"]
  },
  "SrcMulti": null,
  "DstMulti": null,
  "SrcMultiSuffix": null,
  "DstMultiSuffix": ["ː"],
  "DropLast": null,
  "PrePhonWordSteps": [
    { "Trim": ".," },
    { "ToLower": true }
  ]
}
```

## study_language.sh

If you don’t have Go installed, use `apt-get install golang` (on Linux).
On Windows, install [Golang](https://go.dev/) and [Git Bash](https://git-scm.com/downloads).

In Bash, navigate to the `cmd/analysis` folder. Type `go build` to compile the analysis program.

Now, you can initiate `study_language.sh`, provided you have both `dirty.tsv` and the initial `language.json` for your language.

Navigate to the `cmd/analysis` folder. Run `./study_language.sh romanian` in Git Bash, replacing "romanian" with your directory name.

You should see rows like `Improved edit distance to 124482`. This means the algorithm is improving the `language.json`.

If the edit distance is 1 and `language.json` is not improving, then your `language.json` may be incorrect. Please refer to the previous step.

## clean_language.sh

Once you're satisfied with the loss or believe that `study_language.sh` cannot learn any further improvements, it's time to clean the language.

Go to the cmd/analysis folder.
Run `./clean_language.sh romanian` in Git Bash, replacing "romanian" with your directory name.

After this, the `clean.tsv` file will appear in your language folder. Check the number of rows. A significant majority of the rows from the original `dirty.tsv` should be aligned in your `clean.tsv` file:

```
f r u m o s   f r u m o s
m â n c a r e m ɨ n k a r e
a p ă        a p ə
```

If more than 90% of the words are aligned, you can proceed. If not, you are recommended to run study_language.sh further.

## Train phonemizer

The prerequisite for this step is the clean.tsv file.

1. Checkout this repo: `https://github.com/neurlang/classifier`
2. Navigate to the `cmd/train_phonemizer` subdirectory.
3. Compile the program using `go build`.
4. Run train_phonemizer with `-cleantsv PATH_TO_YOUR_CLEAN_TSV_FILE`:
   `./train_phonemizer -cleantsv ../../../goruut/dicts/romanian/clean.tsv`

The algorithm will run for a while. After each retraining of the hashtron network,
files with the pattern `output.*.json.t.lzw` will start appearing.
The number (`*`) means the percentage of how successful the resulting model is.

I got a number of files:
* `output.60.json.t.lzw`
* `output.71.json.t.lzw`
* `output.88.json.t.lzw`
* `output.80.json.t.lzw`
* `output.93.json.t.lzw`

The `output.93.json.t.lzw` is the best file as its success rate is 93%.

5. Move the `output.93.json.t.lzw` into the language dir and rename it to
   `weights1.json.lzw`.
6. Delete the files with lower success rates.

## Adding the glue code (language.go)

* Copy `language.go` from another language and place it in your datadir
* Change `package otherlanguage` to `package yourfoldername` in the first line.
* Double-check in `language.go` that `weights1.json.lzw` is embedded.

## Adding the glue code (dicts.go in dicts/ folder)

* Add import to the top of dicts.go referring to your language folder
* Add new case to the switch statement:
  * `case "UserFriendlyLanguageName":`
  * `return yourlanguage.Language.ReadFile(lzw(filename))`

## Testing the model

1. Navigate to the `cmd/goruut` directory.
2. Recompile using `go build`
3. Run it, pointing it to the default config file: `./goruut -configfile ../../configs/config.json`
4. Issue an HTTP POST REQUEST:

POST http://127.0.0.1:18080/tts/phonemize/sentence
```json
{
    "Language": "Norwegian",
    "Sentence": "bjornson"
}
```
You should see a response like:
```json
{
	"Words": [
		{
			"CleanWord": "bjornson",
			"Linguistic": "bjornson",
			"Phonetic": "bjɔɳsɔn"
		}
	]
}
```

You can test words that were included in your `clean.tsv`, as only those will work.
Furthermore, if your phonemizer model does have less than 100% success rate, some
words from `clean.tsv` may not work.
