# Roadmap for Goruut Phonemizer: Ensuring Long-Term Success

## Vision & Principles
Following Sutton's "Bitter Lesson," the project should emphasize general-purpose, computation-heavy methods over ad-hoc rules. In practice this means:
- Continually improving and scaling learning-based models
- Leveraging search algorithms ("search and learning" scale arbitrarily with compute)
- Building on Goruut's hybrid approach (lexical lookup plus neural G2P)
- Avoiding brittle hand-coded heuristics
- Favoring neural retraining and data-driven modeling as compute gets cheaper

## Current Status
- Go-based IPA phonemizer covering 139 languages (matching Voice2Json's list)
- Preserves punctuation, stress marks and tonal diacritics
- Fast performance (milliseconds per sentence) via prefix lookups and ML inference

**Recent Enhancements (v0.6.1, Apr 2025):**
- Retrained all forward/reverse G2P models
- Added transformer-based disambiguation for English homographs
- Numeric normalization for Arabic/English
- Improved Hebrew diacritic handling
- Overhauled Hebrew phonemization (sorting diacritics and grouping marks)

**Current Challenge:** Many languages remain at "beta" quality (rare/irregular scripts)

## Data & Models (Scale with Compute)
**Strategy:**
- Collect/generate large training sets
- Use powerful models (following DeepPhonemizer's transformer approach)
- Perform large-scale training runs on combined lexica from all languages
- Share parameters across related languages (e.g. Romance group)
- Use transfer learning
- Let learning discover pronunciations from data rather than hard-coding

**Example Success:** Used Goruut's output to fine-tune Whisper ASR model (15,000 synthetic IPA-labeled clips across 70+ languages)

## Enhancing Language Support
For each language:

### Expand Lexica
- Encourage community contributions
- Provide tools to import data from:
  - Wiktionary
  - panPhon/Epitran maps
  - UD lexicons
  - Existing speech corpora

### Documentation
- Polish language addition guide ("dict folder has a README")
- Create automated pipelines/templates for new languages

### Target Weak Languages
Priority areas:
- Semitic scripts (Hebrew, Arabic) - vowel handling
- Chinese - character segmentation
- Tonal languages - tone rules
- Automate diacritic sorting for other scripts
- Improve punctuation/number normalization globally

### Quality Testing
- Create benchmarks (possibly using SpeechSTx)
- Measure phoneme error rates by language
- Crowdsource evaluation via Discord/community

## Tools and Integrations
### TTS/STT Systems
- Feed into Coqui TTS, RHVoice, etc.
- Continue Coqui integration (idiap#88)
- Goal: Become standard G2P for open-source TTS toolkits

### Educational & Singing
- Language learning apps
- Singing synthesis (vocaloid systems)
- Document usage examples (REST API/CLI)

### API/Spaces
- Maintain public demo (hashtron.cloud)
- Keep Hugging Face Space (Pygoruut) active
- Active Discord for user support
- Improve documentation (README, wiki, examples)

## Open-Source Strategy
### Publish Roadmap
- Create "CONTRIBUTING" and "ROADMAP" docs
- Outline near-term (e.g. add Somali support) and long-term goals (e.g. end-to-end neural G2P)

### Community Engagement
- Label issues (e.g. "help wanted")
- Provide contribution templates
- Maintain good test coverage and CI

### Governance & Funding
- Consider minimal governance (list of maintainers)
- Explore funding via:
  - GitHub Sponsors
  - OpenCollective
- List funding options in README

## Long-Term Compute Strategy
- Leverage available GPUs fully
- Set up periodic retraining scripts
- Use transfer learning (start from multilingual models)
- Architect for scale (data-parallel training, cloud GPUs)
- Automate profiling to maintain fast inference

## Testing & Metrics
- Adopt automatic benchmarking
- Compute phoneme error rates (PER) on held-out word lists
- Regular comparisons against other tools (espeak-ng, phonemizer-Python)
- Include edge cases (numbers, foreign phrases)

## Summary
By adhering to:
1. General, scalable methods ("Bitter Lesson")
2. Open-roadmap process

Key steps:
- Expand/clean data
- Train stronger multilingual models
- Smooth language-specific quirks
- Foster contributor community

**Goal:** Continuously improve phonemization quality across all 139 languages for TTS/STT, education, singing synthesis, etc.

## Sources
1. [Goruut's github](https://github.com/neurlang/goruut)
2. [Goruut's discussions](https://github.com/coqui-ai/TTS/discussions/3794)
3. [Transformer G2P projects](https://github.com/spring-media/DeepPhonemizer)
4. [Sutton's "Bitter Lesson"](https://www.cs.utexas.edu/~eunsol/courses/data/bitter_lesson.pdf)
5. [Best practices for open-source roadmaps](https://contribute.cncf.io/maintainers/community/contributor-growth-framework/open-source-roadmaps/)
