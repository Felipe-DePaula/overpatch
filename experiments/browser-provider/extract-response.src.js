// Overpatch — Browser Provider — Response Extractor (experimental)
// Locates the AI_FINAL_OUTPUT_V1 marker in the latest assistant message,
// extracts the JSON, validates it parses, and copies it to clipboard.

(async () => {
  try {
    const MARKER = 'AI_FINAL_OUTPUT_V1';
    const messages = document.querySelectorAll(
      '[data-message-author-role="assistant"]'
    );

    if (!messages.length) {
      alert('No assistant response found');
      return;
    }

    const lastMessage = messages[messages.length - 1];
    const text = lastMessage.textContent;
    const markerIdx = text.lastIndexOf(MARKER);

    if (markerIdx < 0) {
      alert('Marker AI_FINAL_OUTPUT_V1 not found');
      return;
    }

    const rest = text.slice(markerIdx + MARKER.length).trim();
    const jsonStart = rest.indexOf('{');

    if (jsonStart < 0) {
      alert('JSON not found');
      return;
    }

    // Brace-aware scanner that respects strings and escapes.
    let depth = 0;
    let jsonEnd = -1;
    let inString = false;
    let escaped = false;

    for (let i = jsonStart; i < rest.length; i++) {
      const ch = rest[i];
      if (escaped) {
        escaped = false;
        continue;
      }
      if (ch === '\\') {
        escaped = true;
        continue;
      }
      if (ch === '"') {
        inString = !inString;
        continue;
      }
      if (!inString && ch === '{') depth++;
      if (!inString && ch === '}') {
        depth--;
        if (depth === 0) {
          jsonEnd = i;
          break;
        }
      }
    }

    if (jsonEnd < 0) {
      alert('End of JSON not found');
      return;
    }

    const jsonStr = rest.slice(jsonStart, jsonEnd + 1);
    const parsed = JSON.parse(jsonStr);

    if (!parsed || !Array.isArray(parsed.operations)) {
      alert('Invalid JSON: no operations array');
      return;
    }

    await navigator.clipboard.writeText(jsonStr);
    console.log('Overpatch JSON extracted:', parsed);
    alert(
      `JSON copied — ${jsonStr.length} chars, ${parsed.operations.length} operation(s)`
    );
  } catch (e) {
    console.error(e);
    alert('Error: ' + e.message);
  }
})();
