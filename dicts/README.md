# Adding a langauge to Goruut

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

Navigate to the `cmd/analysis` folder. Run `./study_language.sh` romanian in Git Bash, replacing "romanian" with your directory name.

You should see rows like `Improved edit distance to 124482`. This means the algorithm is improving the `language.json`.

If the edit distance is 1 and `language.json` is not improving, then your `language.json` may be incorrect. Please refer to the previous step.

## clean_language.sh

Once you're satisfied with the loss or believe that `study_language.sh` cannot learn any further improvements, it's time to clean the language.

Go to the cmd/analysis folder.
Run `./clean_language.sh` romanian in Git Bash, replacing "romanian" with your directory name.

After this, the `clean.tsv` file will appear in your language folder. Check the number of rows. A significant majority of the rows from the original `dirty.tsv` should be aligned in your `clean.tsv` file:

```
f r u m o s   f r u m o s
m â n c a r e m ɨ n k a r e
a p ă        a p ə
```

If more than 90% of the words are aligned, you can proceed. If not, you are recommended to run study_language.sh further.
