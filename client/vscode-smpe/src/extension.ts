import * as path from 'path';
import * as fs from 'fs';
import * as vscode from 'vscode';
import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
	Executable
} from 'vscode-languageclient/node';
import { QueryProvider } from './zosmf/queryProvider';
import { ResultPanel } from './webview/resultPanel';
import { UssPanel, UssFileContentProvider } from './webview/ussPanel';
import { DatasetPanel, DatasetContentProvider } from './webview/datasetPanel';
import { FreeFormPanel } from './webview/freeFormPanel';
import { MissingMemberChecker } from './workspace/missingMemberChecker';
import { MissingMemberPanel } from './webview/missingMemberPanel';

let client: LanguageClient;
let outputChannel: vscode.OutputChannel;
let queryProvider: QueryProvider;

function log(message: string) {
	const timestamp = new Date().toISOString();
	outputChannel.appendLine(`[${timestamp}] ${message}`);
	console.log(message);
}

function debugLog(message: string) {
	const debug = vscode.workspace.getConfiguration('smpe').get<boolean>('debug', true);
	if (!debug) { return; }
	log(message);
}

export function activate(context: vscode.ExtensionContext) {
	// Create output channel for logging
	outputChannel = vscode.window.createOutputChannel('SMP/E Language Server');
	// Note: outputChannel.show() removed to avoid opening terminal on startup
	// Users can manually open "Output" panel and select "SMP/E Language Server" if needed

	log('SMP/E Language Server extension activating...');
	log(`Platform: ${process.platform}`);
	log(`Extension path: ${context.extensionPath}`);

	// Get configuration
	const config = vscode.workspace.getConfiguration('smpe');
	const configuredServerPath = config.get<string>('serverPath') || '';
	const configuredDataPath = config.get<string>('dataPath') || '';
	const debug = config.get<boolean>('debug') || false;

	debugLog(`Configured serverPath: "${configuredServerPath}"`);
	debugLog(`Configured dataPath: "${configuredDataPath}"`);
	log(`Debug mode: ${debug}`);

	// Bundled paths
	const bundledBinaryName = process.platform === 'win32' ? 'smpe_ls.exe' : 'smpe_ls';
	const bundledBinaryPath = path.join(context.extensionPath, bundledBinaryName);
	const bundledDataPath = path.join(context.extensionPath, 'smpe.json');

	// Determine the full path to the server
	let executable = '';
	let dataPath = '';

	// Priority 1: User configured server path (takes precedence over bundled)
	if (configuredServerPath) {
		debugLog(`Checking configured server path: ${configuredServerPath}`);
		if (fs.existsSync(configuredServerPath)) {
			log(`Using configured server path`);
			executable = configuredServerPath;
		} else {
			log(`WARNING: Configured serverPath does not exist`);
		}
	}

	// Priority 2: Bundled binary
	if (!executable) {
		debugLog(`Looking for bundled binary at: ${bundledBinaryPath}`);
		if (fs.existsSync(bundledBinaryPath)) {
			debugLog(`Found bundled binary`);
			executable = bundledBinaryPath;
		} else {
			debugLog(`Bundled binary NOT found`);
		}
	}

	// Priority 3: Fallback to ~/.local/bin
	if (!executable) {
		const homeDir = process.env.HOME || process.env.USERPROFILE;
		if (homeDir) {
			const localBinPath = path.join(homeDir, '.local', 'bin', 'smpe_ls');
			debugLog(`Trying fallback path: ${localBinPath}`);
			if (fs.existsSync(localBinPath)) {
				debugLog(`Found server at fallback path`);
				executable = localBinPath;
			} else {
				debugLog(`Server NOT found at fallback path`);
			}
		}
	}

	// Determine data path
	// Priority 1: User configured data path
	if (configuredDataPath) {
		debugLog(`Checking configured data path: ${configuredDataPath}`);
		if (fs.existsSync(configuredDataPath)) {
			debugLog(`Using configured data path`);
			dataPath = configuredDataPath;
		} else {
			log(`WARNING: Configured dataPath does not exist`);
		}
	}

	// Priority 2: Bundled data file
	if (!dataPath) {
		debugLog(`Looking for bundled data at: ${bundledDataPath}`);
		if (fs.existsSync(bundledDataPath)) {
			debugLog(`Found bundled data`);
			dataPath = bundledDataPath;
		} else {
			debugLog(`Bundled data NOT found`);
		}
	}

	// Build data path arguments
	const dataPathArgs = dataPath ? ['--data', dataPath] : [];

	// Final check
	if (!executable) {
		const errorMsg = 'SMP/E Language Server binary not found. Please install it or configure smpe.serverPath.';
		log(`ERROR: ${errorMsg}`);
		vscode.window.showErrorMessage(errorMsg);
		return;
	}

	if (!fs.existsSync(executable)) {
		const errorMsg = `SMP/E Language Server binary not found at: ${executable}`;
		log(`ERROR: ${errorMsg}`);
		vscode.window.showErrorMessage(errorMsg);
		return;
	}

	// Build arguments
	const args = (debug ? ['--debug'] : []).concat(dataPathArgs);

	log(`Starting server: ${executable}`);
	log(`Arguments: ${args.join(' ')}`);

	const serverExecutable: Executable = {
		command: executable,
		args: args,
		options: {
			env: process.env,
			shell: process.platform === 'win32'
		}
	};

	const serverOptions: ServerOptions = {
		run: serverExecutable,
		debug: serverExecutable
	};

	// Build initialization options with diagnostic settings
	const diagnosticsConfig = {
		unknownStatement: config.get<boolean>('diagnostics.unknownStatement', true),
		invalidLanguageId: config.get<boolean>('diagnostics.invalidLanguageId', true),
		unbalancedParentheses: config.get<boolean>('diagnostics.unbalancedParentheses', true),
		missingTerminator: config.get<boolean>('diagnostics.missingTerminator', true),
		missingParameter: config.get<boolean>('diagnostics.missingParameter', true),
		unknownOperand: config.get<boolean>('diagnostics.unknownOperand', true),
		duplicateOperand: config.get<boolean>('diagnostics.duplicateOperand', true),
		emptyOperandParameter: config.get<boolean>('diagnostics.emptyOperandParameter', true),
		missingRequiredOperand: config.get<boolean>('diagnostics.missingRequiredOperand', true),
		dependencyViolation: config.get<boolean>('diagnostics.dependencyViolation', true),
		mutuallyExclusive: config.get<boolean>('diagnostics.mutuallyExclusive', true),
		requiredGroup: config.get<boolean>('diagnostics.requiredGroup', true),
		missingInlineData: config.get<boolean>('diagnostics.missingInlineData', true),
		unknownSubOperand: config.get<boolean>('diagnostics.unknownSubOperand', true),
		subOperandValidation: config.get<boolean>('diagnostics.subOperandValidation', true),
		contentBeyondColumn72: config.get<boolean>('diagnostics.contentBeyondColumn72', true),
		standaloneCommentBetweenMCS: config.get<boolean>('diagnostics.standaloneCommentBetweenMCS', true)
	};

	// Build formatting configuration
	const formattingConfig = {
		enabled: config.get<boolean>('formatting.enabled', true),
		indentContinuation: config.get<number>('formatting.indentContinuation', 3),
		oneOperandPerLine: config.get<boolean>('formatting.oneOperandPerLine', true),
		wrapListsAfterN: config.get<number>('formatting.wrapListsAfterN', 2),
		moveLeadingComments: config.get<boolean>('formatting.moveLeadingComments', false)
	};

	debugLog(`Diagnostics config: ${JSON.stringify(diagnosticsConfig)}`);
	debugLog(`Formatting config: ${JSON.stringify(formattingConfig)}`);

	// Client options
	const clientOptions: LanguageClientOptions = {
		documentSelector: [
			{ scheme: 'file', language: 'smpe' },
			{ scheme: 'zowe-ds', language: 'smpe' },
			{ scheme: 'zowe-uss', language: 'smpe' },
			{ scheme: 'zowe-jobs', language: 'smpe' },
			{ scheme: 'untitled', language: 'smpe' }
		],
		synchronize: {
			fileEvents: vscode.workspace.createFileSystemWatcher('**/*.{smpe,mcs,smp}')
		},
		outputChannel: outputChannel,
		initializationOptions: {
			diagnostics: diagnosticsConfig,
			formatting: formattingConfig
		}
	};

	// Auto-detect SMP/E language for files without extension (e.g. Zowe Explorer dataset members)
	// Register BEFORE starting the client so detection runs before LSP activation
	const detectSmpeLanguage = (doc: vscode.TextDocument) => {
		if (doc.languageId === 'smpe') return;
		const autoDetect = vscode.workspace.getConfiguration('smpe').get<boolean>('editor.autoDetectLanguage', true);
		if (!autoDetect) return;

		if (doc.uri.scheme === 'zowe-ds') {
			const llqRaw = vscode.workspace.getConfiguration('smpe').get('zosDatasetsLlq', ['MCS']);
			const llqList: string[] = Array.isArray(llqRaw)
				? llqRaw.map((q: unknown) => String(q).trim().toUpperCase()).filter(q => q.length > 0)
				: String(llqRaw).split(',').map(q => q.trim().toUpperCase()).filter(q => q.length > 0);
			// Remove member name (MEMBERNAME) anywhere in the path, then extract LLQ (last qualifier)
			const datasetName = doc.uri.path.replace(/\([^)]*\)/g, '').toUpperCase();
			const parts = datasetName.split('.');
			const llq = parts[parts.length - 1].replace(/.*\//, '').trim(); // strip any leading profile/path prefix
			if (llqList.includes(llq)) {
				vscode.languages.setTextDocumentLanguage(doc, 'smpe');
				return;
			}
		}

		// Content-based detection: first non-empty line starts with ++
		for (let i = 0; i < Math.min(10, doc.lineCount); i++) {
			const line = doc.lineAt(i).text.trim();
			if (line.startsWith('++')) {
				vscode.languages.setTextDocumentLanguage(doc, 'smpe');
				break;
			}
			if (line.length > 0) break;
		}
	};
	context.subscriptions.push(vscode.workspace.onDidOpenTextDocument(detectSmpeLanguage));
	// Also handle documents that are already open (e.g. opened before extension activated)
	vscode.workspace.textDocuments.forEach(detectSmpeLanguage);

	// Create the language client
	client = new LanguageClient(
		'smpe-ls',
		'SMP/E Language Server',
		serverOptions,
		clientOptions
	);

	// Start the client (and server)
	client.start().then(() => {
		const info = client.initializeResult?.serverInfo;
		if (info) {
			log(`SMP/E Language Server ready: ${info.name} ${info.version}`);
		} else {
			log('SMP/E Language Server client started successfully');
		}
		// Re-check already open documents after server is ready
		vscode.workspace.textDocuments.forEach(detectSmpeLanguage);
	}).catch((error) => {
		log(`ERROR starting client: ${error}`);
		vscode.window.showErrorMessage(`Failed to start SMP/E Language Server: ${error}`);
	});

	// Register format on save handler
	context.subscriptions.push(
		vscode.workspace.onWillSaveTextDocument(async (e) => {
			if (e.document.languageId !== 'smpe') {
				return;
			}

			const currentConfig = vscode.workspace.getConfiguration('smpe');
			const formatOnSave = currentConfig.get<boolean>('formatting.formatOnSave', false);
			const formattingEnabled = currentConfig.get<boolean>('formatting.enabled', true);

			if (formatOnSave && formattingEnabled) {
				log('Format on save triggered');
				const edits = await vscode.commands.executeCommand<vscode.TextEdit[]>(
					'vscode.executeFormatDocumentProvider',
					e.document.uri
				);
				if (edits && edits.length > 0) {
					const workspaceEdit = new vscode.WorkspaceEdit();
					workspaceEdit.set(e.document.uri, edits);
					e.waitUntil(vscode.workspace.applyEdit(workspaceEdit).then(() => {}));
				}
			}
		})
	);

	// Listen for configuration changes and notify the server
	context.subscriptions.push(
		vscode.workspace.onDidChangeConfiguration(e => {
			if (e.affectsConfiguration('smpe.diagnostics') || e.affectsConfiguration('smpe.formatting')) {
				// Get updated configuration
				const updatedConfig = vscode.workspace.getConfiguration('smpe');
				const updatedDiagnosticsConfig = {
					unknownStatement: updatedConfig.get<boolean>('diagnostics.unknownStatement', true),
					invalidLanguageId: updatedConfig.get<boolean>('diagnostics.invalidLanguageId', true),
					unbalancedParentheses: updatedConfig.get<boolean>('diagnostics.unbalancedParentheses', true),
					missingTerminator: updatedConfig.get<boolean>('diagnostics.missingTerminator', true),
					missingParameter: updatedConfig.get<boolean>('diagnostics.missingParameter', true),
					unknownOperand: updatedConfig.get<boolean>('diagnostics.unknownOperand', true),
					duplicateOperand: updatedConfig.get<boolean>('diagnostics.duplicateOperand', true),
					emptyOperandParameter: updatedConfig.get<boolean>('diagnostics.emptyOperandParameter', true),
					missingRequiredOperand: updatedConfig.get<boolean>('diagnostics.missingRequiredOperand', true),
					dependencyViolation: updatedConfig.get<boolean>('diagnostics.dependencyViolation', true),
					mutuallyExclusive: updatedConfig.get<boolean>('diagnostics.mutuallyExclusive', true),
					requiredGroup: updatedConfig.get<boolean>('diagnostics.requiredGroup', true),
					missingInlineData: updatedConfig.get<boolean>('diagnostics.missingInlineData', true),
					unknownSubOperand: updatedConfig.get<boolean>('diagnostics.unknownSubOperand', true),
					subOperandValidation: updatedConfig.get<boolean>('diagnostics.subOperandValidation', true),
					contentBeyondColumn72: updatedConfig.get<boolean>('diagnostics.contentBeyondColumn72', true),
					standaloneCommentBetweenMCS: updatedConfig.get<boolean>('diagnostics.standaloneCommentBetweenMCS', true)
				};

				const updatedFormattingConfig = {
					enabled: updatedConfig.get<boolean>('formatting.enabled', true),
					indentContinuation: updatedConfig.get<number>('formatting.indentContinuation', 3),
					oneOperandPerLine: updatedConfig.get<boolean>('formatting.oneOperandPerLine', true),
					wrapListsAfterN: updatedConfig.get<number>('formatting.wrapListsAfterN', 2),
					moveLeadingComments: updatedConfig.get<boolean>('formatting.moveLeadingComments', false)
				};

				// Send notification to server
				client.sendNotification('workspace/didChangeConfiguration', {
					settings: {
						smpe: {
							diagnostics: updatedDiagnosticsConfig,
							formatting: updatedFormattingConfig
						}
					}
				});

				log('Sent updated configuration to server');
			}
		})
	);

	// Register USS file content provider (read-only virtual documents)
	const ussContentProvider = new UssFileContentProvider();
	context.subscriptions.push(
		vscode.workspace.registerTextDocumentContentProvider(UssFileContentProvider.scheme, ussContentProvider)
	);
	UssPanel.contentProvider = ussContentProvider;

	// Register MVS dataset content provider (read-only virtual documents)
	const datasetContentProvider = new DatasetContentProvider();
	context.subscriptions.push(
		vscode.workspace.registerTextDocumentContentProvider(DatasetContentProvider.scheme, datasetContentProvider)
	);
	DatasetPanel.contentProvider = datasetContentProvider;

	// Initialize z/OSMF Query Provider
	queryProvider = new QueryProvider(context, outputChannel);
	queryProvider.onResult((result) => {
		const panel = ResultPanel.createOrShow(context.extensionUri);
		const lastServer = queryProvider.getLastServer();
		const lastCreds = queryProvider.getLastCredentials();
		if (lastServer && lastCreds) {
			panel.setZosmfContext(queryProvider.getClient(), lastServer, lastCreds);
		}
		panel.showResult(result);
	});

	// Register z/OSMF commands
	context.subscriptions.push(
		vscode.commands.registerCommand('smpe.zosmf.createConfig', () => {
			queryProvider.createConfig();
		}),
		vscode.commands.registerCommand('smpe.zosmf.clearPassword', () => {
			queryProvider.clearPasswords();
		}),
		vscode.commands.registerCommand('smpe.zosmf.querySysmod', () => {
			queryProvider.querySysmod();
		}),
		vscode.commands.registerCommand('smpe.zosmf.queryDddef', () => {
			queryProvider.queryDddef();
		}),
		vscode.commands.registerCommand('smpe.zosmf.queryZones', () => {
			queryProvider.queryZones();
		}),
		vscode.commands.registerCommand('smpe.zosmf.freeFormQuery', () => {
			FreeFormPanel.createOrShow(queryProvider, outputChannel, context);
		}),
		vscode.commands.registerCommand('smpe.codelens.querySysmod', (...sysmods: string[]) => {
			queryProvider.querySysmodDirect(sysmods);
		}),
		vscode.commands.registerCommand('smpe.codelens.queryDddef', (dddefName: string) => {
			queryProvider.queryDddefDirect([dddefName]);
		}),
        vscode.commands.registerCommand('smpe.zosmf.manageFilterHistory', async () => {
            const HISTORY_KEY = 'smpe.filterHistory';
            const HISTORY_MAX = 20;

            const history = context.globalState.get<string[]>(HISTORY_KEY, []);

            if (history.length === 0) {
                vscode.window.showInformationMessage('SMP/E: Filter history is empty.');
                return;
            }

            const entryItems = history.map((f, i) => ({
                label: f.length > 60 ? f.slice(0, 57) + '...' : f,
                description: f.length > 60 ? f : undefined,
                index: i
            }));
            const clearAllItem = { label: '$(trash) Clear All', index: -1 };
            const separatorItem = { label: '', kind: vscode.QuickPickItemKind.Separator, index: -2 };

            const picked = await vscode.window.showQuickPick(
                [...entryItems, separatorItem as any, clearAllItem as any],
                { placeHolder: 'Select a filter entry to edit or delete' }
            );

            if (!picked) { return; }

            if ((picked as any).index === -1) {
                const confirm = await vscode.window.showWarningMessage(
                    'Clear all saved filter history?', { modal: true }, 'Clear All'
                );
                if (confirm === 'Clear All') {
                    await context.globalState.update(HISTORY_KEY, []);
                    FreeFormPanel.currentPanel?.refreshFilterHistory();
                    vscode.window.showInformationMessage('SMP/E: Filter history cleared.');
                }
                return;
            }

            const idx: number = (picked as any).index;
            const current = history[idx];

            const action = await vscode.window.showQuickPick(
                [
                    { label: '$(edit) Edit', id: 'edit' },
                    { label: '$(trash) Delete', id: 'delete' }
                ],
                { placeHolder: `Action for: ${current}` }
            );

            if (!action) { return; }

            if ((action as any).id === 'delete') {
                history.splice(idx, 1);
                await context.globalState.update(HISTORY_KEY, history);
                FreeFormPanel.currentPanel?.refreshFilterHistory();
                vscode.window.showInformationMessage('SMP/E: Filter entry deleted.');
            } else {
                const updated = await vscode.window.showInputBox({
                    prompt: 'Edit filter',
                    value: current
                });
                if (updated === undefined) { return; }
                const trimmed = updated.trim();
                if (!trimmed) {
                    history.splice(idx, 1);
                } else {
                    const withoutDupe = history.filter((f, i) => i === idx || f !== trimmed);
                    withoutDupe[idx] = trimmed;
                    history.splice(0, history.length, ...withoutDupe.slice(0, HISTORY_MAX));
                }
                await context.globalState.update(HISTORY_KEY, history);
                FreeFormPanel.currentPanel?.refreshFilterHistory();
                vscode.window.showInformationMessage('SMP/E: Filter entry updated.');
            }
        }),
	);

	log('z/OSMF Query commands registered');

	context.subscriptions.push(
		vscode.commands.registerCommand('smpe.editor.toggleAutoDetectLanguage', async () => {
			const config = vscode.workspace.getConfiguration('smpe');
			const current = config.get<boolean>('editor.autoDetectLanguage', true);
			await config.update('editor.autoDetectLanguage', !current, vscode.ConfigurationTarget.Global);
			const state = !current ? 'enabled' : 'disabled';
			log(`Auto-detect language mode ${state}`);
			vscode.window.showInformationMessage(`SMP/E: Auto-detect language mode ${state}.`);
		})
	);

	// Resolve smpe_outl binary path (same priority order as smpe_ls)
	const resolveOutlPath = (): string => {
		const configuredOutlPath = vscode.workspace.getConfiguration('smpe').get<string>('outlPath') || '';

		// Priority 1: configured path
		if (configuredOutlPath && fs.existsSync(configuredOutlPath)) {
			debugLog(`Using configured outlPath: ${configuredOutlPath}`);
			return configuredOutlPath;
		} else if (configuredOutlPath) {
			log(`WARNING: Configured outlPath does not exist: ${configuredOutlPath}`);
		}

		// Priority 2: bundled binary
		const bundledOutlName = process.platform === 'win32' ? 'smpe_outl.exe' : 'smpe_outl';
		const bundledOutlPath = path.join(context.extensionPath, bundledOutlName);
		if (fs.existsSync(bundledOutlPath)) {
			debugLog(`Using bundled smpe_outl: ${bundledOutlPath}`);
			return bundledOutlPath;
		}

		// Priority 3: fallback to ~/.local/bin
		const homeDir = process.env.HOME || process.env.USERPROFILE;
		if (homeDir) {
			const localBinPath = path.join(homeDir, '.local', 'bin', 'smpe_outl');
			if (fs.existsSync(localBinPath)) {
				debugLog(`Using fallback smpe_outl: ${localBinPath}`);
				return localBinPath;
			}
		}

		log('WARNING: smpe_outl binary not found');
		return '';
	};

	// Initialize Missing Input Member Checker
	const missingMemberChecker = new MissingMemberChecker(outputChannel, resolveOutlPath(), dataPath);

	context.subscriptions.push(
		vscode.commands.registerCommand('smpe.checkMissingInputMembers', async () => {
			const outlPath = resolveOutlPath();
			if (!outlPath) {
				vscode.window.showErrorMessage('smpe_outl binary not found. Please install it or configure smpe.outlPath.');
				return;
			}
			missingMemberChecker.setOutlBinaryPath(outlPath);
			await vscode.window.withProgress(
				{ location: vscode.ProgressLocation.Notification, title: 'SMP/E: Checking missing input members...', cancellable: false },
				async () => {
					const results = await missingMemberChecker.check();
					MissingMemberPanel.createOrShow(context.extensionUri, results);
				}
			);
		}),
		// Invalidate per-file cache on save
		vscode.workspace.onDidSaveTextDocument(doc => {
			if (doc.fileName.endsWith('.smpe')) {
				missingMemberChecker.invalidate(doc.fileName);
			}
		}),
		// Invalidate file index when workspace files are created/deleted
		vscode.workspace.createFileSystemWatcher('**/*').onDidCreate(() => missingMemberChecker.invalidateFileIndex()),
		vscode.workspace.createFileSystemWatcher('**/*').onDidDelete(() => missingMemberChecker.invalidateFileIndex()),
	);

	log('Missing Input Member Checker registered');
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}
