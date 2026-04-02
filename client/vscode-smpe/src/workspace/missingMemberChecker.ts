/**
 * Missing Input Member Checker
 * Checks whether input files referenced by MCS statements exist in the workspace.
 */

import * as vscode from 'vscode';
import * as path from 'path';

export interface MemberCheckResult {
    smpeFile: string;
    statement: string;
    member: string;
    found: boolean;
    foundPath?: string;
}

const STATEMENT_FILE_MAP: Record<string, string> = {
    '++EXEC':     '.rexx',
    '++PARM':     '.parm',
    '++TBL':      '.tbl',
    '++SRC':      '.hlasm',
    '++MAC':      '.hlasm',
    '++SKL':      '.skl.jcl',
    '++MSG':      '.msg',
    '++HFS':      '.hfs',
    '++MOD':      '.mod',
    '++ZAP':      '.zap',
    '++CLIST':    '.clist',
    '++USER':     '.usr',
    '++PROC':     '.proc',
    '++DATA':     '.data',
    '++SAMP':     '.samp',
    '++HELP':     '.help',
    '++SHSCRIPT': '.sh',
    '++SHELLSRC': '.sh',
    '++PROGRAM':  '.bin',
    '++PNLENU':   '.enu.pnl',
};

export class MissingMemberChecker {
    private outputChannel: vscode.OutputChannel;
    // Cache: smpeFilePath -> results
    private resultCache: Map<string, MemberCheckResult[]> = new Map();
    // Index of all workspace files: filename -> full fsPath
    private fileIndex: Map<string, string> | undefined;

    constructor(outputChannel: vscode.OutputChannel) {
        this.outputChannel = outputChannel;
    }

    private log(message: string): void {
        const timestamp = new Date().toISOString();
        this.outputChannel.appendLine(`[${timestamp}] [MissingMemberChecker] ${message}`);
    }

    /**
     * Invalidate the cache for a specific .smpe file (called on save).
     */
    public invalidate(smpeFilePath: string): void {
        this.resultCache.delete(smpeFilePath);
        this.log(`Cache invalidated for ${smpeFilePath}`);
    }

    /**
     * Invalidate the entire file index (called when workspace files change).
     */
    public invalidateFileIndex(): void {
        this.fileIndex = undefined;
        this.resultCache.clear();
        this.log('File index invalidated');
    }

    /**
     * Build an index of all files in the configured search folders.
     * Key: exact filename (basename), Value: full fsPath (first match wins).
     */
    private async buildFileIndex(): Promise<Map<string, string>> {
        const config = vscode.workspace.getConfiguration('smpe');
        const searchFolders = config.get<string[]>('checkMissingInputMembers.searchFolders', ['customization']);

        const index = new Map<string, string>();
        let pattern: string;

        if (searchFolders.includes('*')) {
            pattern = '**/*';
        } else {
            if (searchFolders.length === 1) {
                pattern = `${searchFolders[0]}/**/*`;
            } else {
                pattern = `{${searchFolders.join(',')}}/**/*`;
            }
        }

        this.log(`Building file index with pattern: ${pattern}`);
        const uris = await vscode.workspace.findFiles(pattern, '**/node_modules/**');

        for (const uri of uris) {
            const basename = path.basename(uri.fsPath);
            if (!index.has(basename)) {
                index.set(basename, uri.fsPath);
            }
        }

        this.log(`File index built: ${index.size} files indexed`);
        return index;
    }

    /**
     * Run the check for all *.smpe files in the workspace.
     * Uses per-file cache; rebuilds file index if needed.
     */
    public async check(): Promise<MemberCheckResult[]> {
        if (!this.fileIndex) {
            this.fileIndex = await this.buildFileIndex();
        }

        const smpeFiles = await vscode.workspace.findFiles('**/*.smpe', '**/node_modules/**');
        this.log(`Found ${smpeFiles.length} .smpe files`);

        const allResults: MemberCheckResult[] = [];

        for (const uri of smpeFiles) {
            const filePath = uri.fsPath;
            if (this.resultCache.has(filePath)) {
                allResults.push(...this.resultCache.get(filePath)!);
                continue;
            }

            const fileResults = await this.checkFile(uri);
            this.resultCache.set(filePath, fileResults);
            allResults.push(...fileResults);
        }

        return allResults;
    }

