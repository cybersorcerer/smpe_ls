/**
 * Missing Input Member Checker
 * Uses smpe_outl --json --meta to parse .smpe files and checks whether
 * the referenced input member files exist in the configured search folders.
 */

import * as vscode from 'vscode';
import * as path from 'path';
import * as cp from 'child_process';

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

// Shape of smpe_outl --json --meta output
interface OutlineSymbol {
    name: string;
    id?: string;
    hasInlineData?: boolean;
    children?: OutlineSymbol[];
}

interface OutlineFile {
    file: string;
    symbols: OutlineSymbol[];
}

export class MissingMemberChecker {
    private outputChannel: vscode.OutputChannel;
    private outlBinaryPath: string;
    private dataBinaryPath: string;

    // Cache: smpeFilePath -> results
    private resultCache: Map<string, MemberCheckResult[]> = new Map();
    // Index of all workspace files: filename -> full fsPath (first match wins)
    private fileIndex: Map<string, string> | undefined;

    constructor(outputChannel: vscode.OutputChannel, outlBinaryPath: string, dataBinaryPath: string) {
        this.outputChannel = outputChannel;
        this.outlBinaryPath = outlBinaryPath;
        this.dataBinaryPath = dataBinaryPath;
    }

    private log(message: string): void {
        const timestamp = new Date().toISOString();
        this.outputChannel.appendLine(`[${timestamp}] [MissingMemberChecker] ${message}`);
    }

    /** Invalidate the per-file cache entry (called on .smpe save). */
    public invalidate(smpeFilePath: string): void {
        this.resultCache.delete(smpeFilePath);
        this.log(`Cache invalidated for ${smpeFilePath}`);
    }

    /** Invalidate the entire file index and all cached results. */
    public invalidateFileIndex(): void {
        this.fileIndex = undefined;
        this.resultCache.clear();
        this.log('File index invalidated');
    }

    /** Update the binary path when the setting changes. */
    public setOutlBinaryPath(p: string): void {
        this.outlBinaryPath = p;
        this.resultCache.clear();
    }

    private async buildFileIndex(): Promise<Map<string, string>> {
        const config = vscode.workspace.getConfiguration('smpe');
        const searchFolders = config.get<string[]>('checkMissingInputMembers.searchFolders', ['customization']);

        const index = new Map<string, string>();
        let pattern: string;

        if (searchFolders.includes('*')) {
            pattern = '**/*';
        } else if (searchFolders.length === 1) {
            pattern = `${searchFolders[0]}/**/*`;
        } else {
            pattern = `{${searchFolders.join(',')}}/**/*`;
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

    /** Run smpe_outl --json --meta on all .smpe files and return parsed outlines. */
    private async runOutl(smpeFiles: vscode.Uri[]): Promise<OutlineFile[]> {
        if (smpeFiles.length === 0) {
            return [];
        }

        const args = ['--json', '--meta'];
        if (this.dataBinaryPath) {
            args.push('--data', this.dataBinaryPath);
        }
        args.push(...smpeFiles.map(u => u.fsPath));

        this.log(`Running: ${this.outlBinaryPath} ${args.join(' ')}`);

        return new Promise((resolve, reject) => {
            cp.execFile(this.outlBinaryPath, args, { maxBuffer: 10 * 1024 * 1024 }, (err, stdout, stderr) => {
                if (stderr) {
                    this.log(`smpe_outl stderr: ${stderr.trim()}`);
                }
                if (err && !stdout) {
                    reject(new Error(`smpe_outl failed: ${err.message}`));
                    return;
                }
                try {
                    resolve(JSON.parse(stdout) as OutlineFile[]);
                } catch (parseErr) {
                    reject(new Error(`smpe_outl JSON parse error: ${parseErr}`));
                }
            });
        });
    }

    /** Run the full check. Returns results for all .smpe files in the workspace. */
    public async check(): Promise<MemberCheckResult[]> {
        if (!this.fileIndex) {
            this.fileIndex = await this.buildFileIndex();
        }

        const allSmpeFiles = await vscode.workspace.findFiles('**/*.smpe', '**/node_modules/**');
        this.log(`Found ${allSmpeFiles.length} .smpe files`);

        // Split into cached and uncached files
        const cached: MemberCheckResult[] = [];
        const uncached: vscode.Uri[] = [];

        for (const uri of allSmpeFiles) {
            if (this.resultCache.has(uri.fsPath)) {
                cached.push(...this.resultCache.get(uri.fsPath)!);
            } else {
                uncached.push(uri);
            }
        }

        if (uncached.length === 0) {
            return cached;
        }

        // Run smpe_outl on all uncached files in one invocation
        let outlines: OutlineFile[];
        try {
            outlines = await this.runOutl(uncached);
        } catch (err) {
            this.log(`ERROR: ${err}`);
            vscode.window.showErrorMessage(`SMP/E: smpe_outl failed — ${err}`);
            return cached;
        }

        // Process results per file
        const fresh: MemberCheckResult[] = [];
        for (const outline of outlines) {
            const fileResults = this.processOutline(outline);
            this.resultCache.set(outline.file, fileResults);
            fresh.push(...fileResults);
        }

        return [...cached, ...fresh];
    }

    private processOutline(outline: OutlineFile): MemberCheckResult[] {
        const results: MemberCheckResult[] = [];
        const smpeFileName = path.basename(outline.file);

        for (const sym of outline.symbols ?? []) {
            // Extract statement name from symbol name (e.g. "++PARM(RSSPRMI)" -> "++PARM")
            const nameMatch = sym.name.match(/^(\+\+[A-Z]+)/);
            if (!nameMatch) { continue; }
            const stmtName = nameMatch[1];

            const ext = STATEMENT_FILE_MAP[stmtName];
            if (!ext) { continue; }

            // Element name is in the id field (e.g. "RSSPRMI")
            const elementName = sym.id;
            if (!elementName) { continue; }

            // Skip if statement has inline data
            if (sym.hasInlineData) { continue; }

            // Skip if TXLIB operand is present (appears as child symbol)
            const hasTxlib = (sym.children ?? []).some(c => c.name.startsWith('TXLIB('));
            if (hasTxlib) { continue; }

            const expectedFile = elementName + ext;
            const foundPath = this.fileIndex!.get(expectedFile);

            results.push({
                smpeFile: smpeFileName,
                statement: stmtName,
                member: expectedFile,
                found: foundPath !== undefined,
                foundPath,
            });
        }

        return results;
    }
}
