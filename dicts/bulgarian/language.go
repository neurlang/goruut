package bulgarian

import "embed"

//go:embed missing* language.json weights*.json.zlib language_reverse.json weights*.bin.zlib weights*.bin.zlib
var Language embed.FS
