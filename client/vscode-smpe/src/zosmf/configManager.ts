/**
 * z/OSMF Configuration Manager
 * Handles YAML config file and SecretStorage for passwords
 */

import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import * as yaml from 'yaml';
import { ZosmfConfig, ZosmfServer, Credentials } from './types';

const CONFIG_FILE_NAME = '.smpe-zosmf.yaml';
const SECRET_KEY_PREFIX = 'smpe.zosmf.password.';

/**
 * Configuration template for new config files
 */
const CONFIG_TEMPLATE = `# z/OSMF Connection Configurations
# Multiple servers can be defined

servers:
  - name: Production
    host: https://zosmf.mainframe.example.com
    port: 443
    csi: SMPE.GLOBAL.CSI
    user: USERID
    rejectUnauthorized: true
    zones:
      - GLOBAL
      - MVST100
      - MVSD100
    defaultZones:
      - GLOBAL

  # Add more servers as needed:
  # - name: Development
  #   host: https://dev.zosmf.example.com
  #   port: 443
  #   csi: SMPE.DEV.CSI
  #   user: DEVUSER
  #   rejectUnauthorized: false

# Optional: Default server name (if not set, user is always prompted)
# defaultServer: Production
`;

export class ConfigManager {
    private secretStorage: vscode.SecretStorage;
    private outputChannel: vscode.OutputChannel;

    constructor(context: vscode.ExtensionContext, outputChannel: vscode.OutputChannel) {
        this.secretStorage = context.secrets;
        this.outputChannel = outputChannel;
    }

    private log(message: string): void {
        const timestamp = new Date().toISOString();
        this.outputChannel.appendLine(`[${timestamp}] [ConfigManager] ${message}`);
    }

    /**
     * Get the path to the config file in the workspace
     */
    getConfigFilePath(): string | undefined {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        if (!workspaceFolders || workspaceFolders.length === 0) {
            return undefined;
        }
        return path.join(workspaceFolders[0].uri.fsPath, CONFIG_FILE_NAME);
    }

    /**
     * Check if config file exists in workspace
     */
    configExists(): boolean {
        const configPath = this.getConfigFilePath();
        return configPath !== undefined && fs.existsSync(configPath);
    }

    /**
     * Create a new config file with template
     */
    async createConfigFile(): Promise<boolean> {
        const configPath = this.getConfigFilePath();
        if (!configPath) {
            vscode.window.showErrorMessage('No workspace folder open. Please open a folder first.');
            return false;
        }

        if (fs.existsSync(configPath)) {
            const overwrite = await vscode.window.showWarningMessage(
                `${CONFIG_FILE_NAME} already exists. Overwrite?`,
                'Yes', 'No'
            );
            if (overwrite !== 'Yes') {
                return false;
            }
        }

        try {
            fs.writeFileSync(configPath, CONFIG_TEMPLATE, 'utf8');
            this.log(`Created config file: ${configPath}`);

            // Open the file in editor
            const doc = await vscode.workspace.openTextDocument(configPath);
            await vscode.window.showTextDocument(doc);

            vscode.window.showInformationMessage(`Created ${CONFIG_FILE_NAME}. Please edit server configurations.`);
            return true;
        } catch (error) {
            const msg = error instanceof Error ? error.message : String(error);
            vscode.window.showErrorMessage(`Failed to create config file: ${msg}`);
            this.log(`Error creating config file: ${msg}`);
            return false;
        }
    }

    /**
     * Load and parse the config file
     */
    loadConfig(): ZosmfConfig | undefined {
        const configPath = this.getConfigFilePath();
        if (!configPath) {
            vscode.window.showErrorMessage('No workspace folder open.');
            return undefined;
        }

        if (!fs.existsSync(configPath)) {
            vscode.window.showErrorMessage(
                `Config file not found. Run "SMP/E: Create z/OSMF Config" first.`,
                'Create Config'
            ).then(selection => {
                if (selection === 'Create Config') {
                    vscode.commands.executeCommand('smpe.zosmf.createConfig');
                }
            });
            return undefined;
        }

        try {
            const content = fs.readFileSync(configPath, 'utf8');
            const config = yaml.parse(content) as ZosmfConfig;

            // Validate config
            if (!config.servers || !Array.isArray(config.servers) || config.servers.length === 0) {
                vscode.window.showErrorMessage('Config file has no servers defined.');
                return undefined;
            }

            // Validate each server
            for (const server of config.servers) {
                if (!server.name || !server.host || !server.csi || !server.user) {
                    vscode.window.showErrorMessage(`Server "${server.name || 'unnamed'}" is missing required fields (name, host, csi, user).`);
                    return undefined;
                }
                // Set default port if not specified
                if (!server.port) {
                    server.port = 443;
                }
                // Set default rejectUnauthorized if not specified
                if (server.rejectUnauthorized === undefined) {
                    server.rejectUnauthorized = true;
                }
            }

            this.log(`Loaded config with ${config.servers.length} server(s)`);
            return config;
        } catch (error) {
            const msg = error instanceof Error ? error.message : String(error);
            vscode.window.showErrorMessage(`Failed to parse config file: ${msg}`);
            this.log(`Error parsing config: ${msg}`);
            return undefined;
        }
    }

