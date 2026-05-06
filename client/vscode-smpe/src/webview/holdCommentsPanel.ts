/**
 * HOLD Comments Panel
 * Shows HOLD comments extracted from a PTF member in SMPPTS
 */

import * as vscode from 'vscode';

export interface HoldBlock {
    sysmodeId: string;
    holdType: string;
    fmid: string;
    reason: string;
    comment: string;
    raw: string;
}

export class HoldCommentsPanel {
    public static currentPanel: HoldCommentsPanel | undefined;
    private readonly panel: vscode.WebviewPanel;
    private disposables: vscode.Disposable[] = [];

    private constructor(panel: vscode.WebviewPanel) {
        this.panel = panel;
        this.panel.onDidDispose(() => this.dispose(), null, this.disposables);
    }

    public static open(sysmodeId: string, holds: HoldBlock[]): void {
        const column = vscode.ViewColumn.Beside;

        if (HoldCommentsPanel.currentPanel) {
            HoldCommentsPanel.currentPanel.panel.reveal(column);
            HoldCommentsPanel.currentPanel.update(sysmodeId, holds);
            return;
        }

        const panel = vscode.window.createWebviewPanel(
            'smpeHoldComments',
            `HOLD Comments: ${sysmodeId}`,
            column,
            { enableScripts: false }
        );

        HoldCommentsPanel.currentPanel = new HoldCommentsPanel(panel);
        HoldCommentsPanel.currentPanel.update(sysmodeId, holds);
    }

    private update(sysmodeId: string, holds: HoldBlock[]): void {
        this.panel.title = `HOLD Comments: ${sysmodeId}`;
        this.panel.webview.html = this.getHtml(sysmodeId, holds);
    }

