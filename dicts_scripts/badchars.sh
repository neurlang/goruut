#!/bin/bash

# ˑ
rgrep -P '\xC9\xA1' . -I --exclude-dir .git
grep \\. -I --exclude-dir .git */dirty.tsv */*/dirty.tsv
rgrep -P '\xCD\xA1' . -I --exclude-dir .git
rgrep -P '\xCD\x9C' . -I --exclude-dir .git
egrep '\[[^"]' -I --exclude-dir .git */dirty.tsv */*/dirty.tsv
egrep '[^"]\]' -I --exclude-dir .git */dirty.tsv */*/dirty.tsv
egrep '	.*/' -I --exclude-dir .git */dirty.tsv */*/dirty.tsv
rgrep ' ' --exclude-dir .git | grep --invert-match 'README.md' | grep --invert-match chinese | grep --invert-match japanese | grep --invert-match tibetan | grep --invert-match english/multi
egrep '[[:upper:]]' */dirty.tsv */*/dirty.tsv