    /**
     * Select a server from the config
     * Returns undefined if user cancels
     */
    async selectServer(config: ZosmfConfig): Promise<ZosmfServer | undefined> {
        // If only one server, use it directly
        if (config.servers.length === 1) {
            this.log(`Using single server: ${config.servers[0].name}`);
            return config.servers[0];
        }

        // If default server is set and exists, offer it as default
        const defaultServer = config.defaultServer
            ? config.servers.find(s => s.name === config.defaultServer)
            : undefined;

        // Build QuickPick items
        const items: vscode.QuickPickItem[] = config.servers.map(server => ({
            label: server.name,
            description: `${server.host} - ${server.csi}`,
            detail: server.defaultZones?.length
                ? `Default zones: ${server.defaultZones.join(', ')}`
                : undefined,
            picked: server === defaultServer
        }));

        const selected = await vscode.window.showQuickPick(items, {
            placeHolder: 'Select z/OSMF server',
            title: 'z/OSMF Server Selection'
        });

        if (!selected) {
            return undefined;
        }

        const server = config.servers.find(s => s.name === selected.label);
        this.log(`Selected server: ${server?.name}`);
        return server;
    }

    /**
     * Get the secret storage key for a server
     */
    private getSecretKey(server: ZosmfServer): string {
        return `${SECRET_KEY_PREFIX}${server.user}@${server.host}`;
    }

    /**
     * Get stored password for a server
     */
    async getPassword(server: ZosmfServer): Promise<string | undefined> {
        const key = this.getSecretKey(server);
        const password = await this.secretStorage.get(key);
        if (password) {
            this.log(`Retrieved stored password for ${server.user}@${server.host}`);
        }
        return password;
    }

    /**
     * Store password for a server
     */
    async storePassword(server: ZosmfServer, password: string): Promise<void> {
        const key = this.getSecretKey(server);
        await this.secretStorage.store(key, password);
        this.log(`Stored password for ${server.user}@${server.host}`);
    }

    /**
     * Delete stored password for a server
     */
    async deletePassword(server: ZosmfServer): Promise<void> {
        const key = this.getSecretKey(server);
        await this.secretStorage.delete(key);
        this.log(`Deleted password for ${server.user}@${server.host}`);
    }

    /**
     * Clear all stored passwords
     */
    async clearAllPasswords(): Promise<void> {
        const config = this.loadConfig();
        if (!config) {
            return;
        }

        for (const server of config.servers) {
            await this.deletePassword(server);
        }
        this.log('Cleared all stored passwords');
        vscode.window.showInformationMessage('All stored z/OSMF passwords have been cleared.');
    }

    /**
     * Prompt user for password and optionally store it
     */
    async promptForPassword(server: ZosmfServer): Promise<string | undefined> {
        const password = await vscode.window.showInputBox({
            prompt: `Enter password for ${server.user}@${server.host}`,
            password: true,
            ignoreFocusOut: true
        });

        if (password) {
            // Ask if user wants to save the password
            const save = await vscode.window.showQuickPick(['Yes', 'No'], {
                placeHolder: 'Save password securely for future use?'
            });

            if (save === 'Yes') {
                await this.storePassword(server, password);
            }
        }

        return password;
    }

    /**
     * Get credentials for a server (from storage or prompt)
     */
    async getCredentials(server: ZosmfServer): Promise<Credentials | undefined> {
        let password = await this.getPassword(server);

        if (!password) {
            password = await this.promptForPassword(server);
        }

        if (!password) {
            return undefined;
        }

        return {
            user: server.user,
            password: password
        };
    }
}
