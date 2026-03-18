/**
 * Webview Panel for USS Directory Listing via z/OSMF Files REST API
 */

import * as vscode from 'vscode';
import { ZosmfClient } from '../zosmf/client';
import { ZosmfServer, Credentials, UssEntry } from '../zosmf/types';

/**
 * Content provider for read-only USS file viewing
 */
export class UssFileContentProvider implements vscode.TextDocumentContentProvider {
    public static readonly scheme = 'smpe-uss';
    private contentMap = new Map<string, string>();
    private onDidChangeEmitter = new vscode.EventEmitter<vscode.Uri>();
    public onDidChange = this.onDidChangeEmitter.event;

    setContent(uri: vscode.Uri, content: string): void {
        this.contentMap.set(uri.toString(), content);
        this.onDidChangeEmitter.fire(uri);
    }

    provideTextDocumentContent(uri: vscode.Uri): string {
        return this.contentMap.get(uri.toString()) || '';
    }
}

export class UssPanel {
    public static currentPanel: UssPanel | undefined;
    public static contentProvider: UssFileContentProvider | undefined;
    private readonly panel: vscode.WebviewPanel;
    private disposables: vscode.Disposable[] = [];
    private client: ZosmfClient;
    private server: ZosmfServer;
    private credentials: Credentials;
    private currentPath: string = '';

    private constructor(
        panel: vscode.WebviewPanel,
        client: ZosmfClient,
        server: ZosmfServer,
        credentials: Credentials
    ) {
        this.panel = panel;
        this.client = client;
        this.server = server;
        this.credentials = credentials;

        this.panel.onDidDispose(() => this.dispose(), null, this.disposables);

        this.panel.webview.onDidReceiveMessage(
            async message => {
                switch (message.command) {
                    case 'navigate':
                        await this.loadDirectory(message.path);
                        break;
                    case 'openFile':
                        await this.openFile(message.path);
                        break;
                }
            },
            null,
            this.disposables
        );
    }

