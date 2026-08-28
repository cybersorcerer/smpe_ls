/**
 * Missing Input Member Checker
 *
 * Uses smpe_outl --json --meta --ranges to parse .smpe files and checks that
 * every referenced input member can be resolved, so a usermod can actually be
 * assembled by the pipeline. There are two ways a statement points at its
 * input member, and both are checked:
 *
 *  - placeholder: a "{{ path }}" line in the statement's inline data area.
 *    The path is relative to the repository root and is checked as written -
 *    it may point anywhere, not just into the search folders. The GitLab
 *    pipeline replaces such a line with the contents of that file.
 *  - convention: no placeholder and no inline data, so the member is expected
 *    as "<element name><extension>" somewhere below the configured search
 *    folders (smpe.checkMissingInputMembers.searchFolders, default
 *    "customization"), matched on file name alone.
 */

import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import * as cp from 'child_process';

/** How the expected input member was derived. */
export type MemberSource = 'placeholder' | 'convention';

export interface MemberCheckResult {
    smpeFile: string;
    statement: string;
    member: string;
    found: boolean;
    foundPath?: string;
    source: MemberSource;
}

/**
 * A "{{ path }}" line. The pipeline accepts an optional leading "./", optional
 * spaces inside the braces and an indented line, so all of those are matched
 * here as well.
 */
const PLACEHOLDER_RE = /^\s*\{\{\s*(.+?)\s*\}\}\s*$/;

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
    '++PROC':     '.jcl',
    '++DATA':     '.data',
    '++SAMP':     '.samp',
    '++HELP':     '.help',
    '++SHSCRIPT': '.sh',
    '++SHELLSRC': '.sh',
    '++PROGRAM':  '.bin',
    '++PNLENU':   '.enu.pnl',
};

// Shape of smpe_outl --json --meta --ranges output
interface OutlinePosition {
    line: number;
    character: number;
}

interface OutlineRange {
    start: OutlinePosition;
    end: OutlinePosition;
}

interface OutlineSymbol {
    name: string;
    id?: string;
    hasInlineData?: boolean;
    range?: OutlineRange;
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

        // --ranges is needed to locate each statement's inline data area,
        // which is where the "{{ path }}" placeholders live.
        const args = ['--json', '--meta', '--ranges'];
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
            const lines = this.readLines(outline.file);
            const repoRoot = this.repoRootFor(outline.file);
            const fileResults = this.processOutline(outline, lines, repoRoot);
            this.resultCache.set(outline.file, fileResults);
            fresh.push(...fileResults);
        }

        return [...cached, ...fresh];
    }

    /**
     * Read the source lines of a .smpe file, or undefined if unreadable.
     * Needed because smpe_outl reports statement ranges but not the inline
     * data content in which the placeholders sit.
     */
    private readLines(filePath: string): string[] | undefined {
        try {
            return fs.readFileSync(filePath, 'utf8').split('\n');
        } catch (err) {
            this.log(`Cannot read ${filePath}: ${err}`);
            return undefined;
        }
    }

    /**
     * Repository root for a .smpe file: the workspace folder containing it.
     * Placeholder paths are relative to this.
     */
    private repoRootFor(filePath: string): string | undefined {
        const folder = vscode.workspace.getWorkspaceFolder(vscode.Uri.file(filePath));
        return folder?.uri.fsPath;
    }

    /**
     * Collect the "{{ path }}" placeholders in lines [from, to).
     * A statement's inline data starts after its range end and runs up to the
     * next statement, so that span is what gets scanned.
     */
    private placeholdersIn(lines: string[], from: number, to: number): string[] {
        const found: string[] = [];
        for (let i = Math.max(0, from); i < Math.min(to, lines.length); i++) {
            const m = lines[i].match(PLACEHOLDER_RE);
            if (m) {
                found.push(m[1]);
            }
        }
        return found;
    }

    private processOutline(outline: OutlineFile, lines: string[] | undefined, repoRoot: string | undefined): MemberCheckResult[] {
        const results: MemberCheckResult[] = [];
        const smpeFileName = path.basename(outline.file);
        const symbols = outline.symbols ?? [];

        for (let i = 0; i < symbols.length; i++) {
            const sym = symbols[i];

            // Extract statement name from symbol name (e.g. "++PARM(RSSPRMI)" -> "++PARM")
            const nameMatch = sym.name.match(/^(\+\+[A-Z]+)/);
            if (!nameMatch) { continue; }
            const stmtName = nameMatch[1];

            // 1. Placeholder check. Independent of STATEMENT_FILE_MAP: the path
            // is stated explicitly, so it works for statements the map does not
            // cover (++JCLIN and friends) too.
            const placeholders = this.placeholderPathsFor(symbols, i, lines);
            if (placeholders.length > 0) {
                for (const rel of placeholders) {
                    results.push(this.checkPlaceholder(smpeFileName, stmtName, rel, repoRoot));
                }
                continue;
            }

            // 2. Convention check. Only for statements whose element file name
            // can be derived, and only when the data is not supplied otherwise.
            const ext = STATEMENT_FILE_MAP[stmtName];
            if (!ext) { continue; }

            // Element name is in the id field (e.g. "RSSPRMI")
            const elementName = sym.id;
            if (!elementName) { continue; }

            // Skip if statement carries its inline data directly
            if (sym.hasInlineData) { continue; }

            // Skip if an operand supplies the data from elsewhere, or deletes
            // the element - in those cases no member file is expected.
            const hasExternalSource = (sym.children ?? []).some(c =>
                c.name.startsWith('TXLIB(') ||
                c.name.startsWith('FROMDS(') ||
                c.name.startsWith('RELFILE(') ||
                c.name === 'DELETE' || c.name.startsWith('DELETE('));
            if (hasExternalSource) { continue; }

            const expectedFile = elementName + ext;
            const foundPath = this.fileIndex!.get(expectedFile);

            results.push({
                smpeFile: smpeFileName,
                statement: stmtName,
                member: expectedFile,
                found: foundPath !== undefined,
                foundPath,
                source: 'convention',
            });
        }

        return results;
    }

    /** Placeholder paths in the inline data area following symbols[i]. */
    private placeholderPathsFor(symbols: OutlineSymbol[], i: number, lines: string[] | undefined): string[] {
        if (!lines) { return []; }
        const end = symbols[i].range?.end.line;
        if (end === undefined) { return []; }

        // The inline data runs from the line after this statement up to the
        // next statement, or to the end of the file for the last one.
        const nextStart = symbols[i + 1]?.range?.start.line ?? lines.length;
        return this.placeholdersIn(lines, end + 1, nextStart);
    }

    /** Resolve a placeholder path against the repository root. */
    private checkPlaceholder(smpeFileName: string, stmtName: string, rel: string, repoRoot: string | undefined): MemberCheckResult {
        if (!repoRoot) {
            this.log(`No workspace folder for ${smpeFileName}, cannot resolve "${rel}"`);
            return { smpeFile: smpeFileName, statement: stmtName, member: rel, found: false, source: 'placeholder' };
        }

        const abs = path.resolve(repoRoot, rel);
        const exists = fs.existsSync(abs) && fs.statSync(abs).isFile();

        return {
            smpeFile: smpeFileName,
            statement: stmtName,
            member: rel,
            found: exists,
            foundPath: exists ? abs : undefined,
            source: 'placeholder',
        };
    }
}
