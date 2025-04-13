# Developing the goruut Project

## Project Structure Overview

| Directory          | Purpose |
|--------------------|---------|
| `.github/workflows` | CI/CD workflows |
| `.vscode`          | VSCode config |
| `app`              | Main application logic |
| `cmd`              | CLI application source |
| `configs`          | Configuration files |
| `controllers`      | Application call controllers |
| `dicts`            | Language pronunciation dictionaries |
| `dicts_scripts`    | Dictionary management scripts |
| `doc`              | Documentation |
| `helpers`          | Utility functions |
| `lib`              | Library code for integration into 3rd party projects |
| `models`           | Data structures for controllers requests/responses |
| `repo`             | Main goruut application code (repository pattern) |
| `usecases`         | Application use cases |
| `views`            | HTTP UI components |

## Repository pattern

![Repo Pattern](https://miro.medium.com/v2/resize:fit:4800/format:webp/1*VdeHRnrn3us9WguuvneLDQ.png)

* Business Layer: Controllers -> UseCases -> Services
* Repository: Services -> Various AI models, Phonetic Dictionaries

## Adding a New Language

1. Create language folder in `dicts/`
2. Prepare `lexicon.tsv` (word/sentence to IPA mapping)
3. Create `language.json` with language-specific config
4. Run analysis scripts:
   
### Study prefix grammar
   `cmd/analysis2/study_language.sh [params]`
   
### Align words
   `cmd/dicttomap/clean_language.sh [params]`
   
### Train G2P model
   `train_language.sh [params]`

5. Add glue code (copy/modify existing `language.go`)
6. Modify `dicts.go` to include new language
7. Backtest the model:
   
   `cd cmd/backtest`
   `go build`
   `./backtest -langname <language>`

## Coding Style

Before submitting a pull request, please ensure your code meets the project's formatting requirements:

Format current directory from CLI:

`gofmt -w -s *.go`

## Enabling Debug

Add the following to `configs/config.json`:

`"Logging": {"Level":"trace"},`

Then run goruut with the modified config:

`./goruut --configfile ../../configs/config.json`

## Go Requirements

Yes, Go is required for training. The shell scripts are just wrappers around Go binaries for easier use.

## Lexicon Alignment

Typically 20%-80% of core `lexicon.tsv` gets aligned into `clean.tsv`. This depends on the regularity of the pronunciation:
- English (irregular pronunciation): ~50% alignment
- Uyghur (regular pronunciation): >94% alignment

## Homograph Training (multi.tsv)

**Format:** `sentence<tab>pronunciation`

1. Number of words in sentence and in pronunciation must match.
2. Words which are NOT homographs can utilize the `_` placeholder as their pronunciation.

**Sentence Length Guidelines:**
- Technically unlimited
- At 16 words (context window), words start wrapping in the homograph transformer
- **Recommended:** Keep sentences under 16 words

Note: Future versions may increase the transformer's context window if needed.

## Testing

Run all tests:

`go test ./...`

For verbose output:

`go test -v ./...`

More tests wanted!

## Contribution Workflow

1. Fork the repository
2. Clone your fork locally
3. Create a new branch:
   
`git checkout -b your-branch-name`

4. Make changes
5. Format code with `gofmt`
6. Test your changes
7. Push to your fork
8. Create a pull request

## Troubleshooting

### Common Issues
- **Compilation errors:** Check Go version (requires 1.19+)
- **Dependency issues:** Use `go mod tidy`
- **Local development:** Use `replace` in `go.mod` for local forks

### VSCode Users
- Restart Go language server (`gopls`) if experiencing issues

## Community

- **Issues/Bugs:** GitHub issue tracker
- **Discussions:** Discord (link in README)

When asking for help:
1. Check existing docs/issues first
2. Provide clear reproduction steps
3. Include expected vs actual behavior

## Non-Code Contributions

We welcome:
- Documentation improvements
- Bug reports
- Community discussions
- Translations
- Mentoring new contributors

Happy developing! We appreciate your contributions to goruut.