    public static async open(
        client: ZosmfClient,
        server: ZosmfServer,
        credentials: Credentials,
        ussPath: string
    ): Promise<void> {
        const column = vscode.ViewColumn.One;

        if (UssPanel.currentPanel) {
            try {
                UssPanel.currentPanel.client = client;
                UssPanel.currentPanel.server = server;
                UssPanel.currentPanel.credentials = credentials;
                UssPanel.currentPanel.panel.reveal(column);
                await UssPanel.currentPanel.loadDirectory(ussPath);
                return;
            } catch {
                // Panel was disposed — fall through to create a new one
                UssPanel.currentPanel = undefined;
            }
        }

        const panel = vscode.window.createWebviewPanel(
            'smpeUssBrowser',
            `USS: ${ussPath}`,
            column,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        UssPanel.currentPanel = new UssPanel(panel, client, server, credentials);
        await UssPanel.currentPanel.loadDirectory(ussPath);
    }

    private async loadDirectory(ussPath: string): Promise<void> {
        // Normalize path
        const normalizedPath = ussPath.replace(/\/+$/, '') || '/';
        this.currentPath = normalizedPath;
        this.panel.title = `USS: ${normalizedPath}`;

        // Show loading state
        this.panel.webview.html = this.getLoadingHtml(normalizedPath);

        try {
            const listing = await this.client.listUssDirectory(
                this.server,
                this.credentials,
                normalizedPath
            );
            // Use the resolved path (after PATHPREFIX stripping) for navigation
            const actualPath = listing.resolvedPath || normalizedPath;
            this.currentPath = actualPath;
            this.panel.title = `USS: ${actualPath}`;
            this.panel.webview.html = this.getDirectoryHtml(actualPath, listing.items);
        } catch (error) {
            const msg = error instanceof Error ? error.message : String(error);
            this.panel.webview.html = this.getErrorHtml(normalizedPath, msg);
        }
    }

    private async openFile(filePath: string): Promise<void> {
        if (!UssPanel.contentProvider) {
            vscode.window.showErrorMessage('USS file content provider not registered');
            return;
        }

        try {
            const content = await vscode.window.withProgress(
                {
                    location: vscode.ProgressLocation.Notification,
                    title: `Reading ${filePath}...`,
                    cancellable: false
                },
                async () => {
                    return await this.client.readUssFile(
                        this.server,
                        this.credentials,
                        filePath
                    );
                }
            );

            // Build a read-only URI: smpe-uss:/server/path/to/file.ext
            const fileName = filePath.split('/').pop() || 'file';
            const uri = vscode.Uri.from({
                scheme: UssFileContentProvider.scheme,
                authority: this.server.name,
                path: filePath,
                query: `lang=${this.detectLanguage(fileName)}`
            });

            UssPanel.contentProvider.setContent(uri, content);

            const doc = await vscode.workspace.openTextDocument(uri);
            await vscode.languages.setTextDocumentLanguage(doc, this.detectLanguage(fileName));
            await vscode.window.showTextDocument(doc, { preview: true, preserveFocus: false });
        } catch (error) {
            const msg = error instanceof Error ? error.message : String(error);
            vscode.window.showErrorMessage(`Failed to read ${filePath}: ${msg}`);
        }
    }

    private detectLanguage(filePath: string): string {
        const ext = filePath.split('.').pop()?.toLowerCase() || '';
        const langMap: Record<string, string> = {
            'sh': 'shellscript', 'bash': 'shellscript', 'ksh': 'shellscript',
            'xml': 'xml', 'json': 'json', 'yaml': 'yaml', 'yml': 'yaml',
            'py': 'python', 'pl': 'perl', 'c': 'c', 'h': 'c', 'cpp': 'cpp',
            'java': 'java', 'js': 'javascript', 'properties': 'properties',
            'conf': 'ini', 'cfg': 'ini', 'txt': 'plaintext'
        };
        return langMap[ext] || 'plaintext';
    }

    private getNonce(): string {
        let text = '';
        const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        for (let i = 0; i < 32; i++) {
            text += possible.charAt(Math.floor(Math.random() * possible.length));
        }
        return text;
    }

    private escapeHtml(text: string): string {
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }

    private formatSize(bytes: number): string {
        if (bytes < 1024) { return `${bytes} B`; }
        if (bytes < 1024 * 1024) { return `${(bytes / 1024).toFixed(1)} KB`; }
        return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    }

    private formatMode(mode: string): string {
        // z/OSMF returns mode as octal string like "drwxr-xr-x" or similar
        return mode;
    }

    private isDirectory(entry: UssEntry): boolean {
        return entry.mode.startsWith('d');
    }

    private getLoadingHtml(path: string): string {
        const nonce = this.getNonce();
        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'nonce-${nonce}';">
    <style nonce="${nonce}">
        body { font-family: var(--vscode-editor-font-family, monospace); color: var(--vscode-foreground); background: var(--vscode-editor-background); padding: 16px; }
        .loading { text-align: center; padding: 32px; color: var(--vscode-descriptionForeground); }
    </style>
</head>
<body><div class="loading">Loading ${this.escapeHtml(path)}...</div></body>
</html>`;
    }

    private getErrorHtml(path: string, error: string): string {
        const nonce = this.getNonce();
        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'nonce-${nonce}';">
    <style nonce="${nonce}">
        body { font-family: var(--vscode-editor-font-family, monospace); color: var(--vscode-foreground); background: var(--vscode-editor-background); padding: 16px; }
        .error { padding: 16px; background: var(--vscode-inputValidation-errorBackground); border: 1px solid var(--vscode-inputValidation-errorBorder); border-radius: 4px; }
        h2 { margin: 0 0 8px; }
    </style>
</head>
<body>
    <h2>USS: ${this.escapeHtml(path)}</h2>
    <div class="error">${this.escapeHtml(error)}</div>
</body>
</html>`;
    }

    private getDirectoryHtml(path: string, entries: UssEntry[]): string {
        const nonce = this.getNonce();

        // Sort: directories first, then alphabetically
        const sorted = [...entries].sort((a, b) => {
            const aDir = this.isDirectory(a);
            const bDir = this.isDirectory(b);
            if (aDir !== bDir) { return aDir ? -1 : 1; }
            return a.name.localeCompare(b.name);
        });

        // Filter out . and ..
        const filtered = sorted.filter(e => e.name !== '.' && e.name !== '..');

        // Build parent path for ".." navigation
        const parentPath = path.substring(0, path.lastIndexOf('/')) || '/';
        const showParent = path !== '/';

        const rows = filtered.map(entry => {
            const isDir = this.isDirectory(entry);
            const icon = isDir ? '&#128193;' : '&#128196;';
            const fullPath = `${path}/${entry.name}`;
            const linkClass = isDir ? 'dir-link' : 'file-link';
            const command = isDir ? 'navigate' : 'openFile';

            return `<tr>
                <td><span class="${linkClass}" data-command="${command}" data-path="${this.escapeHtml(fullPath)}">${icon} ${this.escapeHtml(entry.name)}</span></td>
                <td>${this.escapeHtml(this.formatMode(entry.mode))}</td>
                <td class="size">${isDir ? '-' : this.formatSize(entry.size)}</td>
                <td>${this.escapeHtml(entry.user)}</td>
                <td>${this.escapeHtml(entry.group)}</td>
                <td>${this.escapeHtml(entry.mtime)}</td>
            </tr>`;
        }).join('');

        const parentRow = showParent
            ? `<tr><td><span class="dir-link" data-command="navigate" data-path="${this.escapeHtml(parentPath)}">&#128193; ..</span></td><td></td><td></td><td></td><td></td><td></td></tr>`
            : '';

        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'nonce-${nonce}'; script-src 'nonce-${nonce}';">
    <style nonce="${nonce}">
        body {
            font-family: var(--vscode-editor-font-family, monospace);
            font-size: var(--vscode-font-size);
            color: var(--vscode-foreground);
            background: var(--vscode-editor-background);
            padding: 16px;
            margin: 0;
        }
        .header {
            margin-bottom: 16px;
            padding-bottom: 8px;
            border-bottom: 1px solid var(--vscode-panel-border);
        }
        .header h2 { margin: 0; }
        .header-info {
            font-size: 0.9em;
            color: var(--vscode-descriptionForeground);
            margin-top: 4px;
        }
        .breadcrumb {
            font-size: 0.9em;
            color: var(--vscode-descriptionForeground);
            margin-top: 4px;
        }
        .breadcrumb span {
            cursor: pointer;
            color: var(--vscode-textLink-foreground);
        }
        .breadcrumb span:hover {
            text-decoration: underline;
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
        }
        th {
            background: var(--vscode-keybindingTable-headerBackground, rgba(128,128,128,0.15));
            font-weight: 600;
            position: sticky;
            top: 0;
        }
        tbody tr:nth-child(odd) {
            background: var(--vscode-keybindingTable-rowsBackground, rgba(128,128,128,0.04));
        }
        tbody tr:hover {
            background: var(--vscode-list-hoverBackground);
        }
        .size { text-align: right; }
        .dir-link, .file-link {
            cursor: pointer;
            color: var(--vscode-textLink-foreground);
        }
        .dir-link:hover, .file-link:hover {
            text-decoration: underline;
            color: var(--vscode-textLink-activeForeground);
        }
        .dir-link { font-weight: 600; }
        .count-badge {
            background: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            padding: 2px 6px;
            border-radius: 10px;
            font-size: 0.8em;
            margin-left: 8px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h2>USS: ${this.escapeHtml(path)}<span class="count-badge">${filtered.length} entries</span></h2>
        <div class="header-info">Server: ${this.escapeHtml(this.server.name)}</div>
        <div class="breadcrumb" id="breadcrumb"></div>
    </div>
    <table>
        <thead>
            <tr>
                <th>Name</th>
                <th>Mode</th>
                <th class="size">Size</th>
                <th>User</th>
                <th>Group</th>
                <th>Modified</th>
            </tr>
        </thead>
        <tbody>${parentRow}${rows}</tbody>
    </table>
    <script nonce="${nonce}">
        const vscode = acquireVsCodeApi();
        const currentPath = ${JSON.stringify(path)};

        // Build breadcrumb navigation
        const bc = document.getElementById('breadcrumb');
        const parts = currentPath.split('/').filter(p => p.length > 0);
        let bcHtml = '<span data-path="/">/</span>';
        let accumulated = '';
        for (const part of parts) {
            accumulated += '/' + part;
            bcHtml += ' <span data-path="' + accumulated + '">' + part + '</span> /';
        }
        bc.innerHTML = bcHtml;

        // Handle all clicks via event delegation
        document.addEventListener('click', (e) => {
            // Breadcrumb navigation
            const bcSpan = e.target.closest('.breadcrumb span');
            if (bcSpan && bcSpan.dataset.path) {
                vscode.postMessage({ command: 'navigate', path: bcSpan.dataset.path });
                return;
            }
            // Directory and file links
            const link = e.target.closest('.dir-link, .file-link');
            if (link) {
                vscode.postMessage({ command: link.dataset.command, path: link.dataset.path });
            }
        });
    </script>
</body>
</html>`;
    }

    public dispose(): void {
        UssPanel.currentPanel = undefined;
        this.panel.dispose();
        while (this.disposables.length) {
            const disposable = this.disposables.pop();
            if (disposable) {
                disposable.dispose();
            }
        }
    }
}
