# Adding a language to Goruut

## Roadmap

![Adding a langauge to Goruut](adding_a_language.drawio.svg)

## Creating the language folder

Create a folder in `dicts/` according to the name of your language. This folder will be referred to as the language folder.

If your language is like Chinese, Japanese or Tibetan by not using the space separator, your language will be referred to as "padspace" language.

## dirty.tsv

Create a `dirty.tsv` text file in the language folder. This is a TSV file that serves as the dictionary, mapping the language words (or sentences, if your language doesn't separate words by spaces) to their IPA pronunciations.

Example `dirty.tsv` content (Romanian language):

0 | TAB | 1
---------|-|----------
frumos   | | fruˈmos
mâncare  | | mɨnˈkare
apă      | | ˈapə
om       | | om
femeie   | | feˈmeje
dragoste | | ˈdraɡoste
copil    | | koˈpil
floare   | | ˈfloare
pădure   | | pəˈdure
soare    | | ˈsoare

In case there are multiple possible IPA pronunciations for a specific language word, use multiple rows in `dirty.tsv` (Romanian language):

0 | TAB | 1
----------|-|----------
înțelege  | | ɨntseˈledʒe
înțelege  | | ɨntseˈleʒe

**Padspace languages** must use both whole sentences and individual words in their `dirty.tsv` file.

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
    { "ToLower": true }
  ]
}
```

## study_language.sh

If you don’t have Go installed, use `apt-get install golang` (on Linux).
On Windows, install [Golang](https://go.dev/) and [Git Bash](https://git-scm.com/downloads).

In Bash, navigate to the `cmd/analysis2` folder. Type `go build` to compile the analysis2 program.

Now, you can initiate `study_language.sh`, provided you have both `dirty.tsv` and the initial `language.json` for your language.

Navigate to the `cmd/analysis2` folder. Run `./study_language.sh romanian` in Git Bash, replacing "romanian" with your directory name.

If the edit distance is 1 and `language.json` is not improving, then your `language.json` may be incorrect. Please refer to the previous step.

**Padspace languages** must use the `-padspace` parameter to `./study_language.sh <name>` script.

**Hyperparameter** controlling the tradeoff between number of rules learned (coverage) and decision complexity of the grammar (model size) named `--rowlossimportance` is available. It is by default set to the value of 5. You can try different values to tweak your language.

## clean_language.sh

Once you're satisfied with the loss or believe that `study_language.sh` cannot learn any further improvements because it ended after you ran it multiple times, it's time to clean the language.

**Mandatory**: You will be needing the `dicttomap` executable which you can compile using `go build` in the `cmd/dicttomap` directory.

Go to the cmd/analysis2 folder.
Run `./clean_language.sh romanian` in Git Bash, replacing "romanian" with your directory name. For this step, you also need the `dicttomap` binary.

After this, the `clean.tsv` file will appear in your language folder. Check the number of rows. A significant majority of the rows from the original `dirty.tsv` should be aligned in your `clean.tsv` file:

0 | 1 | 2 | 3 | 4 | 5 | 6 | TAB | 0 | 1 | 2 | 3 | 4 | 5 | 6
--|---|---|---|---|---|---|-----|---|---|---|---|---|---|---
f | r | u | m | o | s |   |     | f | r | u | m | o | s |
m | â | n | c | a | r | e |     | m | ɨ | n | k | a | r | e
a | p | ă |   |   |   |   |     | a | p | ə |   |   |   |

Note: For -padspace languages (CJT), ensure you handle padding appropriately.

**Padspace languages** must use the `-padspace` parameter to `./clean_language.sh <name>` script.

### coverage.sh

1. Navigate to the `cmd/backtest` directory.
2. Run `./coverage.sh romanian`
3. If more than say 50% of the words are covered, you can proceed. If not, you are
   recommended to run `study_language.sh` then `clean_language.sh` further.
4. Only look at the direction you are training. Forward direction is language to IPA, reverse direction is IPA to language.

## train_language.sh

The prerequisite for this step is the clean.tsv file.

1. Checkout this repo: `https://github.com/neurlang/classifier`
2. Navigate to the `cmd/train_phonemizer` subdirectory.
3. Compile the program using `go build`.
4. Similarly, compile the `backtest` program in the `cmd/backtest` directory.
5. Run `train_language.sh romanian -maxpremodulo <value>` where `<value>` is 5 times the complexity from the cleaning process (recommended).

The algorithm will run for a while. After each improvement of the hashtron network,
file with the name `weights4.json.lzw` will start appearing in your language's folder.

To resume training later, use `train_language.sh romanian -resume -maxpremodulo <value>` 

**Note**: If training a brand new language, backtesting may not work unless you add the glue code, then compile `backtest`, then train language.

## Adding the glue code (language.go)

* Copy `language.go` from another language and place it in your datadir
* Change `package otherlanguage` to `package yourfoldername` in the first line.
* Double-check in `language.go` that `weights4.json.lzw` is embedded.

## Adding the glue code (dicts.go in dicts/ folder)

* Add import to the top of dicts.go referring to your language folder
* Add new case to the switch statement:
  * `case "UserFriendlyLanguageName":`
  * `return yourlanguage.Language.ReadFile(lzw(filename))`
* In the second function LangName, add the langname according to your
  UserFriendlyLanguageName and the actual folder.
* You need to add the language folder into the `LangName(...)` for backtesting code to work.

## Backtesting the model

1. Navigate to the `cmd/backtest` directory.
2. Recompile using `go build`
3. Run the backtest, providing the parameter -langname
4. Example: `./backtest -langname romanian`
5. It will run for a while, printing the end-to-end success rate on words
   in your language.
