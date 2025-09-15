#!/bin/bash

# ˑ
rgrep "$(printf '\315\241\|\315\234\|\311\241')" . -I --exclude-dir .git
grep \\. -I --exclude-dir .git */lexicon.tsv */*/lexicon.tsv
egrep '\[[^"]' -I --exclude-dir .git */lexicon.tsv */*/lexicon.tsv
egrep '[^"]\]' -I --exclude-dir .git */lexicon.tsv */*/lexicon.tsv
egrep '	.*/' -I --exclude-dir .git */lexicon.tsv */*/lexicon.tsv
rgrep ' ' --exclude-dir .git | \
	grep --invert-match 'README.md' | \
	grep --invert-match chinese | \
	grep --invert-match cantonese | \
	grep --invert-match minnan | \
	grep --invert-match japanese | \
	grep --invert-match tibetan | \
	grep --invert-match shantaiyai | \
	grep --invert-match english/multi
egrep '[[:upper:]]' */lexicon.tsv */*/lexicon.tsv
