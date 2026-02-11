/**
 * Free Form Query Webview Panel
 * Combined input form and result table for z/OSMF CSI queries
 */

import * as vscode from 'vscode';
import { QueryProvider } from '../zosmf/queryProvider';
import { ZosmfServer, ZosmfEntry, ZosmfSubentry } from '../zosmf/types';

export class FreeFormPanel {
    public static currentPanel: FreeFormPanel | undefined;
    private readonly panel: vscode.WebviewPanel;
    private readonly queryProvider: QueryProvider;
    private readonly outputChannel: vscode.OutputChannel;
    private disposables: vscode.Disposable[] = [];

    private constructor(
        panel: vscode.WebviewPanel,
        queryProvider: QueryProvider,
        outputChannel: vscode.OutputChannel
    ) {
        this.panel = panel;
        this.queryProvider = queryProvider;
        this.outputChannel = outputChannel;

        this.panel.onDidDispose(() => this.dispose(), null, this.disposables);

        this.panel.webview.onDidReceiveMessage(
            message => this.handleMessage(message),
            null,
            this.disposables
        );
    }

    private log(message: string): void {
        const timestamp = new Date().toISOString();
        this.outputChannel.appendLine(`[${timestamp}] [FreeFormPanel] ${message}`);
    }

    public static createOrShow(
        queryProvider: QueryProvider,
        outputChannel: vscode.OutputChannel
    ): FreeFormPanel {
        const column = vscode.ViewColumn.One;

        if (FreeFormPanel.currentPanel) {
            FreeFormPanel.currentPanel.panel.reveal(column);
            return FreeFormPanel.currentPanel;
        }

        const panel = vscode.window.createWebviewPanel(
            'smpeFreeFormQuery',
            'SMP/E Free Form Query',
            column,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        FreeFormPanel.currentPanel = new FreeFormPanel(panel, queryProvider, outputChannel);
        FreeFormPanel.currentPanel.initPanel();
        return FreeFormPanel.currentPanel;
    }

    private initPanel(): void {
        this.panel.webview.html = this.getHtmlContent([], undefined);
        this.loadServers();
    }

    private loadServers(): void {
        const configManager = this.queryProvider.getConfigManager();
        const config = configManager.loadConfig();
        if (config) {
            const servers = config.servers.map(s => ({
                name: s.name,
                host: s.host,
                defaultZones: s.defaultZones || [],
                hasZonePatterns: !!(s.zones && s.zones.length > 0)
            }));
            this.panel.webview.postMessage({ command: 'servers', data: servers });
        }
    }

    private async handleMessage(message: any): Promise<void> {
        switch (message.command) {
            case 'executeQuery':
                await this.executeQuery(message);
                break;
            case 'export':
                await this.exportResults(message.format, message.data);
                break;
            case 'copy':
                await vscode.env.clipboard.writeText(message.data);
                vscode.window.showInformationMessage('Copied to clipboard');
                break;
        }
    }

    private async executeQuery(message: any): Promise<void> {
        const { serverName, zones, entryType, subentries, filter } = message;

        this.log(`Free form query: server=${serverName}, zones=${zones}, entry=${entryType}, subentries=${subentries}, filter=${filter}`);

        const configManager = this.queryProvider.getConfigManager();
        const config = configManager.loadConfig();
        if (!config) {
            this.panel.webview.postMessage({ command: 'error', message: 'Failed to load configuration' });
            return;
        }

        const server = config.servers.find(s => s.name === serverName);
        if (!server) {
            this.panel.webview.postMessage({ command: 'error', message: `Server '${serverName}' not found` });
            return;
        }

        const credentials = await configManager.getCredentials(server);
        if (!credentials) {
            this.panel.webview.postMessage({ command: 'error', message: 'Authentication cancelled' });
            return;
        }

        // Parse zone input and resolve patterns
        const zoneList = (zones as string).split(',').map((z: string) => z.trim().toUpperCase()).filter((z: string) => z.length > 0);
        const resolvedZones = this.queryProvider.resolveZonePatterns(server, zoneList);

        if (resolvedZones.length === 0) {
            this.panel.webview.postMessage({ command: 'error', message: 'No zones resolved from input' });
            return;
        }

        // Parse subentries
        const subentryList = (subentries as string).split(',').map((s: string) => s.trim().toUpperCase()).filter((s: string) => s.length > 0);

        this.panel.webview.postMessage({ command: 'progress', message: 'Sending query to z/OSMF...' });

        try {
            const client = this.queryProvider.getClient();
            const result = await client.queryFreeForm(
                server,
                credentials,
                resolvedZones,
                entryType.toUpperCase(),
                subentryList,
                filter || '',
                (msg) => {
                    this.panel.webview.postMessage({ command: 'progress', message: msg });
                }
            );

            this.log(`Query completed: ${(result.entries || []).length} entries`);
            this.panel.webview.postMessage({
                command: 'result',
                data: result,
                subentries: subentryList
            });
        } catch (error) {
            const msg = error instanceof Error ? error.message : String(error);
            this.log(`Query error: ${msg}`);
            this.panel.webview.postMessage({ command: 'error', message: msg });
        }
    }

    private async exportResults(format: 'json' | 'csv', data: any): Promise<void> {
        const content = format === 'json'
            ? JSON.stringify(data, null, 2)
            : data as string;

        const defaultName = `smpe-freeform-${Date.now()}.${format}`;
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

    private getHtmlContent(servers: any[], result: any): string {
        const nonce = this.getNonce();

        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'nonce-${nonce}'; script-src 'nonce-${nonce}';">
    <title>SMP/E Free Form Query</title>
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
        h2 {
            margin: 0 0 16px 0;
            color: var(--vscode-foreground);
        }
        .form-section {
            margin-bottom: 16px;
            padding: 12px;
            border: 1px solid var(--vscode-panel-border);
            border-radius: 4px;
        }
        .form-row {
            display: flex;
            align-items: center;
            margin-bottom: 8px;
            gap: 8px;
        }
        .form-row:last-child {
            margin-bottom: 0;
        }
        label {
            min-width: 100px;
            font-weight: 600;
            color: var(--vscode-foreground);
        }
        input, select {
            flex: 1;
            padding: 6px 8px;
            background-color: var(--vscode-input-background);
            color: var(--vscode-input-foreground);
            border: 1px solid var(--vscode-input-border, var(--vscode-panel-border));
            border-radius: 2px;
            font-family: var(--vscode-font-family);
            font-size: var(--vscode-font-size);
        }
        input:focus, select:focus {
            outline: 1px solid var(--vscode-focusBorder);
            border-color: var(--vscode-focusBorder);
        }
        input::placeholder {
            color: var(--vscode-input-placeholderForeground);
        }
        .button-row {
            display: flex;
            gap: 8px;
            margin-top: 12px;
        }
        button {
            background-color: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 8px 16px;
            cursor: pointer;
            border-radius: 2px;
            font-size: var(--vscode-font-size);
        }
        button:hover {
            background-color: var(--vscode-button-hoverBackground);
        }
        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
        button.secondary {
            background-color: var(--vscode-button-secondaryBackground);
            color: var(--vscode-button-secondaryForeground);
        }
        button.secondary:hover {
            background-color: var(--vscode-button-secondaryHoverBackground);
        }
        .result-section {
            margin-top: 16px;
        }
        .result-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
            padding-bottom: 8px;
            border-bottom: 1px solid var(--vscode-panel-border);
        }
        .result-header h3 {
            margin: 0;
        }
        .count-badge {
            background-color: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            padding: 2px 6px;
            border-radius: 10px;
            font-size: 0.8em;
            margin-left: 8px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            font-size: 0.9em;
        }
        th, td {
            text-align: left;
            padding: 6px 8px;
            border-bottom: 1px solid var(--vscode-panel-border);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            max-width: 300px;
        }
        th {
            background-color: var(--vscode-editor-background);
            color: var(--vscode-foreground);
            font-weight: 600;
            position: sticky;
            top: 0;
        }
        tr:hover {
            background-color: var(--vscode-list-hoverBackground);
        }
        .status-bar {
            margin-top: 8px;
            padding: 8px;
            font-size: 0.9em;
            color: var(--vscode-descriptionForeground);
        }
        .status-bar.error {
            color: var(--vscode-testing-iconFailed, #f14c4c);
        }
        .no-results {
            text-align: center;
            padding: 32px;
            color: var(--vscode-descriptionForeground);
        }
        .toolbar {
            display: flex;
            gap: 8px;
        }
        .hint {
            font-size: 0.85em;
            color: var(--vscode-descriptionForeground);
            margin-top: 2px;
        }
    </style>
</head>
<body>
    <h2>SMP/E Free Form CSI Query</h2>

    <div class="form-section">
        <div class="form-row">
            <label for="server">Server</label>
            <select id="server">
                <option value="">Loading...</option>
            </select>
        </div>
        <div class="form-row">
            <label for="zones">Zones</label>
            <input type="text" id="zones" placeholder="GLOBAL, MVS* (comma-separated, * and ? wildcards)" />
        </div>
        <div class="form-row">
            <label for="entryType">Entry Type</label>
            <select id="entryType">
                <option value="SYSMOD">SYSMOD</option>
                <option value="DDDEF">DDDEF</option>
                <option value="TARGETZONE">TARGETZONE</option>
                <option value="DLIB">DLIB</option>
                <option value="GLOBALZONE">GLOBALZONE</option>
                <option value="ASSEM">ASSEM</option>
                <option value="DATA">DATA</option>
                <option value="LMOD">LMOD</option>
                <option value="MAC">MAC</option>
                <option value="MOD">MOD</option>
                <option value="SRC">SRC</option>
            </select>
        </div>
        <div class="form-row">
            <label for="subentries">Subentries</label>
            <input type="text" id="subentries" placeholder="FMID,ERROR,RECDATE (comma-separated)" />
        </div>
        <div class="form-row">
            <label for="filter">Filter</label>
            <input type="text" id="filter" placeholder="ENAME='UA12345' (optional)" />
        </div>
        <div class="button-row">
            <button id="executeBtn" onclick="executeQuery()">Execute Query</button>
        </div>
    </div>

    <div id="statusBar" class="status-bar" style="display:none;"></div>

    <div id="resultSection" class="result-section" style="display:none;">
        <div class="result-header">
            <div>
                <h3>Results<span id="countBadge" class="count-badge">0</span></h3>
            </div>
            <div class="toolbar">
                <button class="secondary" onclick="exportJson()">Export JSON</button>
                <button class="secondary" onclick="exportCsv()">Export CSV</button>
            </div>
        </div>
        <div id="tableContainer"></div>
    </div>

    <script nonce="${nonce}">
        const vscode = acquireVsCodeApi();
        let currentResult = null;
        let currentSubentries = [];

        // Handle messages from extension
        window.addEventListener('message', event => {
            const message = event.data;
            switch (message.command) {
                case 'servers':
                    populateServers(message.data);
                    break;
                case 'result':
                    showResult(message.data, message.subentries);
                    break;
                case 'error':
                    showError(message.message);
                    break;
                case 'progress':
                    showProgress(message.message);
                    break;
            }
        });

        function populateServers(servers) {
            const select = document.getElementById('server');
            select.innerHTML = '';
            for (const s of servers) {
                const opt = document.createElement('option');
                opt.value = s.name;
                opt.textContent = s.name + ' (' + s.host + ')';
                select.appendChild(opt);
            }
            // Set default zones from first server
            if (servers.length > 0 && servers[0].defaultZones.length > 0) {
                document.getElementById('zones').value = servers[0].defaultZones.join(', ');
            }

            // Update zones when server changes
            select.addEventListener('change', () => {
                const selected = servers.find(s => s.name === select.value);
                if (selected && selected.defaultZones.length > 0) {
                    document.getElementById('zones').value = selected.defaultZones.join(', ');
                }
            });
        }

        function executeQuery() {
            const serverName = document.getElementById('server').value;
            const zones = document.getElementById('zones').value;
            const entryType = document.getElementById('entryType').value;
            const subentries = document.getElementById('subentries').value;
            const filter = document.getElementById('filter').value;

            if (!serverName) {
                showError('Please select a server');
                return;
            }
            if (!zones.trim()) {
                showError('Please enter at least one zone');
                return;
            }
            if (!subentries.trim()) {
                showError('Please enter at least one subentry');
                return;
            }

            document.getElementById('executeBtn').disabled = true;
            showProgress('Connecting...');

            vscode.postMessage({
                command: 'executeQuery',
                serverName: serverName,
                zones: zones,
                entryType: entryType,
                subentries: subentries,
                filter: filter
            });
        }

        function showResult(result, subentries) {
            currentResult = result;
            currentSubentries = subentries;
            document.getElementById('executeBtn').disabled = false;

            const entries = result.entries || [];
            document.getElementById('countBadge').textContent = entries.length;

            const statusBar = document.getElementById('statusBar');
            statusBar.style.display = 'block';
            statusBar.className = 'status-bar';
            statusBar.textContent = 'Query completed - ' + entries.length + ' entries returned';

            const resultSection = document.getElementById('resultSection');
            resultSection.style.display = 'block';

            const container = document.getElementById('tableContainer');

            if (entries.length === 0) {
                container.innerHTML = '<p class="no-results">No results found</p>';
                return;
            }

            // Build dynamic table
            let html = '<table><thead><tr>';
            html += '<th>Zone</th><th>Entry</th>';
            for (const sub of subentries) {
                html += '<th>' + escapeHtml(sub) + '</th>';
            }
            html += '</tr></thead><tbody>';

            for (const entry of entries) {
                html += '<tr>';
                html += '<td>' + escapeHtml(entry.zonename || '') + '</td>';
                html += '<td>' + escapeHtml(entry.entryname || '') + '</td>';

                const subData = extractSubentryData(entry.subentries || []);
                for (const sub of subentries) {
                    html += '<td>' + escapeHtml(subData[sub] || '') + '</td>';
                }
                html += '</tr>';
            }

            html += '</tbody></table>';
            container.innerHTML = html;
        }

        function extractSubentryData(subentries) {
            const data = {};
            for (const sub of subentries) {
                for (const key of Object.keys(sub)) {
                    if (key !== 'VER' && sub[key]) {
                        const value = sub[key];
                        if (Array.isArray(value)) {
                            data[key] = value.join(', ');
                        }
                    }
                }
            }
            return data;
        }

        function showError(msg) {
            document.getElementById('executeBtn').disabled = false;
            const statusBar = document.getElementById('statusBar');
            statusBar.style.display = 'block';
            statusBar.className = 'status-bar error';
            statusBar.textContent = 'Error: ' + msg;
        }

        function showProgress(msg) {
            const statusBar = document.getElementById('statusBar');
            statusBar.style.display = 'block';
            statusBar.className = 'status-bar';
            statusBar.textContent = msg;
        }

        function exportJson() {
            if (currentResult) {
                vscode.postMessage({ command: 'export', format: 'json', data: currentResult });
            }
        }

        function exportCsv() {
            if (currentResult && currentResult.entries) {
                let csv = 'Zone,Entry';
                for (const sub of currentSubentries) {
                    csv += ',' + sub;
                }
                csv += '\\n';

                for (const entry of currentResult.entries) {
                    const subData = extractSubentryData(entry.subentries || []);
                    csv += escapeCsv(entry.zonename || '') + ',' + escapeCsv(entry.entryname || '');
                    for (const sub of currentSubentries) {
                        csv += ',' + escapeCsv(subData[sub] || '');
                    }
                    csv += '\\n';
                }
                vscode.postMessage({ command: 'export', format: 'csv', data: csv });
            }
        }

        function escapeCsv(value) {
            if (value.includes(',') || value.includes('"') || value.includes('\\n')) {
                return '"' + value.replace(/"/g, '""') + '"';
            }
            return value;
        }

        function escapeHtml(text) {
            return text
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;');
        }
    </script>
</body>
</html>`;
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
        FreeFormPanel.currentPanel = undefined;
        this.panel.dispose();
        while (this.disposables.length) {
            const disposable = this.disposables.pop();
            if (disposable) {
                disposable.dispose();
            }
        }
    }
}
