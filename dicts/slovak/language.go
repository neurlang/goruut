package slovak

import "embed"

//go:embed *.tsv language.json weights1.json.lzw language_reverse.json weights1_reverse.json.lzw
var Language embed.FS
