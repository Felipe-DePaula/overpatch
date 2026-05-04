// Overpatch — Browser Provider — Prompt Injector (experimental)
// Pastes the Overpatch planning prompt into a chat UI and sends it.
// The user is expected to have already attached the project dump as a file.

(async () => {
  const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

  const editor =
    document.querySelector('#prompt-textarea') ||
    document.querySelector('div[contenteditable="true"]');

  if (!editor) {
    alert('Editor not found');
    return;
  }

  editor.focus();
  await sleep(200);

  const PROBLEM = 'Disable the login route temporarily, do not delete anything.';

  const PROMPT = [
    'You are a generator of JSON instructions for deterministic automation.',
    '',
    'TASK: Read the attached TXT dump of the project and solve the problem below.',
    '',
    `PROBLEM TO SOLVE: ${PROBLEM}`,
    '',
    'OUTPUT CONTRACT:',
    'Your response must begin with exactly this line:',
    'AI_FINAL_OUTPUT_V1',
    '',
    'On the next line, write ONLY a valid, minified JSON in a SINGLE LINE.',
    '',
    'FORBIDDEN:',
    '- writing anything before AI_FINAL_OUTPUT_V1',
    '- writing anything after the JSON',
    '- using markdown',
    '- explaining outside the JSON',
    '- line breaks inside the JSON',
    '- line breaks inside JSON strings',
    '- returning entire large files when a deterministic operation suffices',
    '',
    'REQUIRED FORMAT:',
    'AI_FINAL_OUTPUT_V1',
    '{"schema_version":"overpatch/v1","status":"success","reason":"","operations":[{"id":"op_001","action":"replace_lines","path":"relative/path/file.ext","find_lines":["line 1","line 2"],"replace_lines":["new line 1","new line 2"],"expected_occurrences":1}]}',
    '',
    'ALLOWED VALUES FOR "status":',
    '- "success": a solution was generated',
    '- "no_changes": dump analyzed, no change needed',
    '- "failed": could not produce a solution',
    '',
    'ALLOWED ACTIONS:',
    '- replace_text',
    '- replace_lines',
    '- insert_before_lines',
    '- insert_after_lines',
    '- create',
    '- delete',
    '',
    'GENERAL RULES:',
    '- If status is "success", operations must contain at least 1 item.',
    '- If status is "no_changes", use operations:[] and explain in reason.',
    '- If status is "failed", use operations:[] and explain in reason.',
    '- Never use "success" with operations:[].',
    '- Do not invent missing files.',
    '- Do not invent requirements.',
    '- If the problem is empty, generic, or ambiguous, respond with status:"failed".',
    '',
    'CRITICAL JSON RULES:',
    '- Output must pass JSON.parse with no manual correction.',
    '- JSON must be minified on a single line.',
    '- No line breaks inside the JSON.',
    '- No comments.',
    '- No trailing commas.',
    '- Escape internal double quotes as \\".',
    '- AI_FINAL_OUTPUT_V1 marker must appear exactly once, at the start.',
    '',
    'Now analyze the attached dump and emit only the contracted output.',
  ].join('\n');

  document.execCommand('insertText', false, PROMPT);
  await sleep(800);

  const sendButton =
    document.querySelector('button[data-testid="send-button"]') ||
    document.querySelector('button[aria-label*="Send" i]') ||
    document.querySelector('button[aria-label*="Enviar" i]');

  if (!sendButton) {
    alert('Send button not found');
    return;
  }

  sendButton.click();
  console.log('Overpatch prompt sent.');
})();
