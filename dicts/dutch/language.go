package dutch

import "embed"

//go:embed *.tsv language.json weights0.json.gz
var Language embed.FS
