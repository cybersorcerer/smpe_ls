/**
 * z/OSMF Query Provider
 * Orchestrates query commands with user interaction
 */

import * as vscode from 'vscode';
import { ConfigManager } from './configManager';
import { ZosmfClient } from './client';
import { ZosmfServer, QueryResult, DisplayResult, QueryType } from './types';

export class QueryProvider {
    private configManager: ConfigManager;
    private client: ZosmfClient;
    private outputChannel: vscode.OutputChannel;
    private onResultCallback?: (result: DisplayResult) => void;

    constructor(
        context: vscode.ExtensionContext,
        outputChannel: vscode.OutputChannel
    ) {
        this.configManager = new ConfigManager(context, outputChannel);
        this.client = new ZosmfClient(outputChannel);
        this.outputChannel = outputChannel;
    }

    private log(message: string): void {
        const timestamp = new Date().toISOString();
        this.outputChannel.appendLine(`[${timestamp}] [QueryProvider] ${message}`);
    }

    /**
     * Set callback for query results
     */
    onResult(callback: (result: DisplayResult) => void): void {
        this.onResultCallback = callback;
    }

    /**
     * Create z/OSMF config file command
     */
    async createConfig(): Promise<void> {
        await this.configManager.createConfigFile();
    }

    /**
     * Clear stored passwords command
     */
    async clearPasswords(): Promise<void> {
        const config = this.configManager.loadConfig();
        if (!config) {
            return;
        }

        if (config.servers.length === 1) {
            await this.configManager.deletePassword(config.servers[0]);
            vscode.window.showInformationMessage(`Password cleared for ${config.servers[0].name}`);
            return;
        }

        const items: vscode.QuickPickItem[] = [
            { label: 'All Servers', description: 'Clear passwords for all configured servers' },
            ...config.servers.map(s => ({
                label: s.name,
                description: `${s.user}@${s.host}`
            }))
        ];

        const selected = await vscode.window.showQuickPick(items, {
            placeHolder: 'Select server to clear password'
        });

        if (!selected) {
            return;
        }

        if (selected.label === 'All Servers') {
            await this.configManager.clearAllPasswords();
        } else {
            const server = config.servers.find(s => s.name === selected.label);
            if (server) {
                await this.configManager.deletePassword(server);
                vscode.window.showInformationMessage(`Password cleared for ${server.name}`);
            }
        }
    }

    /**
     * Get server and credentials with user interaction
     */
    private async getServerAndCredentials(): Promise<{ server: ZosmfServer; credentials: { user: string; password: string } } | undefined> {
        const config = this.configManager.loadConfig();
        if (!config) {
            return undefined;
        }

        const server = await this.configManager.selectServer(config);
        if (!server) {
            return undefined;
        }

        const credentials = await this.configManager.getCredentials(server);
        if (!credentials) {
            return undefined;
        }

        return { server, credentials };
    }

    /**
     * Match a zone pattern (with * and ?) against available zones
     */
    private matchZonePattern(pattern: string, zones: string[]): string[] {
        const regexStr = '^' + pattern
            .replace(/[.+^${}()|[\]\\]/g, '\\$&')
            .replace(/\*/g, '.*')
            .replace(/\?/g, '.')
            + '$';
        const regex = new RegExp(regexStr, 'i');
        return zones.filter(z => regex.test(z));
    }

    /**
     * Prompt for zone names (supports * and ? wildcards when zones are defined in config)
     */
    private async promptForZones(server: ZosmfServer): Promise<string[] | undefined> {
        const defaultValue = server.defaultZones?.join(', ') || '';
        const hasZones = server.zones && server.zones.length > 0;

        const input = await vscode.window.showInputBox({
            prompt: hasZones ? 'Enter zone name(s) or pattern (* and ? wildcards)' : 'Enter zone name(s)',
            placeHolder: hasZones ? 'GLOBAL, MVS* (comma-separated, wildcards supported)' : 'GLOBAL, TARGET (comma-separated)',
            value: defaultValue,
            ignoreFocusOut: true,
            validateInput: (value) => {
                if (!value.trim()) {
                    return 'At least one zone name is required';
                }
                return undefined;
            }
        });

        if (!input) {
            return undefined;
        }

        const entries = input.split(',').map(z => z.trim().toUpperCase()).filter(z => z.length > 0);

        if (!hasZones) {
            return entries;
        }

        const resolvedZones: string[] = [];
        for (const entry of entries) {
            if (entry.includes('*') || entry.includes('?')) {
                const matched = this.matchZonePattern(entry, server.zones!);
                if (matched.length === 0) {
                    vscode.window.showWarningMessage(`No zones match pattern '${entry}'`);
                } else {
                    this.log(`Pattern '${entry}' matched zones: ${matched.join(', ')}`);
                    resolvedZones.push(...matched);
                }
            } else {
                resolvedZones.push(entry);
            }
        }

        if (resolvedZones.length === 0) {
            vscode.window.showErrorMessage('No zones resolved from input');
            return undefined;
        }

        return [...new Set(resolvedZones)];
    }

    /**
     * Get the config manager (for external use by FreeFormPanel)
     */
    getConfigManager(): ConfigManager {
        return this.configManager;
    }

    /**
     * Get the client (for external use by FreeFormPanel)
     */
    getClient(): ZosmfClient {
        return this.client;
    }

