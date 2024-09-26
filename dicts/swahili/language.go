package swahili

import "embed"

//go:embed *.tsv language.json weights1.json.lzw
var Language embed.FS
