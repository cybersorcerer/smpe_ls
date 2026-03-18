/**
 * Webview Panel for MVS Dataset Member Listing via z/OSMF Dataset REST API
 */

import * as vscode from 'vscode';
import { ZosmfClient } from '../zosmf/client';
import { ZosmfServer, Credentials, DatasetMember } from '../zosmf/types';

/**
 * Content provider for read-only dataset/member viewing
 */
export class DatasetContentProvider implements vscode.TextDocumentContentProvider {
    public static readonly scheme = 'smpe-ds';
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

export class DatasetPanel {
    public static currentPanel: DatasetPanel | undefined;
    public static contentProvider: DatasetContentProvider | undefined;
    private readonly panel: vscode.WebviewPanel;
    private disposables: vscode.Disposable[] = [];
    private client: ZosmfClient;
    private server: ZosmfServer;
    private credentials: Credentials;
    private currentDataset: string = '';

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
                    case 'openMember':
                        await this.openMember(message.dataset, message.member);
                        break;
                }
            },
            null,
            this.disposables
        );
    }

    /**
     * Open a dataset. If it's a PDS, show member listing.
     * If it's sequential (member list fails), open content directly.
     */
    public static async open(
        client: ZosmfClient,
        server: ZosmfServer,
        credentials: Credentials,
        datasetName: string
    ): Promise<void> {
        const column = vscode.ViewColumn.One;

        // First, try to list members (PDS)
        try {
            const listing = await vscode.window.withProgress(
                {
                    location: vscode.ProgressLocation.Notification,
                    title: `Loading ${datasetName}...`,
                    cancellable: false
                },
                async () => {
                    return await client.listDatasetMembers(server, credentials, datasetName);
                }
            );

            // PDS — show member listing panel
            if (DatasetPanel.currentPanel) {
                try {
                    DatasetPanel.currentPanel.client = client;
                    DatasetPanel.currentPanel.server = server;
                    DatasetPanel.currentPanel.credentials = credentials;
                    DatasetPanel.currentPanel.panel.reveal(column);
                    DatasetPanel.currentPanel.showMembers(datasetName, listing.items);
                    return;
                } catch {
                    DatasetPanel.currentPanel = undefined;
                }
            }

            const panel = vscode.window.createWebviewPanel(
                'smpeDatasetBrowser',
                `DS: ${datasetName}`,
                column,
                {
                    enableScripts: true,
                    retainContextWhenHidden: true
                }
            );

            DatasetPanel.currentPanel = new DatasetPanel(panel, client, server, credentials);
            DatasetPanel.currentPanel.showMembers(datasetName, listing.items);
        } catch (error) {
            // Not a PDS or error — try reading as sequential dataset
            const msg = error instanceof Error ? error.message : String(error);

            // If HTTP 500 or member list not applicable, try sequential read
            try {
                await DatasetPanel.openSequentialDataset(client, server, credentials, datasetName);
            } catch (seqError) {
                const seqMsg = seqError instanceof Error ? seqError.message : String(seqError);
                vscode.window.showErrorMessage(
                    `Failed to open ${datasetName}: Member list: ${msg} | Sequential read: ${seqMsg}`
                );
            }
        }
    }

    /**
     * Open a sequential dataset as read-only document
     */
    private static async openSequentialDataset(
        client: ZosmfClient,
        server: ZosmfServer,
        credentials: Credentials,
        datasetName: string
    ): Promise<void> {
        if (!DatasetPanel.contentProvider) {
            vscode.window.showErrorMessage('Dataset content provider not registered');
            return;
        }

        const content = await vscode.window.withProgress(
            {
                location: vscode.ProgressLocation.Notification,
                title: `Reading ${datasetName}...`,
                cancellable: false
            },
            async () => {
                return await client.readDataset(server, credentials, datasetName);
            }
        );

        const uri = vscode.Uri.from({
            scheme: DatasetContentProvider.scheme,
            authority: server.name,
            path: `/${datasetName}`
        });

        DatasetPanel.contentProvider.setContent(uri, content);

        const doc = await vscode.workspace.openTextDocument(uri);
        await vscode.window.showTextDocument(doc, { preview: true, preserveFocus: false });
    }

    private showMembers(datasetName: string, members: DatasetMember[]): void {
        this.currentDataset = datasetName;
        this.panel.title = `DS: ${datasetName}`;
        this.panel.webview.html = this.getMemberListHtml(datasetName, members);
    }

    private async openMember(datasetName: string, memberName: string): Promise<void> {
        if (!DatasetPanel.contentProvider) {
            vscode.window.showErrorMessage('Dataset content provider not registered');
            return;
        }

        try {
            const content = await vscode.window.withProgress(
                {
                    location: vscode.ProgressLocation.Notification,
                    title: `Reading ${datasetName}(${memberName})...`,
                    cancellable: false
                },
                async () => {
                    return await this.client.readDataset(
                        this.server,
                        this.credentials,
                        datasetName,
                        memberName
                    );
                }
            );

            const uri = vscode.Uri.from({
                scheme: DatasetContentProvider.scheme,
                authority: this.server.name,
                path: `/${datasetName}/${memberName}`
            });

            DatasetPanel.contentProvider.setContent(uri, content);

            const doc = await vscode.workspace.openTextDocument(uri);
            await vscode.window.showTextDocument(doc, { preview: true, preserveFocus: false });
        } catch (error) {
            const msg = error instanceof Error ? error.message : String(error);
            vscode.window.showErrorMessage(`Failed to read ${datasetName}(${memberName}): ${msg}`);
        }
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

    private getMemberListHtml(datasetName: string, members: DatasetMember[]): string {
        const nonce = this.getNonce();

        const sorted = [...members].sort((a, b) => a.member.localeCompare(b.member));

        const rows = sorted.map(m => {
            return `<tr>
                <td><span class="member-link" data-dataset="${this.escapeHtml(datasetName)}" data-member="${this.escapeHtml(m.member)}">&#128196; ${this.escapeHtml(m.member)}</span></td>
                <td>${m.user ? this.escapeHtml(m.user) : ''}</td>
                <td>${m.c4date ? this.escapeHtml(m.c4date) : ''}</td>
                <td>${m.m4date ? this.escapeHtml(m.m4date) : ''}</td>
                <td>${m.vers !== undefined ? m.vers : ''}</td>
                <td>${m.mod !== undefined ? m.mod : ''}</td>
            </tr>`;
        }).join('');

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
        .member-link {
            cursor: pointer;
            color: var(--vscode-textLink-foreground);
            font-weight: 600;
        }
        .member-link:hover {
            text-decoration: underline;
            color: var(--vscode-textLink-activeForeground);
        }
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
        <h2>${this.escapeHtml(datasetName)}<span class="count-badge">${sorted.length} members</span></h2>
        <div class="header-info">Server: ${this.escapeHtml(this.server.name)}</div>
    </div>
    <table>
        <thead>
            <tr>
                <th>Member</th>
                <th>User</th>
                <th>Created</th>
                <th>Modified</th>
                <th>Ver</th>
                <th>Mod</th>
            </tr>
        </thead>
        <tbody>${rows}</tbody>
    </table>
    <script nonce="${nonce}">
        const vscode = acquireVsCodeApi();
        document.addEventListener('click', (e) => {
            const link = e.target.closest('.member-link');
            if (link) {
                vscode.postMessage({
                    command: 'openMember',
                    dataset: link.dataset.dataset,
                    member: link.dataset.member
                });
            }
        });
    </script>
</body>
</html>`;
    }

    public dispose(): void {
        DatasetPanel.currentPanel = undefined;
        this.panel.dispose();
        while (this.disposables.length) {
            const disposable = this.disposables.pop();
            if (disposable) {
                disposable.dispose();
            }
        }
    }
}
