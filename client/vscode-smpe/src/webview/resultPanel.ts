/**
 * Webview Panel for z/OSMF Query Results
 */

import * as vscode from 'vscode';
import { DisplayResult, ZosmfEntry, ZosmfSubentry, ZosmfServer, Credentials } from '../zosmf/types';
import { ZosmfClient } from '../zosmf/client';
import { UssPanel } from './ussPanel';
import { DatasetPanel } from './datasetPanel';

export class ResultPanel {
    public static currentPanel: ResultPanel | undefined;
    private readonly panel: vscode.WebviewPanel;
    private readonly extensionUri: vscode.Uri;
    private disposables: vscode.Disposable[] = [];
    private currentResult: DisplayResult | undefined;
    private zosmfClient: ZosmfClient | undefined;
    private zosmfServer: ZosmfServer | undefined;
    private zosmfCredentials: Credentials | undefined;

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri) {
        this.panel = panel;
        this.extensionUri = extensionUri;

        this.panel.onDidDispose(() => this.dispose(), null, this.disposables);

        // Handle messages from webview
        this.panel.webview.onDidReceiveMessage(
            message => {
                switch (message.command) {
                    case 'export':
                        this.exportResults(message.format);
                        break;
                    case 'copy':
                        this.copyToClipboard(message.data);
                        break;
                    case 'openUssPath':
                        this.openUssPath(message.path);
                        break;
                    case 'openDataset':
                        this.openDataset(message.dataset);
                        break;
                }
            },
            null,
            this.disposables
        );
    }

    public static createOrShow(extensionUri: vscode.Uri): ResultPanel {
        const column = vscode.ViewColumn.One;

        if (ResultPanel.currentPanel) {
            ResultPanel.currentPanel.panel.reveal(column);
            return ResultPanel.currentPanel;
        }

        const panel = vscode.window.createWebviewPanel(
            'smpeQueryResult',
            'SMP/E Query Result',
            column,
            {
                enableScripts: true,
                retainContextWhenHidden: true,
                localResourceRoots: [extensionUri]
            }
        );

        ResultPanel.currentPanel = new ResultPanel(panel, extensionUri);
        return ResultPanel.currentPanel;
    }

    /**
     * Set z/OSMF connection context for USS browsing
     */
    public setZosmfContext(client: ZosmfClient, server: ZosmfServer, credentials: Credentials): void {
        this.zosmfClient = client;
        this.zosmfServer = server;
        this.zosmfCredentials = credentials;
    }

    public showResult(result: DisplayResult): void {
        this.currentResult = result;
        this.panel.title = `SMP/E: ${this.getTitle(result)}`;
        this.panel.webview.html = this.getHtmlContent(result);
    }

    private getTitle(result: DisplayResult): string {
        switch (result.queryType) {
            case 'sysmod':
                return `SYSMODs (${result.serverName})`;
            case 'dddef':
                return `DDDEFs (${result.serverName})`;
            case 'zone':
                return `Zones (${result.serverName})`;
            default:
                return result.serverName;
        }
    }

    private async exportResults(format: 'json' | 'csv'): Promise<void> {
        if (!this.currentResult) {
            return;
        }

        const content = format === 'json'
            ? JSON.stringify(this.currentResult.result, null, 2)
            : this.convertToCsv(this.currentResult);

        const defaultName = `smpe-${this.currentResult.queryType}-${Date.now()}.${format}`;

        const uri = await vscode.window.showSaveDialog({
            defaultUri: vscode.Uri.file(defaultName),
            filters: format === 'json'
                ? { 'JSON': ['json'] }
                : { 'CSV': ['csv'] }
        });

        if (uri) {
            const encoder = new TextEncoder();
            await vscode.workspace.fs.writeFile(uri, encoder.encode(content));
            vscode.window.showInformationMessage(`Exported to ${uri.fsPath}`);
        }
    }

    private convertToCsv(result: DisplayResult): string {
        const lines: string[] = [];
        const entries = result.result.entries || [];

        if (entries.length === 0) {
            return '';
        }

        // Determine type from first entry
        const firstEntry = entries[0];
        if (firstEntry.entrytype === 'SYSMOD' || entries.some(e => e.entrytype === 'SYSMOD')) {
            lines.push('Zone,EntryName,Type,FMID,SMODTYPE,RECDATE,RECTIME,ERROR,RELATED');
            for (const entry of entries) {
                const subData = this.extractSubentryData(entry.subentries);
                lines.push([
                    this.escapeCsv(entry.zonename),
                    this.escapeCsv(entry.entryname),
                    this.escapeCsv(entry.entrytype),
                    this.escapeCsv(subData['FMID'] || ''),
                    this.escapeCsv(subData['SMODTYPE'] || ''),
                    this.escapeCsv(subData['RECDATE'] || ''),
                    this.escapeCsv(subData['RECTIME'] || ''),
                    this.escapeCsv(subData['ERROR'] || ''),
                    this.escapeCsv(subData['RELATED'] || '')
                ].join(','));
            }
        } else if (firstEntry.entrytype === 'DDDEF' || entries.some(e => e.entrytype === 'DDDEF')) {
            lines.push('Zone,DDNAME,DATASET,DISP,DATACLAS,MGMTCLAS,STORCLAS');
            for (const entry of entries) {
                const subData = this.extractSubentryData(entry.subentries);
                lines.push([
                    this.escapeCsv(entry.zonename),
                    this.escapeCsv(entry.entryname),
                    this.escapeCsv(subData['DATASET'] || ''),
                    this.escapeCsv(subData['DISP'] || ''),
                    this.escapeCsv(subData['DATACLAS'] || ''),
                    this.escapeCsv(subData['MGMTCLAS'] || ''),
                    this.escapeCsv(subData['STORCLAS'] || '')
                ].join(','));
            }
        } else if (firstEntry.entrytype === 'GLOBALZONE' || entries.some(e => e.entrytype === 'GLOBALZONE')) {
            lines.push('Zone,Type,Related');
            for (const entry of entries) {
                const subData = this.extractSubentryData(entry.subentries);
                const zoneIndex = subData['ZONEINDEX'] || '';
                // ZONEINDEX format: "ZONENAME,DSNAME,TYPE"
                if (zoneIndex) {
                    const zoneLines = zoneIndex.split('|');
                    for (const zl of zoneLines) {
                        const parts = zl.split(',');
                        if (parts.length >= 3) {
                            lines.push([
                                this.escapeCsv(parts[0]),
                                this.escapeCsv(parts[2]),
                                this.escapeCsv(parts[1])
                            ].join(','));
                        }
                    }
                }
            }
        }

        return lines.join('\n');
    }

    private extractSubentryData(subentries: ZosmfSubentry[]): Record<string, string> {
        const data: Record<string, string> = {};
        for (const sub of subentries) {
            for (const key of Object.keys(sub)) {
                if (key !== 'VER' && sub[key]) {
                    const value = sub[key];
                    if (Array.isArray(value)) {
                        // Handle nested objects in arrays (e.g. SUPBY: [{"SYSMOD": "HBB77E0"}])
                        const flat = value.map(item => {
                            if (typeof item === 'object' && item !== null) {
                                return Object.values(item).flat().join(',');
                            }
                            return String(item);
                        });
                        data[key] = flat.join('|');
                    }
                }
            }
        }
        return data;
    }

    private escapeCsv(value: string): string {
        if (value.includes(',') || value.includes('"') || value.includes('\n')) {
            return `"${value.replace(/"/g, '""')}"`;
        }
        return value;
    }

    private async copyToClipboard(data: string): Promise<void> {
        await vscode.env.clipboard.writeText(data);
        vscode.window.showInformationMessage('Copied to clipboard');
    }

    private async openUssPath(ussPath: string): Promise<void> {
        if (!this.zosmfClient || !this.zosmfServer || !this.zosmfCredentials) {
            vscode.window.showWarningMessage('No z/OSMF connection available. Please run a query first.');
            return;
        }
        await UssPanel.open(this.zosmfClient, this.zosmfServer, this.zosmfCredentials, ussPath);
    }

    private async openDataset(datasetName: string): Promise<void> {
        if (!this.zosmfClient || !this.zosmfServer || !this.zosmfCredentials) {
            vscode.window.showWarningMessage('No z/OSMF connection available. Please run a query first.');
            return;
        }
        await DatasetPanel.open(this.zosmfClient, this.zosmfServer, this.zosmfCredentials, datasetName);
    }

    private getHtmlContent(result: DisplayResult): string {
        const nonce = this.getNonce();
        const entries = result.result.entries || [];

        let tableHtml: string;
        if (entries.length === 0) {
            tableHtml = '<p class="no-results">No results found</p>';
        } else {
            tableHtml = this.renderEntriesTable(entries, result.queryType, result.requestedIds);
        }

        const messagesHtml = result.result.messages && result.result.messages.length > 0
            ? `<div class="messages">${result.result.messages.map(m => `<p>${this.escapeHtml(m)}</p>`).join('')}</div>`
            : '';

        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'nonce-${nonce}'; script-src 'nonce-${nonce}';">
    <title>SMP/E Query Result</title>
    <style nonce="${nonce}">
        :root {
            --vscode-font-family: var(--vscode-editor-font-family, monospace);
        }
        body {
            font-family: var(--vscode-font-family);
            font-size: var(--vscode-font-size);
            color: var(--vscode-foreground);
            background-color: var(--vscode-editor-background);
            padding: 16px;
            margin: 0;
        }
        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 16px;
            padding-bottom: 8px;
            border-bottom: 1px solid var(--vscode-panel-border);
        }
        .header h2 {
            margin: 0;
            color: var(--vscode-foreground);
        }
        .header-info {
            font-size: 0.9em;
            color: var(--vscode-descriptionForeground);
        }
        .toolbar {
            display: flex;
            gap: 8px;
        }
        .toolbar button {
            background-color: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 6px 12px;
            cursor: pointer;
            border-radius: 2px;
        }
        .toolbar button:hover {
            background-color: var(--vscode-button-hoverBackground);
        }
        .table-scroll {
            overflow-x: auto;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            font-size: 0.9em;
        }
        th, td {
            text-align: left;
            padding: 8px;
            border-bottom: 1px solid var(--vscode-panel-border);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            max-width: 300px;
        }
        th {
            background-color: var(--vscode-keybindingTable-headerBackground, rgba(128, 128, 128, 0.15));
            color: var(--vscode-foreground);
            font-weight: 600;
            position: sticky;
            top: 0;
        }
        tbody tr:nth-child(odd) {
            background-color: var(--vscode-keybindingTable-rowsBackground, rgba(128, 128, 128, 0.04));
        }
        tbody tr:hover {
            background-color: var(--vscode-list-hoverBackground);
        }
        .entry-sysmod {
            color: var(--vscode-symbolIcon-functionForeground, #dcdcaa);
        }
        .entry-targetzone {
            color: var(--vscode-symbolIcon-classForeground, #4ec9b0);
        }
        .entry-dddef {
            color: var(--vscode-symbolIcon-variableForeground, #9cdcfe);
        }
        .entry-globalzone {
            color: var(--vscode-symbolIcon-namespaceForeground, #c586c0);
        }
        .error-yes {
            color: var(--vscode-testing-iconFailed, #f14c4c);
        }
        .error-no {
            color: var(--vscode-testing-iconPassed, #73c991);
        }
        .messages {
            margin-top: 16px;
            padding: 8px;
            background-color: var(--vscode-inputValidation-infoBackground);
            border: 1px solid var(--vscode-inputValidation-infoBorder);
            border-radius: 4px;
        }
        .messages p {
            margin: 4px 0;
        }
        .no-results {
            text-align: center;
            padding: 32px;
            color: var(--vscode-descriptionForeground);
        }
        .subentry-info {
            font-size: 0.85em;
            color: var(--vscode-descriptionForeground);
        }
        .count-badge {
            background-color: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            padding: 2px 6px;
            border-radius: 10px;
            font-size: 0.8em;
            margin-left: 8px;
        }
        .cell-tooltip {
            display: none;
            position: fixed;
            background-color: var(--vscode-editorHoverWidget-background, var(--vscode-editor-background));
            color: var(--vscode-editorHoverWidget-foreground, var(--vscode-foreground));
            border: 1px solid var(--vscode-editorHoverWidget-border, var(--vscode-panel-border));
            padding: 4px 8px;
            border-radius: 3px;
            font-size: var(--vscode-font-size);
            max-width: 600px;
            white-space: pre-wrap;
            word-break: break-all;
            z-index: 1000;
            pointer-events: none;
            box-shadow: 0 2px 8px rgba(0,0,0,0.2);
        }
        .cell-tooltip.visible {
            display: block;
        }
        .uss-link, .ds-link {
            color: var(--vscode-textLink-foreground);
            text-decoration: none;
            cursor: pointer;
        }
        .uss-link:hover, .ds-link:hover {
            text-decoration: underline;
            color: var(--vscode-textLink-activeForeground);
        }
    </style>
</head>
<body>
    <div class="header">
        <div>
            <h2>${this.escapeHtml(this.getTitle(result))}<span class="count-badge">${entries.length} entries</span></h2>
            <div class="header-info">
                Server: ${this.escapeHtml(result.serverName)} |
                Time: ${result.timestamp.toLocaleString()}
            </div>
        </div>
        <div class="toolbar">
            <button id="exportJsonBtn">Export JSON</button>
            <button id="exportCsvBtn">Export CSV</button>
        </div>
    </div>
    <div id="cellTooltip" class="cell-tooltip"></div>
    <div class="table-scroll">${tableHtml}</div>
    ${messagesHtml}
    <script nonce="${nonce}">
        const vscode = acquireVsCodeApi();

        function exportJson() {
            vscode.postMessage({ command: 'export', format: 'json' });
        }

        function exportCsv() {
            vscode.postMessage({ command: 'export', format: 'csv' });
        }

        function copyText(text) {
            vscode.postMessage({ command: 'copy', data: text });
        }

        // Attach event listeners (onclick attributes are blocked by CSP)
        document.getElementById('exportJsonBtn').addEventListener('click', exportJson);
        document.getElementById('exportCsvBtn').addEventListener('click', exportCsv);

        // Handle USS path and dataset links
        document.addEventListener('click', (e) => {
            const ussLink = e.target.closest('.uss-link');
            if (ussLink) {
                e.preventDefault();
                vscode.postMessage({ command: 'openUssPath', path: ussLink.dataset.path });
                return;
            }
            const dsLink = e.target.closest('.ds-link');
            if (dsLink) {
                e.preventDefault();
                vscode.postMessage({ command: 'openDataset', dataset: dsLink.dataset.dataset });
            }
        });

        // Custom tooltip for truncated cells
        const tooltip = document.getElementById('cellTooltip');
        let tooltipTimeout = null;
        document.addEventListener('mouseover', (e) => {
            const td = e.target.closest('td');
            if (!td) { return; }
            const text = td.textContent || '';
            if (td.scrollWidth <= td.clientWidth && text.length <= 40) { return; }
            clearTimeout(tooltipTimeout);
            tooltip.textContent = text;
            tooltip.classList.add('visible');
            const rect = td.getBoundingClientRect();
            tooltip.style.left = rect.left + 'px';
            tooltip.style.top = (rect.bottom + 4) + 'px';
        });
        document.addEventListener('mouseout', (e) => {
            const td = e.target.closest('td');
            if (!td) { return; }
            tooltipTimeout = setTimeout(() => {
                tooltip.classList.remove('visible');
            }, 100);
        });
    </script>
</body>
</html>`;
    }

    private renderEntriesTable(entries: ZosmfEntry[], queryType: string, requestedIds?: string[]): string {
        // Determine columns based on query type and entry types
        if (queryType === 'zone') {
            return this.renderZoneIndexTable(entries);
        } else if (queryType === 'dddef') {
            return this.renderDddefTable(entries, requestedIds);
        } else {
            return this.renderSysmodTable(entries, requestedIds);
        }
    }

    private renderSysmodTable(entries: ZosmfEntry[], requestedIds?: string[]): string {
        const sorted = [...entries].sort((a, b) =>
            a.zonename.localeCompare(b.zonename) || a.entryname.localeCompare(b.entryname)
        );

        // Find SYSMODs that were requested but not found in results
        const foundNames = new Set(sorted.filter(e => e.entrytype === 'SYSMOD').map(e => e.entryname));
        const dot = '.';

        const rows = sorted.map(entry => {
            const subData = this.extractSubentryData(entry.subentries);
            const entryClass = `entry-${entry.entrytype.toLowerCase()}`;
            const errorClass = subData['ERROR'] === 'NO' ? 'error-no' : (subData['ERROR'] ? 'error-yes' : '');

            return `<tr>
                <td>${this.escapeHtml(entry.zonename)}</td>
                <td class="${entryClass}">${this.escapeHtml(entry.entryname)}</td>
                <td>${this.escapeHtml(entry.entrytype)}</td>
                <td>${this.escapeHtml(subData['SMODTYPE'] || subData['RELATED'] || '')}</td>
                <td>${this.escapeHtml(subData['FMID'] || '')}</td>
                <td>${this.escapeHtml(subData['RECDATE'] || '')}</td>
                <td>${this.escapeHtml(subData['RECTIME'] || '')}</td>
                <td class="${errorClass}">${this.escapeHtml(subData['ERROR'] || '')}</td>
                <td>${this.escapeHtml(subData['REWORK'] || '')}</td>
            </tr>`;
        }).join('');

        // Add rows for requested SYSMODs not found in results
        const zones = [...new Set(sorted.map(e => e.zonename))].sort();
        const zoneName = zones.length > 0 ? zones.join(', ') : dot;
        const missingRows = (requestedIds || [])
            .filter(id => !foundNames.has(id))
            .sort()
            .map(id => `<tr>
                <td>${this.escapeHtml(zoneName)}</td>
                <td>${this.escapeHtml(id)}</td>
                <td>${dot}</td>
                <td>${dot}</td>
                <td>${dot}</td>
                <td>${dot}</td>
                <td>${dot}</td>
                <td>${dot}</td>
                <td>${dot}</td>
            </tr>`)
            .join('');

        return `<table>
            <thead>
                <tr>
                    <th>Zone</th>
                    <th>Entry</th>
                    <th>Type</th>
                    <th>SMODTYPE/Related</th>
                    <th>FMID</th>
                    <th>RECDATE</th>
                    <th>RECTIME</th>
                    <th>ERROR</th>
                    <th>REWORK</th>
                </tr>
            </thead>
            <tbody>${rows}${missingRows}</tbody>
        </table>`;
    }

    private renderDddefTable(entries: ZosmfEntry[], requestedIds?: string[]): string {
        const filtered = entries.filter(e => e.entrytype === 'DDDEF').sort((a, b) =>
            a.zonename.localeCompare(b.zonename) || a.entryname.localeCompare(b.entryname)
        );
        const hasZosmf = !!(this.zosmfClient && this.zosmfServer && this.zosmfCredentials);
        const rows = filtered.map(entry => {
            const subData = this.extractSubentryData(entry.subentries);
            const ussPath = subData['PATH'] || '';
            const datasetName = subData['DATASET'] || '';
            const isUssPath = hasZosmf && ussPath.startsWith('/');
            const isDataset = hasZosmf && datasetName.length > 0;
            const pathCell = isUssPath
                ? `<td><a href="#" class="uss-link" data-path="${this.escapeHtml(ussPath)}">${this.escapeHtml(ussPath)}</a></td>`
                : `<td>${this.escapeHtml(ussPath)}</td>`;
            const datasetCell = isDataset
                ? `<td><a href="#" class="ds-link" data-dataset="${this.escapeHtml(datasetName)}">${this.escapeHtml(datasetName)}</a></td>`
                : `<td>${this.escapeHtml(datasetName)}</td>`;

            return `<tr>
                <td>${this.escapeHtml(entry.zonename)}</td>
                <td class="entry-dddef">${this.escapeHtml(entry.entryname)}</td>
                ${datasetCell}
                ${pathCell}
                <td>${this.escapeHtml(subData['INITDISP'] || '')}</td>
                <td>${this.escapeHtml(subData['DISP'] || '')}</td>
                <td>${this.escapeHtml(subData['DATACLAS'] || '')}</td>
                <td>${this.escapeHtml(subData['MGMTCLAS'] || '')}</td>
                <td>${this.escapeHtml(subData['STORCLAS'] || '')}</td>
            </tr>`;
        }).join('');

        const foundNames = new Set(filtered.map(e => e.entryname));
        const dot = '.';
        const zones = [...new Set(filtered.map(e => e.zonename))].sort();
        const zoneName = zones.length > 0 ? zones.join(', ') : dot;
        const missingRows = (requestedIds || [])
            .filter(id => !foundNames.has(id))
            .sort()
            .map(id => `<tr>
                <td>${this.escapeHtml(zoneName)}</td>
                <td>${this.escapeHtml(id)}</td>
                <td>${dot}</td>
                <td>${dot}</td>
                <td>${dot}</td>
                <td>${dot}</td>
                <td>${dot}</td>
                <td>${dot}</td>
                <td>${dot}</td>
            </tr>`)
            .join('');

        return `<table>
            <thead>
                <tr>
                    <th>Zone</th>
                    <th>DDNAME</th>
                    <th>DATASET</th>
                    <th>PATH</th>
                    <th>INITDISP</th>
                    <th>DISP</th>
                    <th>DATACLAS</th>
                    <th>MGMTCLAS</th>
                    <th>STORCLAS</th>
                </tr>
            </thead>
            <tbody>${rows}${missingRows}</tbody>
        </table>`;
    }

    private renderZoneIndexTable(entries: ZosmfEntry[]): string {
        const rows: string[] = [];

        for (const entry of entries) {
            if (entry.entrytype === 'GLOBALZONE') {
                const subData = this.extractSubentryData(entry.subentries);
                const zoneIndexStr = subData['ZONEINDEX'] || '';
                // ZONEINDEX is an array of strings like "ZONENAME,CSI_DSN,TYPE"
                const zoneIndexParts = zoneIndexStr.split('|');
                for (const zi of zoneIndexParts) {
                    const parts = zi.split(',');
                    if (parts.length >= 3) {
                        const zoneName = parts[0];
                        const csiDsn = parts[1];
                        const zoneType = parts[2];
                        rows.push(`<tr>
                            <td class="entry-globalzone">${this.escapeHtml(zoneName)}</td>
                            <td>${this.escapeHtml(zoneType)}</td>
                            <td>${this.escapeHtml(csiDsn)}</td>
                        </tr>`);
                    }
                }
            }
        }

        if (rows.length === 0) {
            return '<p class="no-results">No zone index entries found</p>';
        }

        return `<table>
            <thead>
                <tr>
                    <th>Zone</th>
                    <th>Type</th>
                    <th>CSI Dataset</th>
                </tr>
            </thead>
            <tbody>${rows.join('')}</tbody>
        </table>`;
    }

    private escapeHtml(text: string): string {
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }

    private getNonce(): string {
        let text = '';
        const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        for (let i = 0; i < 32; i++) {
            text += possible.charAt(Math.floor(Math.random() * possible.length));
        }
        return text;
    }

    public dispose(): void {
        ResultPanel.currentPanel = undefined;
        this.panel.dispose();
        while (this.disposables.length) {
            const disposable = this.disposables.pop();
            if (disposable) {
                disposable.dispose();
            }
        }
    }
}
