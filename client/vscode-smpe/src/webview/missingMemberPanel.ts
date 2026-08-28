/**
 * Webview Panel for Missing Input Member Check Results
 */

import * as vscode from 'vscode';
import { MemberCheckResult } from '../workspace/missingMemberChecker';

export class MissingMemberPanel {
    public static currentPanel: MissingMemberPanel | undefined;
    private readonly panel: vscode.WebviewPanel;
    private readonly extensionUri: vscode.Uri;
    private disposables: vscode.Disposable[] = [];

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri) {
        this.panel = panel;
        this.extensionUri = extensionUri;
        this.panel.onDidDispose(() => this.dispose(), null, this.disposables);
    }

    public static createOrShow(extensionUri: vscode.Uri, results: MemberCheckResult[]): MissingMemberPanel {
        const column = vscode.ViewColumn.One;

        if (MissingMemberPanel.currentPanel) {
            MissingMemberPanel.currentPanel.panel.reveal(column);
            MissingMemberPanel.currentPanel.update(results);
            return MissingMemberPanel.currentPanel;
        }

        const panel = vscode.window.createWebviewPanel(
            'smpeMissingMembers',
            'SMP/E: Missing Input Members',
            column,
            {
                enableScripts: true,
                retainContextWhenHidden: true,
                localResourceRoots: [extensionUri]
            }
        );

        MissingMemberPanel.currentPanel = new MissingMemberPanel(panel, extensionUri);
        MissingMemberPanel.currentPanel.update(results);
        return MissingMemberPanel.currentPanel;
    }

    public update(results: MemberCheckResult[]): void {
        this.panel.webview.html = this.getHtml(results);
    }

    private dispose(): void {
        MissingMemberPanel.currentPanel = undefined;
        this.panel.dispose();
        for (const d of this.disposables) {
            d.dispose();
        }
        this.disposables = [];
    }

    private escapeHtml(text: string): string {
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;');
    }

    private getHtml(results: MemberCheckResult[]): string {
        const total = results.length;
        const missing = results.filter(r => !r.found).length;
        const found = total - missing;

        const rows = results.map(r => {
            const foundCell = r.found
                ? `<td class="found-yes" title="${this.escapeHtml(r.foundPath ?? '')}">Yes</td>`
                : `<td class="found-no">No</td>`;
            return `<tr data-smpe="${this.escapeHtml(r.smpeFile)}" data-stmt="${this.escapeHtml(r.statement)}" data-member="${this.escapeHtml(r.member)}" data-source="${this.escapeHtml(r.source)}" data-found="${r.found ? 'yes' : 'no'}">
                <td>${this.escapeHtml(r.smpeFile)}</td>
                <td>${this.escapeHtml(r.statement)}</td>
                <td>${this.escapeHtml(r.member)}</td>
                <td>${this.escapeHtml(r.source)}</td>
                ${foundCell}
            </tr>`;
        }).join('\n');

        return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>SMP/E Missing Input Members</title>
<style>
  body {
    font-family: var(--vscode-font-family);
    font-size: var(--vscode-font-size);
    color: var(--vscode-foreground);
    background-color: var(--vscode-editor-background);
    padding: 16px;
    margin: 0;
  }
  h1 {
    font-size: 1.2em;
    margin-bottom: 12px;
    color: var(--vscode-foreground);
  }
  .toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
    flex-wrap: wrap;
  }
  .toolbar input {
    background: var(--vscode-input-background);
    color: var(--vscode-input-foreground);
    border: 1px solid var(--vscode-input-border, #555);
    padding: 4px 8px;
    font-size: var(--vscode-font-size);
    border-radius: 2px;
    width: 220px;
  }
  .toolbar select {
    background: var(--vscode-dropdown-background);
    color: var(--vscode-dropdown-foreground);
    border: 1px solid var(--vscode-dropdown-border, #555);
    padding: 4px 6px;
    font-size: var(--vscode-font-size);
    border-radius: 2px;
  }
  .toolbar button {
    background: var(--vscode-button-background);
    color: var(--vscode-button-foreground);
    border: none;
    padding: 4px 10px;
    font-size: var(--vscode-font-size);
    border-radius: 2px;
    cursor: pointer;
  }
  .toolbar button:hover {
    background: var(--vscode-button-hoverBackground);
  }
  .summary {
    margin-left: auto;
    font-size: 0.9em;
    color: var(--vscode-descriptionForeground);
    white-space: nowrap;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--vscode-font-size);
  }
  thead th {
    background: var(--vscode-editorGroupHeader-tabsBackground);
    color: var(--vscode-foreground);
    text-align: left;
    padding: 6px 10px;
    border-bottom: 2px solid var(--vscode-panel-border, #444);
    cursor: pointer;
    user-select: none;
    white-space: nowrap;
  }
  thead th:hover {
    background: var(--vscode-list-hoverBackground);
  }
  thead th .sort-indicator {
    margin-left: 4px;
    opacity: 0.5;
    font-size: 0.8em;
  }
  thead th.sort-asc .sort-indicator::after { content: '▲'; opacity: 1; }
  thead th.sort-desc .sort-indicator::after { content: '▼'; opacity: 1; }
  thead th:not(.sort-asc):not(.sort-desc) .sort-indicator::after { content: '⇅'; }
  tbody tr {
    border-bottom: 1px solid var(--vscode-panel-border, #333);
  }
  tbody tr:hover {
    background: var(--vscode-list-hoverBackground);
  }
  tbody tr.hidden {
    display: none;
  }
  td {
    padding: 5px 10px;
    vertical-align: middle;
  }
  .found-yes {
    color: var(--vscode-testing-iconPassed, #4ec94e);
    font-weight: 500;
  }
  .found-no {
    color: var(--vscode-errorForeground, #f44747);
    font-weight: 700;
  }
  .no-results {
    padding: 20px;
    color: var(--vscode-descriptionForeground);
    text-align: center;
  }
</style>
</head>
<body>
<h1>SMP/E: Missing Input Members</h1>
<div class="toolbar">
  <input type="text" id="filter-input" placeholder="Filter..." oninput="applyFilter()">
  <label for="filter-found" style="font-size:0.9em">Show:</label>
  <select id="filter-found" onchange="applyFilter()">
    <option value="all">All</option>
    <option value="no">Missing only</option>
    <option value="yes">Found only</option>
  </select>
  <button onclick="clearFilter()">Clear</button>
  <span class="summary" id="summary">${total} items &mdash; ${found} found, ${missing} missing</span>
</div>
<table id="results-table">
  <thead>
    <tr>
      <th data-col="0" onclick="sortBy(0)">SMP/E File<span class="sort-indicator"></span></th>
      <th data-col="1" onclick="sortBy(1)">Statement<span class="sort-indicator"></span></th>
      <th data-col="2" onclick="sortBy(2)">Member<span class="sort-indicator"></span></th>
      <th data-col="3" onclick="sortBy(3)">Source<span class="sort-indicator"></span></th>
      <th data-col="4" onclick="sortBy(4)">Found<span class="sort-indicator"></span></th>
    </tr>
  </thead>
  <tbody id="results-body">
    ${rows || '<tr><td colspan="5" class="no-results">No statements to check.</td></tr>'}
  </tbody>
</table>
<script>
  let sortCol = -1;
  let sortAsc = true;

  function applyFilter() {
    const text = document.getElementById('filter-input').value.toLowerCase();
    const foundFilter = document.getElementById('filter-found').value;
    const rows = document.querySelectorAll('#results-body tr[data-smpe]');
    let visible = 0;

    rows.forEach(row => {
      const smpe = row.dataset.smpe.toLowerCase();
      const stmt = row.dataset.stmt.toLowerCase();
      const member = row.dataset.member.toLowerCase();
      const source = row.dataset.source.toLowerCase();
      const found = row.dataset.found;

      const textMatch = !text || smpe.includes(text) || stmt.includes(text) || member.includes(text) || source.includes(text);
      const foundMatch = foundFilter === 'all' || found === foundFilter;

      if (textMatch && foundMatch) {
        row.classList.remove('hidden');
        visible++;
      } else {
        row.classList.add('hidden');
      }
    });

    updateSummary(visible);
  }

  function clearFilter() {
    document.getElementById('filter-input').value = '';
    document.getElementById('filter-found').value = 'all';
    applyFilter();
  }

  function updateSummary(visible) {
    const rows = document.querySelectorAll('#results-body tr[data-smpe]');
    const total = rows.length;
    const missing = Array.from(rows).filter(r => r.dataset.found === 'no').length;
    const found = total - missing;
    const suffix = visible < total ? ' (filtered: ' + visible + ')' : '';
    document.getElementById('summary').innerHTML =
      total + ' items &mdash; ' + found + ' found, ' + missing + ' missing' + suffix;
  }

  function sortBy(col) {
    const ths = document.querySelectorAll('thead th');
    ths.forEach((th, i) => {
      th.classList.remove('sort-asc', 'sort-desc');
    });

    if (sortCol === col) {
      sortAsc = !sortAsc;
    } else {
      sortCol = col;
      sortAsc = true;
    }

    ths[col].classList.add(sortAsc ? 'sort-asc' : 'sort-desc');

    const tbody = document.getElementById('results-body');
    const rows = Array.from(tbody.querySelectorAll('tr[data-smpe]'));

    rows.sort((a, b) => {
      const aVal = a.cells[col].textContent.trim().toLowerCase();
      const bVal = b.cells[col].textContent.trim().toLowerCase();
      if (aVal < bVal) { return sortAsc ? -1 : 1; }
      if (aVal > bVal) { return sortAsc ? 1 : -1; }
      return 0;
    });

    rows.forEach(row => tbody.appendChild(row));
  }
</script>
</body>
</html>`;
    }
}
