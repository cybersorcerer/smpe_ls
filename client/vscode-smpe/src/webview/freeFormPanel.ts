/**
 * Free Form Query Webview Panel
 * Combined input form and result table for z/OSMF CSI queries
 */

import * as vscode from 'vscode';
import { QueryProvider } from '../zosmf/queryProvider';
import { ZosmfServer, Credentials } from '../zosmf/types';
import { ZosmfClient } from '../zosmf/client';
import { UssPanel } from './ussPanel';
import { DatasetPanel } from './datasetPanel';

export class FreeFormPanel {
    public static currentPanel: FreeFormPanel | undefined;
    private readonly panel: vscode.WebviewPanel;
    private readonly queryProvider: QueryProvider;
    private readonly outputChannel: vscode.OutputChannel;
    private disposables: vscode.Disposable[] = [];
    private lastClient: ZosmfClient | undefined;
    private lastServer: ZosmfServer | undefined;
    private lastCredentials: Credentials | undefined;

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

    private debugLog(message: string): void {
        const debug = vscode.workspace.getConfiguration('smpe').get<boolean>('debug', true);
        if (!debug) { return; }
        this.log(message);
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
                csiList: Array.isArray(s.csi) ? s.csi : [s.csi],
                defaultCsi: s.defaultCsi || '',
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
            case 'openUssPath':
                await this.openUssPath(message.path);
                break;
            case 'openDataset':
                await this.openDataset(message.dataset);
                break;
        }
    }

    private async executeQuery(message: any): Promise<void> {
        const { serverName, selectedCsi, zones, entryType, subentries, filter } = message;

        this.log(`Free form query: server=${serverName}, csi=${selectedCsi}, zones=${zones}, entry=${entryType}, subentries=${subentries}, filter=${filter}`);

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

        const resolvedServer = { ...server, csi: selectedCsi || (Array.isArray(server.csi) ? server.csi[0] : server.csi) };

        const credentials = await configManager.getCredentials(resolvedServer);
        if (!credentials) {
            this.panel.webview.postMessage({ command: 'error', message: 'Authentication cancelled' });
            return;
        }

        // Parse zone input and resolve patterns
        const zoneList = (zones as string).split(',').map((z: string) => z.trim().toUpperCase()).filter((z: string) => z.length > 0);
        const resolvedZones = this.queryProvider.resolveZonePatterns(resolvedServer, zoneList);

        if (resolvedZones.length === 0) {
            this.panel.webview.postMessage({ command: 'error', message: 'No zones resolved from input' });
            return;
        }

        // Parse subentries
        const subentryList = (subentries as string).split(',').map((s: string) => s.trim().toUpperCase()).filter((s: string) => s.length > 0);

        this.log(`Free Form Query - EntryType: ${entryType.toUpperCase()}, Zones: ${resolvedZones.join(',')}, Subentries: [${subentryList.join(',')}], Filter: "${filter || ''}"`);

        this.panel.webview.postMessage({ command: 'progress', message: 'Sending query to z/OSMF...' });

        // Store context for USS/Dataset browsing
        this.lastClient = this.queryProvider.getClient();
        this.lastServer = resolvedServer;
        this.lastCredentials = credentials;

        try {
            const client = this.lastClient;
            const result = await client.queryFreeForm(
                resolvedServer,
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
            // Log raw subentry data for debugging
            for (const entry of (result.entries || []).slice(0, 3)) {
                this.debugLog(`Raw subentries for ${entry.entryname}: ${JSON.stringify(entry.subentries)}`);
            }
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

    private async openUssPath(ussPath: string): Promise<void> {
        if (!this.lastClient || !this.lastServer || !this.lastCredentials) {
            vscode.window.showWarningMessage('No z/OSMF connection available. Please execute a query first.');
            return;
        }
        await UssPanel.open(this.lastClient, this.lastServer, this.lastCredentials, ussPath);
    }

    private async openDataset(datasetName: string): Promise<void> {
        if (!this.lastClient || !this.lastServer || !this.lastCredentials) {
            vscode.window.showWarningMessage('No z/OSMF connection available. Please execute a query first.');
            return;
        }
        await DatasetPanel.open(this.lastClient, this.lastServer, this.lastCredentials, datasetName);
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
        #tableContainer {
            overflow-x: auto;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            font-size: var(--vscode-font-size);
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
        .subentry-picker {
            display: none;
            margin-bottom: 8px;
            padding: 8px;
            border: 1px solid var(--vscode-panel-border);
            border-radius: 2px;
            background-color: var(--vscode-editor-background);
        }
        .subentry-picker.visible {
            display: block;
        }
        .subentry-grid {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 2px 12px;
        }
        .subentry-grid label {
            display: flex;
            align-items: center;
            gap: 4px;
            min-width: unset;
            font-weight: normal;
            font-size: 0.9em;
            cursor: pointer;
            padding: 2px 0;
        }
        .subentry-grid input[type="checkbox"] {
            flex: none;
            margin: 0;
        }
        .subentry-picker-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 6px;
        }
        .subentry-picker-header span {
            font-weight: 600;
            font-size: 0.9em;
        }
        .toggle-link {
            background: none;
            border: none;
            color: var(--vscode-textLink-foreground);
            cursor: pointer;
            font-size: var(--vscode-font-size);
            padding: 0;
            text-decoration: underline;
        }
        .toggle-link:hover {
            color: var(--vscode-textLink-activeForeground);
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
        /* Entry Type Picker */
        .entrytype-picker {
            display: none;
            margin-bottom: 8px;
            padding: 8px;
            border: 1px solid var(--vscode-panel-border);
            border-radius: 2px;
            background-color: var(--vscode-editor-background);
        }
        .entrytype-picker.visible {
            display: block;
        }
        .entrytype-picker-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 6px;
        }
        .entrytype-picker-header span {
            font-weight: 600;
            font-size: 0.9em;
        }
        .entrytype-radio-grid {
            display: grid;
            grid-template-columns: repeat(6, 1fr);
            gap: 2px 12px;
        }
        .entrytype-radio-grid label {
            display: flex;
            align-items: center;
            gap: 4px;
            min-width: unset;
            font-weight: normal;
            font-size: 0.9em;
            cursor: pointer;
            padding: 2px 0;
        }
        .entrytype-radio-grid input[type="radio"] {
            flex: none;
            margin: 0;
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
            <label for="csi">CSI</label>
            <select id="csi">
                <option value="">Loading...</option>
            </select>
        </div>
        <div class="form-row">
            <label for="zones">Zones</label>
            <input type="text" id="zones" placeholder="GLOBAL, MVS* (comma-separated, * and ? wildcards)" />
        </div>
        <div class="form-row">
            <label for="entryType">Entry Type</label>
            <input type="text" id="entryType" value="SYSMOD" placeholder="e.g. SYSMOD, MOD, UNIX1" autocomplete="off" />
            <button id="toggleEntryTypePickerBtn" title="Pick entry type">Pick...</button>
        </div>
        <div id="entryTypePicker" class="entrytype-picker">
            <div class="entrytype-picker-header">
                <span>Select Entry Type</span>
            </div>
            <div id="entryTypePickerGroups"></div>
        </div>
        <div class="form-row">
            <label for="subentries">Subentries</label>
            <input type="text" id="subentries" placeholder="FMID,ERROR,RECDATE (comma-separated)" />
            <button id="togglePickerBtn" title="Pick subentries">Pick...</button>
        </div>
        <div id="subentryPicker" class="subentry-picker">
            <div class="subentry-picker-header">
                <span id="pickerTitle">Available Subentries</span>
                <div style="display: flex; gap: 8px;">
                    <button id="addSubentriesBtn">Add Selected</button>
                </div>
            </div>
            <div id="subentryGrid" class="subentry-grid"></div>
        </div>
        <div class="form-row">
            <label for="filter">Filter</label>
            <input type="text" id="filter" placeholder="ENAME='UA12345' (optional)" />
        </div>
        <div class="button-row">
            <button id="executeBtn">Execute Query</button>
        </div>
    </div>

    <div id="cellTooltip" class="cell-tooltip"></div>
    <div id="statusBar" class="status-bar" style="display:none;"></div>

    <div id="resultSection" class="result-section" style="display:none;">
        <div class="result-header">
            <div>
                <h3>Results<span id="countBadge" class="count-badge">0</span></h3>
            </div>
            <div class="toolbar">
                <button id="exportJsonBtn">Export JSON</button>
                <button id="exportCsvBtn">Export CSV</button>
            </div>
        </div>
        <div id="tableContainer"></div>
    </div>

    <script nonce="${nonce}">
        const vscode = acquireVsCodeApi();
        let currentResult = null;
        let currentSubentries = [];
        let currentEntryType = '';

        // Valid subentries per entry type (IBM z/OS 3.1 SMP/E Reference)
        const SUBENTRIES_BY_TYPE = {
            SYSMOD: ['ACCEPT','ACCID','APPID','APPLY','ASSEM','BYPASS','CIFREQ','DELBY','DELETE','DELLMOD','DESCRIPTION','DLMOD','ELEMENT','ELEMMOV','EMOVE','ENAME','ERROR','FEATURE','FESN','FMID','HOLDDATA','IFREQ','INSTALLDATE','INSTALLTIME','JAR','JARUPD','JCLIN','LASTSUP','LASTUPD','LASTUPDTYPE','MAC','MACUPD','MOD','NPRE','PRE','PROGRAM','RECDATE','RECTIME','REGEN','RENLMOD','REQ','RESDATE','RESTIME','RESTORE','REWORK','RLMOD','SMODTYPE','SOURCEID','SRC','SRCUPD','SREL','SUPBY','SUPING','SZAP','TLIBPREFIX','UCLDATE','UCLTIME','VERSION','XZAP'],
            DDDEF: ['CONCAT','DATACLAS','DATASET','DIR','DISP','DSNTYPE','DSPREFIX','ENAME','INITDISP','MGMTCLAS','PATH','PROTECT','SPACE','STORCLAS','SYSOUT','UNIT','UNITS','VOLUME','WAITFORDSN'],
            TARGETZONE: ['ENAME','OPTIONS','RELATED','SREL','TIEDTO','UPGLEVEL','XZLINK','ZDESC'],
            DLIB: ['ENAME','LASTUPD','LASTUPDTYPE','SYSLIB'],
            GLOBALZONE: ['ENAME','FMID','OPTIONS','SREL','UPGLEVEL','ZDESC','ZONEINDEX'],
            ASSEM: ['ASMIN','ENAME','LASTUPD','LASTUPDTYPE'],
            DATA: ['ALIAS','DISTLIB','ENAME','FMID','LASTUPD','LASTUPDTYPE','RMID','SYSLIB'],
            LMOD: ['CALLLIBS','COPIED','ENAME','LASTUPD','LASTUPDTYPE','LEPARM','LECNTL','LMODALIAS','LMODSYMLINK','MODDEL','RC','SIDEDECKLIB','SYSLIB','UTIN','XZMOD','XZMODP'],
            MAC: ['DISTLIB','ENAME','FMID','GENASM','LASTUPD','LASTUPDTYPE','MALIAS','RMID','SYSLIB','UMID'],
            MOD: ['ASSEMBLE','CSECT','DALIAS','DISTLIB','ENAME','FMID','LASTUPD','LASTUPDTYPE','LEPARM','LMOD','RMID','RMIDASM','TALIAS','UMID','XZLMOD','XZLMODP'],
            SRC: ['DISTLIB','ENAME','FMID','LASTUPD','LASTUPDTYPE','RMID','SYSLIB','UMID'],
            // Additional element types
            JAR: ['DISTLIB','ENAME','FMID','JARPARM','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB','TXLIB','UMID','VERSION'],
            JARUPD: ['ENAME','JARPARM','LASTUPD','LASTUPDTYPE','LINK','SYMLINK','SYMPATH','TXLIB'],
            PROGRAM: ['ALIAS','DISTLIB','ENAME','FMID','LASTUPD','LASTUPDTYPE','LKLIB','RMID','SYSLIB','VERSION'],
            DLIBZONE: ['ACCJCLIN','ENAME','OPTIONS','RELATED','SREL','UPGLEVEL','ZDESC'],
            FEATURE: ['DESCRIPTION','ENAME','FMID','PRODUCT','RECDATE','RECTIME','REWORK','UCLDATE','UCLTIME'],
            FMIDSET: ['ENAME','FMID'],
            HOLDDATA: ['ENAME','HOLDCLASS','HOLDDATA','HOLDDATE','HOLDFIXCAT','HOLDFMID','HOLDREASON','HOLDRESOLVER','HOLDTYPE'],
            OPTIONS: ['AMS','ASM','CHANGEFILE','COMP','COMPACT','COPY','DSPREFIX','DSSPACE','ENAME','EXRTYDD','FIXCAT','HFSCOPY','IOSUP','LKED','MSGFILTER','MSGWIDTH','NOPURGE','NOREJECT','ORDERRET','PAGELEN','PEMAX','RECZGRP','RECEXCGRP','RETRY','RETRYDDN','SAVEMTS','SAVESTS','SUPPHOLD','UPDATE','ZAP'],
            ORDER: ['APARS','CONTENT','DOWNLDATE','DOWNLTIME','ENAME','ORDERDATE','ORDERID','ORDERSERVER','ORDERTIME','PKGID','PTFS','STATUS','USERID','ZONES'],
            PRODUCT: ['DESCRIPTION','PRODID','PRODSUP','RECDATE','RECTIME','REWORK','SREL','UCLDATE','UCLTIME','URL','VENDOR','VRM'],
            UTILITY: ['ENAME','LIST','NAME','PRINT','RC','UTILPARM'],
            ZONESET: ['ENAME','XZREQCHK','ZONENAME'],
            // HFS (Hierarchical File System) element types share the same subentries
            AIX1: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            AIX2: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            AIX3: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            AIX4: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            AIX5: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            CLIENT1: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            CLIENT2: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            CLIENT3: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            CLIENT4: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            CLIENT5: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            OS21: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            OS22: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            OS23: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            OS24: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            OS25: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            UNIX1: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            UNIX2: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            UNIX3: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            UNIX4: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            UNIX5: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            WIN1: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            WIN2: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            WIN3: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            WIN4: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB'],
            WIN5: ['DISTLIB','ENAME','FMID','HFSPARM','INSTMODE','LASTUPD','LASTUPDTYPE','LINK','RMID','SHSCRIPT','SYMLINK','SYMPATH','SYSLIB']
        };

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
            const csiSelect = document.getElementById('csi');
            select.innerHTML = '';
            for (const s of servers) {
                const opt = document.createElement('option');
                opt.value = s.name;
                opt.textContent = s.name + ' (' + s.host + ')';
                select.appendChild(opt);
            }

            function updateServerDetails(serverName) {
                const selected = servers.find(s => s.name === serverName);
                if (!selected) return;
                // Update CSI dropdown
                csiSelect.innerHTML = '';
                for (const csi of selected.csiList) {
                    const opt = document.createElement('option');
                    opt.value = csi;
                    opt.textContent = csi;
                    if (csi === selected.defaultCsi) opt.selected = true;
                    csiSelect.appendChild(opt);
                }
                // Update default zones
                if (selected.defaultZones.length > 0) {
                    document.getElementById('zones').value = selected.defaultZones.join(', ');
                }
            }

            // Initialize with first server
            if (servers.length > 0) {
                updateServerDetails(servers[0].name);
            }

            // Update when server changes
            select.addEventListener('change', () => {
                updateServerDetails(select.value);
            });
        }

        function executeQuery() {
            const serverName = document.getElementById('server').value;
            const selectedCsi = document.getElementById('csi').value;
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
            document.getElementById('subentryPicker').classList.remove('visible');
            showProgress('Connecting...');

            vscode.postMessage({
                command: 'executeQuery',
                serverName: serverName,
                selectedCsi: selectedCsi,
                zones: zones,
                entryType: entryType,
                subentries: subentries,
                filter: filter
            });

            // Store entry type for link rendering in results
            currentEntryType = entryType.trim().toUpperCase();
        }

        function showResult(result, subentries) {
            currentResult = result;
            currentSubentries = subentries;
            document.getElementById('executeBtn').disabled = false;

            const entries = (result.entries || []).slice().sort((a, b) =>
                (a.zonename || '').localeCompare(b.zonename || '') ||
                (a.entryname || '').localeCompare(b.entryname || '')
            );
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
                    const val = subData[sub] || '';
                    if (currentEntryType === 'DDDEF' && sub === 'PATH' && val.startsWith('/')) {
                        html += '<td><a href="#" class="uss-link" data-path="' + escapeHtml(val) + '">' + escapeHtml(val) + '</a></td>';
                    } else if (currentEntryType === 'DDDEF' && sub === 'DATASET' && val.length > 0) {
                        html += '<td><a href="#" class="ds-link" data-dataset="' + escapeHtml(val) + '">' + escapeHtml(val) + '</a></td>';
                    } else {
                        html += '<td>' + escapeHtml(val) + '</td>';
                    }
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
                            const flat = value.map(item => {
                                if (typeof item === 'object' && item !== null) {
                                    return Object.values(item).flat().join(',');
                                }
                                return String(item);
                            });
                            data[key] = flat.join(', ');
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

        // Subentry picker: build checkbox grid for current entry type
        function updateSubentryGrid() {
            const entryType = document.getElementById('entryType').value;
            const grid = document.getElementById('subentryGrid');
            const subs = (SUBENTRIES_BY_TYPE[entryType] || []).slice().sort();
            grid.innerHTML = '';
            for (const sub of subs) {
                const lbl = document.createElement('label');
                const cb = document.createElement('input');
                cb.type = 'checkbox';
                cb.value = sub;
                lbl.appendChild(cb);
                lbl.appendChild(document.createTextNode(sub));
                grid.appendChild(lbl);
            }
        }

        // Toggle picker visibility
        document.getElementById('togglePickerBtn').addEventListener('click', () => {
            const picker = document.getElementById('subentryPicker');
            picker.classList.toggle('visible');
        });

        // Add checked subentries to text field
        document.getElementById('addSubentriesBtn').addEventListener('click', () => {
            const grid = document.getElementById('subentryGrid');
            const input = document.getElementById('subentries');
            const checked = Array.from(grid.querySelectorAll('input[type="checkbox"]:checked'));
            if (checked.length === 0) { return; }

            const selected = checked.map(cb => cb.value);
            const current = input.value.trim();
            const existing = current ? current.split(',').map(s => s.trim().toUpperCase()) : [];
            const toAdd = selected.filter(s => !existing.includes(s));
            if (toAdd.length === 0) { return; }

            if (current) {
                input.value = current + ',' + toAdd.join(',');
            } else {
                input.value = toAdd.join(',');
            }

            // Uncheck all and close picker
            checked.forEach(cb => { cb.checked = false; });
            document.getElementById('subentryPicker').classList.remove('visible');
        });

        // ── Entry Type Picker ───────────────────────────────────────────────
        const ALL_ENTRY_TYPES = Object.keys(SUBENTRIES_BY_TYPE).sort();

        // Build the radio-button grid once
        (function buildEntryTypePicker() {
            const grid = document.createElement('div');
            grid.className = 'entrytype-radio-grid';

            for (const item of ALL_ENTRY_TYPES) {
                const lbl = document.createElement('label');
                const rb = document.createElement('input');
                rb.type = 'radio';
                rb.name = 'entryTypePick';
                rb.value = item;
                rb.addEventListener('change', () => {
                    document.getElementById('entryType').value = item;
                    document.getElementById('entryTypePicker').classList.remove('visible');
                    updateSubentryGrid();
                    document.getElementById('subentries').value = '';
                    document.getElementById('filter').value = '';
                    document.getElementById('subentryPicker').classList.remove('visible');
                });
                lbl.appendChild(rb);
                lbl.appendChild(document.createTextNode(item));
                grid.appendChild(lbl);
            }

            document.getElementById('entryTypePickerGroups').appendChild(grid);
        })();

        document.getElementById('toggleEntryTypePickerBtn').addEventListener('click', () => {
            const picker = document.getElementById('entryTypePicker');
            const isVisible = picker.classList.toggle('visible');
            if (isVisible) {
                // Pre-select the current value in the radio grid
                const current = document.getElementById('entryType').value.trim().toUpperCase();
                const radios = picker.querySelectorAll('input[type="radio"]');
                radios.forEach(r => { r.checked = r.value === current; });
            }
        });

        document.getElementById('entryType').addEventListener('input', () => {
            updateSubentryGrid();
            document.getElementById('subentries').value = '';
            document.getElementById('filter').value = '';
            document.getElementById('subentryPicker').classList.remove('visible');
        });

        // Initial population
        updateSubentryGrid();

        // Attach event listeners (onclick attributes are blocked by CSP)
        document.getElementById('executeBtn').addEventListener('click', executeQuery);
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

        // Custom tooltip for truncated cells (title attr doesn't work in VSCode webviews)
        const tooltip = document.getElementById('cellTooltip');
        let tooltipTimeout = null;
        document.addEventListener('mouseover', (e) => {
            const td = e.target.closest('td');
            if (!td) { return; }
            const text = td.textContent || '';
            // Show tooltip if content is truncated OR text is longer than visible width
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
