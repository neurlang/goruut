package views

import "embed"

//go:embed v0/*.svg
//go:embed v0/*.html
//go:embed v0/*.css
//go:embed v0/*.js
//go:embed v0/*.xml
var Data embed.FS
