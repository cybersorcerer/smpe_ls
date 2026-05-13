import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import * as yaml from 'yaml';
import { SavedQuery, SavedQueriesConfig } from './types';

const SAVED_QUERIES_FILE = '.smpe-saved-queries.yaml';

export class SavedQueriesManager {
    private outputChannel: vscode.OutputChannel;

    constructor(outputChannel: vscode.OutputChannel) {
        this.outputChannel = outputChannel;
    }

    private log(message: string): void {
        const timestamp = new Date().toISOString();
        this.outputChannel.appendLine(`[${timestamp}] [SavedQueriesManager] ${message}`);
    }

    private getFilePath(): string | undefined {
        const folders = vscode.workspace.workspaceFolders;
        if (!folders || folders.length === 0) { return undefined; }
        return path.join(folders[0].uri.fsPath, SAVED_QUERIES_FILE);
    }

    loadQueries(): SavedQuery[] {
        const filePath = this.getFilePath();
        if (!filePath || !fs.existsSync(filePath)) { return []; }
        try {
            const content = fs.readFileSync(filePath, 'utf8');
            const config = yaml.parse(content) as SavedQueriesConfig;
            if (!config?.queries || !Array.isArray(config.queries)) { return []; }
            return config.queries;
        } catch (error) {
            this.log(`Error loading saved queries: ${error instanceof Error ? error.message : String(error)}`);
            return [];
        }
    }

    saveQuery(query: SavedQuery): void {
        const filePath = this.getFilePath();
        if (!filePath) {
            vscode.window.showErrorMessage('No workspace folder open. Cannot save query.');
            return;
        }
        const isNew = !fs.existsSync(filePath);
        const queries = isNew ? [] : this.loadQueries();
        const idx = queries.findIndex(q => q.name === query.name);
        if (idx >= 0) {
            queries[idx] = query;
        } else {
            queries.push(query);
        }
        const config: SavedQueriesConfig = { queries };
        fs.writeFileSync(filePath, yaml.stringify(config), 'utf8');
        this.log(`Saved query "${query.name}" to ${filePath}`);
        if (isNew) {
            vscode.window.showInformationMessage(`Datei ${SAVED_QUERIES_FILE} wurde im Projekt-Root erstellt.`);
        }
    }

    deleteQuery(name: string): void {
        const filePath = this.getFilePath();
        if (!filePath || !fs.existsSync(filePath)) { return; }
        const queries = this.loadQueries().filter(q => q.name !== name);
        const config: SavedQueriesConfig = { queries };
        fs.writeFileSync(filePath, yaml.stringify(config), 'utf8');
        this.log(`Deleted query "${name}"`);
    }
}
