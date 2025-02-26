package central

import "embed"

//go:embed *.tsv language.json weights*.json.zlib language_reverse.json weights*_reverse.json.zlib
var Language embed.FS