    /**
     * Resolve zone patterns against server's configured zones
     */
    resolveZonePatterns(server: ZosmfServer, zoneInput: string[]): string[] {
        const hasZones = server.zones && server.zones.length > 0;
        if (!hasZones) {
            return zoneInput;
        }

        const resolvedZones: string[] = [];
        for (const entry of zoneInput) {
            if (entry.includes('*') || entry.includes('?')) {
                const matched = this.matchZonePattern(entry, server.zones!);
                if (matched.length === 0) {
                    this.log(`No zones match pattern '${entry}'`);
                } else {
                    this.log(`Pattern '${entry}' matched zones: ${matched.join(', ')}`);
                    resolvedZones.push(...matched);
                }
            } else {
                resolvedZones.push(entry);
            }
        }

        return [...new Set(resolvedZones)];
    }

    /**
     * Prompt for SYSMOD names
     */
    private async promptForSysmods(): Promise<string[] | undefined> {
        const input = await vscode.window.showInputBox({
            prompt: 'Enter SYSMOD name(s)',
            placeHolder: 'UA12345, UA67890 (comma-separated, or * for all)',
            ignoreFocusOut: true,
            validateInput: (value) => {
                if (!value.trim()) {
                    return 'At least one SYSMOD name is required';
                }
                return undefined;
            }
        });

        if (!input) {
            return undefined;
        }

        return input.split(',').map(s => s.trim().toUpperCase()).filter(s => s.length > 0);
    }

    /**
     * Prompt for DDDEF names
     */
    private async promptForDddefs(): Promise<string[] | undefined> {
        const input = await vscode.window.showInputBox({
            prompt: 'Enter DDDEF name(s)',
            placeHolder: 'SMPLTS, SMPSTS (comma-separated, or * for all)',
            ignoreFocusOut: true,
            validateInput: (value) => {
                if (!value.trim()) {
                    return 'At least one DDDEF name is required';
                }
                return undefined;
            }
        });

        if (!input) {
            return undefined;
        }

        return input.split(',').map(d => d.trim().toUpperCase()).filter(d => d.length > 0);
    }

    /**
     * Execute query with progress
     */
    private async executeWithProgress<T>(
        title: string,
        task: (progress: (msg: string) => void) => Promise<T>
    ): Promise<T | undefined> {
        return vscode.window.withProgress(
            {
                location: vscode.ProgressLocation.Notification,
                title: title,
                cancellable: false
            },
            async (progress) => {
                const updateProgress = (message: string) => {
                    progress.report({ message });
                };
                try {
                    return await task(updateProgress);
                } catch (error) {
                    const msg = error instanceof Error ? error.message : String(error);
                    vscode.window.showErrorMessage(`Query failed: ${msg}`);
                    this.log(`Query error: ${msg}`);
                    return undefined;
                }
            }
        );
    }

    /**
     * Query SYSMOD command
     */
    async querySysmod(): Promise<void> {
        const ctx = await this.getServerAndCredentials();
        if (!ctx) {
            return;
        }

        const zones = await this.promptForZones(ctx.server);
        if (!zones) {
            return;
        }

        const sysmods = await this.promptForSysmods();
        if (!sysmods) {
            return;
        }

        this.log(`Querying SYSMODs: ${sysmods.join(', ')} in zones: ${zones.join(', ')}`);

        const result = await this.executeWithProgress(
            `Querying SYSMODs on ${ctx.server.name}`,
            (progress) => this.client.querySysmod(
                ctx.server,
                ctx.credentials,
                zones,
                sysmods,
                progress
            )
        );

        if (result) {
            this.handleResult(ctx.server.name, 'sysmod', result);
        }
    }

    /**
     * Query DDDEF command
     */
    async queryDddef(): Promise<void> {
        const ctx = await this.getServerAndCredentials();
        if (!ctx) {
            return;
        }

        const zones = await this.promptForZones(ctx.server);
        if (!zones) {
            return;
        }

        const dddefs = await this.promptForDddefs();
        if (!dddefs) {
            return;
        }

        this.log(`Querying DDDEFs: ${dddefs.join(', ')} in zones: ${zones.join(', ')}`);

        const result = await this.executeWithProgress(
            `Querying DDDEFs on ${ctx.server.name}`,
            (progress) => this.client.queryDddef(
                ctx.server,
                ctx.credentials,
                zones,
                dddefs,
                progress
            )
        );

        if (result) {
            this.handleResult(ctx.server.name, 'dddef', result);
        }
    }

    /**
     * Query zones command
     */
    async queryZones(): Promise<void> {
        const ctx = await this.getServerAndCredentials();
        if (!ctx) {
            return;
        }

        this.log(`Querying zones on ${ctx.server.name}`);

        const result = await this.executeWithProgress(
            `Listing zones on ${ctx.server.name}`,
            (progress) => this.client.queryZones(
                ctx.server,
                ctx.credentials,
                progress
            )
        );

        if (result) {
            this.handleResult(ctx.server.name, 'zone', result);
        }
    }

    /**
     * Handle query result
     */
    private handleResult(serverName: string, queryType: QueryType, result: QueryResult): void {
        const displayResult: DisplayResult = {
            serverName,
            queryType,
            timestamp: new Date(),
            result
        };

        this.log(`Query completed: ${JSON.stringify(result).substring(0, 500)}...`);

        if (this.onResultCallback) {
            this.onResultCallback(displayResult);
        } else {
            // Fallback: show in output channel
            this.outputChannel.appendLine('\n=== Query Result ===');
            this.outputChannel.appendLine(`Server: ${serverName}`);
            this.outputChannel.appendLine(`Type: ${queryType}`);
            this.outputChannel.appendLine(`Time: ${displayResult.timestamp.toISOString()}`);
            this.outputChannel.appendLine(JSON.stringify(result, null, 2));
            this.outputChannel.show();
        }
    }
}
