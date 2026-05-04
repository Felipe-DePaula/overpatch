# Browser Provider (experimental)

Two bookmarklets that turn a chat UI (ChatGPT, Claude.ai) into a manual Overpatch provider.

## Files

- `inject-prompt.src.js` — readable source for the prompt injector
- `inject-prompt.bookmarklet.js` — minified, `javascript:`-prefixed, ready to paste into a bookmark
- `extract-response.src.js` — readable source for the response extractor
- `extract-response.bookmarklet.js` — minified bookmarklet version

## How it works

1. Open a chat UI. Attach the project dump as a file.
2. Click **inject-prompt** bookmark. It writes the Overpatch prompt template into the input box and sends.
3. Wait for the model's response.
4. Click **extract-response** bookmark. It locates the `AI_FINAL_OUTPUT_V1` marker, extracts the JSON, validates it parses, and copies to clipboard.
5. Paste into a file and run `overpatch validate <file>`.

## Status

**Experimental.** This is a research artifact, not a shipped feature. It exists for two reasons:

- Validating the protocol with real LLM outputs before automating provider integrations.
- Letting users with a Pro/Plus subscription drive Overpatch without API keys.

In v0.6, this becomes one adapter among several behind the Provider Gateway. The bookmarklets stay as the `manual` provider mode.

## Caveats

- Selectors are tied to the current ChatGPT DOM. They will break.
- ToS of chat providers may restrict automation. Use on your own responsibility.
- The clipboard step is manual. There is no security boundary here — the user copies the output into Overpatch explicitly.

## Editing

Edit the `.src.js` files. To regenerate bookmarklets, strip newlines, prefix with `javascript:`, and URL-encode where needed. A simple build script will live in `scripts/` once the toolchain warrants it.

For now, hand-maintain both. The pair is small.