    private async checkFile(uri: vscode.Uri): Promise<MemberCheckResult[]> {
        const results: MemberCheckResult[] = [];
        const filePath = uri.fsPath;
        const smpeFileName = path.basename(filePath);

        let content: string;
        try {
            const bytes = await vscode.workspace.fs.readFile(uri);
            content = Buffer.from(bytes).toString('utf8');
        } catch (err) {
            this.log(`Error reading ${filePath}: ${err}`);
            return results;
        }

        const statements = this.parseStatements(content);

        for (const stmt of statements) {
            const ext = STATEMENT_FILE_MAP[stmt.name];
            if (!ext) {
                continue;
            }
            if (!stmt.elementName) {
                continue;
            }
            if (stmt.hasInlineData) {
                continue;
            }
            if (stmt.hasTxlib) {
                continue;
            }

            const expectedFile = stmt.elementName + ext;
            const foundPath = this.fileIndex!.get(expectedFile);

            results.push({
                smpeFile: smpeFileName,
                statement: stmt.name,
                member: expectedFile,
                found: foundPath !== undefined,
                foundPath: foundPath,
            });
        }

        return results;
    }

    /**
     * Minimal statement parser — extracts the information needed for the check
     * without depending on the Go LSP server at runtime.
     */
    private parseStatements(content: string): ParsedStatement[] {
        const statements: ParsedStatement[] = [];
        const lines = content.split('\n');

        let i = 0;
        while (i < lines.length) {
            const line = lines[i];
            const trimmed = line.trimStart();

            // Skip comment lines
            if (trimmed.startsWith('++')) {
                // Extract statement name and optional parameter
                const match = trimmed.match(/^(\+\+[A-Z]+)\s*(?:\(([^)]*)\))?/);
                if (match) {
                    const stmtName = match[1];
                    const elementName = match[2] ? match[2].trim() : undefined;

                    if (STATEMENT_FILE_MAP[stmtName] || this.isAnyStatement(stmtName)) {
                        // Collect operands: scan until next ++ or terminator
                        let hasTxlib = false;
                        let hasInlineData = false;
                        let terminated = false;

                        // Check same line for terminator and TXLIB
                        const restOfLine = trimmed.slice(match[0].length);
                        if (restOfLine.includes('TXLIB(') || restOfLine.includes('TXLIB (')) {
                            hasTxlib = true;
                        }
                        if (restOfLine.includes('.')) {
                            terminated = true;
                        }

                        // Scan continuation lines
                        let j = i + 1;
                        while (!terminated && j < lines.length) {
                            const nextTrimmed = lines[j].trimStart();

                            // New ++ statement starts
                            if (nextTrimmed.startsWith('++')) {
                                break;
                            }

                            // Terminator line
                            if (nextTrimmed.startsWith('.')) {
                                terminated = true;
                                j++;
                                break;
                            }

                            // Check for TXLIB operand
                            if (nextTrimmed.includes('TXLIB(') || nextTrimmed.includes('TXLIB (')) {
                                hasTxlib = true;
                            }

                            // If terminated on previous line, check for inline data after terminator
                            j++;
                        }

                        // Inline data: any non-empty, non-comment line after the statement block
                        // that is NOT a ++ statement and comes before the next ++ statement
                        if (!hasTxlib && STATEMENT_FILE_MAP[stmtName]) {
                            // Scan from j looking for inline data
                            let k = j;
                            while (k < lines.length) {
                                const candidate = lines[k].trimStart();
                                if (candidate.startsWith('++')) {
                                    break;
                                }
                                // Non-empty, non-comment line = inline data
                                if (candidate.length > 0 && !candidate.startsWith('*')) {
                                    hasInlineData = true;
                                    break;
                                }
                                k++;
                            }
                        }

                        if (STATEMENT_FILE_MAP[stmtName]) {
                            statements.push({ name: stmtName, elementName, hasTxlib, hasInlineData });
                        }

                        i = j;
                        continue;
                    }
                }
            }

            i++;
        }

        return statements;
    }

    private isAnyStatement(name: string): boolean {
        return name.startsWith('++');
    }
}

interface ParsedStatement {
    name: string;
    elementName: string | undefined;
    hasTxlib: boolean;
    hasInlineData: boolean;
}