    private getHtml(sysmodeId: string, holds: HoldBlock[]): string {
        if (holds.length === 0) {
            return `<!DOCTYPE html><html><body>
                <p style="font-family:var(--vscode-font-family);color:var(--vscode-foreground);padding:16px;">
                    No HOLD comments found for ${escapeHtml(sysmodeId)}.
                </p>
            </body></html>`;
        }

        let blocksHtml = '';
        for (const h of holds) {
            blocksHtml += `
            <div class="hold-block">
                <div class="hold-header">
                    <span class="hold-type ${h.holdType.toLowerCase()}">${escapeHtml(h.holdType)}</span>
                    <span class="hold-meta">FMID: <strong>${escapeHtml(h.fmid)}</strong></span>
                    <span class="hold-meta">REASON: <strong>${escapeHtml(h.reason)}</strong></span>
                </div>
                <pre class="comment-text">${escapeHtml(h.raw.trim())}</pre>
            </div>`;
        }

        return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline';">
<style>
    body {
        font-family: var(--vscode-font-family);
        font-size: var(--vscode-font-size);
        color: var(--vscode-foreground);
        background: var(--vscode-editor-background);
        padding: 16px;
        margin: 0;
    }
    h2 {
        color: var(--vscode-foreground);
        font-size: 1.1em;
        margin: 0 0 16px 0;
        border-bottom: 1px solid var(--vscode-panel-border);
        padding-bottom: 8px;
    }
    .hold-block {
        border: 1px solid var(--vscode-panel-border);
        border-radius: 4px;
        margin-bottom: 16px;
        overflow: hidden;
    }
    .hold-header {
        background: var(--vscode-editor-inactiveSelectionBackground);
        padding: 8px 12px;
        display: flex;
        align-items: center;
        gap: 16px;
        flex-wrap: wrap;
    }
    .hold-type {
        font-weight: bold;
        padding: 2px 8px;
        border-radius: 3px;
        font-size: 0.85em;
        text-transform: uppercase;
    }
    .hold-type.error   { background: var(--vscode-inputValidation-errorBackground); color: var(--vscode-inputValidation-errorForeground, #fff); }
    .hold-type.system  { background: var(--vscode-inputValidation-warningBackground); color: var(--vscode-inputValidation-warningForeground, #000); }
    .hold-type.fixcat  { background: var(--vscode-inputValidation-infoBackground); color: var(--vscode-inputValidation-infoForeground, #fff); }
    .hold-type.user    { background: var(--vscode-badge-background); color: var(--vscode-badge-foreground); }
    .hold-meta {
        font-size: 0.9em;
        color: var(--vscode-descriptionForeground);
    }
    .hold-meta strong {
        color: var(--vscode-foreground);
    }
    .comment-text {
        margin: 0;
        padding: 12px;
        white-space: pre-wrap;
        word-break: break-word;
        font-family: var(--vscode-editor-font-family, monospace);
        font-size: var(--vscode-editor-font-size);
        line-height: 1.5;
        background: var(--vscode-editor-background);
    }
    .no-comment {
        margin: 0;
        padding: 12px;
        color: var(--vscode-descriptionForeground);
        font-style: italic;
    }
</style>
</head>
<body>
    <h2>HOLD Comments — ${escapeHtml(sysmodeId)}</h2>
    ${blocksHtml}
</body>
</html>`;
    }

    private dispose(): void {
        HoldCommentsPanel.currentPanel = undefined;
        this.panel.dispose();
        for (const d of this.disposables) { d.dispose(); }
        this.disposables = [];
    }
}

function escapeHtml(s: string): string {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

/**
 * Parse ++HOLD blocks from PTS member text using regex.
 * Handles both COMMENT( on same line and COMMENT\n( on next line.
 */
export function parseHoldComments(text: string, sysmodeId: string): HoldBlock[] {
    // Normalize line endings
    const normalized = text.replace(/\r\n/g, '\n').replace(/\r/g, '\n');

    // Split on any MCS statement boundary (lines starting with ++)
    const lines = normalized.split('\n');
    const holdBlocks: string[] = [];
    let current: string[] | null = null;
    for (const line of lines) {
        if (/^\+\+HOLD\b/i.test(line.trimStart())) {
            if (current) { holdBlocks.push(current.join('\n')); }
            current = [line];
        } else if (current && /^\+\+\S/i.test(line.trimStart())) {
            holdBlocks.push(current.join('\n'));
            current = null;
        } else if (current) {
            current.push(line);
        }
    }
    if (current) { holdBlocks.push(current.join('\n')); }

    const result: HoldBlock[] = [];

    for (const block of holdBlocks) {
        // Extract hold type: ERROR|ERR, SYSTEM|SYS, FIXCAT|F6, USER
        let holdType = '';
        if (/\bFIXCAT\b|\bF6\b/i.test(block))       { holdType = 'FIXCAT'; }
        else if (/\bERROR\b|\bERR\b/i.test(block))   { holdType = 'ERROR'; }
        else if (/\bSYSTEM\b|\bSYS\b/i.test(block))  { holdType = 'SYSTEM'; }
        else if (/\bUSER\b/i.test(block))             { holdType = 'USER'; }

        // Extract FMID
        const fmidMatch = block.match(/\bFMID\s*\(\s*([^)]+)\s*\)/i);
        const fmid = fmidMatch ? fmidMatch[1].trim() : '';

        // Extract REASON
        const reasonMatch = block.match(/\bREASON\s*\(\s*([^)]+)\s*\)/i);
        const reason = reasonMatch ? reasonMatch[1].trim() : '';

        // Extract COMMENT content via paren-counting
        let comment = '';
        const commentStartMatch = block.match(/\bCOMMENT\s*[\s\S]*?\(/);
        if (commentStartMatch && commentStartMatch.index !== undefined) {
            const openIdx = commentStartMatch.index + commentStartMatch[0].length;
            let depth = 1;
            let i = openIdx;
            while (i < block.length && depth > 0) {
                if (block[i] === '(') { depth++; }
                else if (block[i] === ')') { depth--; }
                if (depth > 0) { i++; }
            }
            comment = block.slice(openIdx, i).trim();
        }

        result.push({ sysmodeId, holdType, fmid, reason, comment, raw: block });
    }

    return result;
}
