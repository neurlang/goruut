package russian

import "embed"

//go:embed *.tsv language.json weights1.json.zlib language_reverse.json weights1_reverse.json.zlib
var Language embed.FS
