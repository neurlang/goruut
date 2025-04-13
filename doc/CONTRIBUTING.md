# Contributing language model/lexicon to goruut

There are two options using which you - the author of a language lexicon - can
contribute to the goruut project.

- Contributing a model
- Contributing a lexicon

**If you are developer - see DEVELOPING.md instead**

## Contributing a model

In this scenario, the lexicon remains proprietary - owned by the author.
However, the author trains a model using their private lexicon, which can help
the whole community.

The model itself will be distributed under a permissive license, such as MIT /
BSD / Apache 2 - with the agreement of the author.

The author will have full copyright on their model, and can publish a new version
of the model at any time (once per dev cycle), or ask the model to be removed
from the project.

A pull request must be opened to the neurlang/goruut repository.

The pull request must contain (for the specific language):
- language.go
- language.json
- language_reverse.json
- weights4.json.zlib
- weights4_reverse.json.zlib
- LICENSE.md (**according the choice of the author**)

The pull request can't contain (for the specific language):
- lexicon.tsv (**will leak the lexicon**)
- clean.tsv (**will partially leak the lexicon**)
- clean_reverse.tsv (**will partially leak the lexicon**)
- missing.all.zlib (**will partially leak the lexicon**)

**Warning:** If you leak stolen data, your pull request will be deleted.

## Contributing a lexicon

In this scenario, the lexicon is released to the public - owned by the author -
but is distributed under a permissive license, such as MIT / BSD / Apache 2.

A pull request must be opened to the neurlang/datasets repository.

The pull request must contain (for the specific language):
- lexicon.tsv
- LICENSE.md (or an equivalent licensing message in the main README.md)
