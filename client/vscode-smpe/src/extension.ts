import * as path from 'path';
import * as fs from 'fs';
import * as vscode from 'vscode';
import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
	Executable
} from 'vscode-languageclient/node';

let client: LanguageClient;
let outputChannel: vscode.OutputChannel;

function log(message: string) {
	const timestamp = new Date().toISOString();
	outputChannel.appendLine(`[${timestamp}] ${message}`);
	console.log(message);
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

	log(`Configured serverPath: "${configuredServerPath}"`);
	log(`Configured dataPath: "${configuredDataPath}"`);
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
		log(`Checking configured server path: ${configuredServerPath}`);
		if (fs.existsSync(configuredServerPath)) {
			log(`Using configured server path`);
			executable = configuredServerPath;
		} else {
			log(`WARNING: Configured serverPath does not exist`);
		}
	}

	// Priority 2: Bundled binary
	if (!executable) {
		log(`Looking for bundled binary at: ${bundledBinaryPath}`);
		if (fs.existsSync(bundledBinaryPath)) {
			log(`Found bundled binary`);
			executable = bundledBinaryPath;
		} else {
			log(`Bundled binary NOT found`);
		}
	}

	// Priority 3: Fallback to ~/.local/bin
	if (!executable) {
		const homeDir = process.env.HOME || process.env.USERPROFILE;
		if (homeDir) {
			const localBinPath = path.join(homeDir, '.local', 'bin', 'smpe_ls');
			log(`Trying fallback path: ${localBinPath}`);
			if (fs.existsSync(localBinPath)) {
				log(`Found server at fallback path`);
				executable = localBinPath;
			} else {
				log(`Server NOT found at fallback path`);
			}
		}
	}

	// Determine data path
	// Priority 1: User configured data path
	if (configuredDataPath) {
		log(`Checking configured data path: ${configuredDataPath}`);
		if (fs.existsSync(configuredDataPath)) {
			log(`Using configured data path`);
			dataPath = configuredDataPath;
		} else {
			log(`WARNING: Configured dataPath does not exist`);
		}
	}

	// Priority 2: Bundled data file
	if (!dataPath) {
		log(`Looking for bundled data at: ${bundledDataPath}`);
		if (fs.existsSync(bundledDataPath)) {
			log(`Found bundled data`);
			dataPath = bundledDataPath;
		} else {
			log(`Bundled data NOT found`);
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
		moveLeadingComments: config.get<boolean>('formatting.moveLeadingComments', false)
	};

	log(`Diagnostics config: ${JSON.stringify(diagnosticsConfig)}`);
	log(`Formatting config: ${JSON.stringify(formattingConfig)}`);

	// Client options
	const clientOptions: LanguageClientOptions = {
		documentSelector: [
			{ scheme: 'file', language: 'smpe' }
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

	// Create the language client
	client = new LanguageClient(
		'smpe-ls',
		'SMP/E Language Server',
		serverOptions,
		clientOptions
	);

	// Start the client (and server)
	client.start().then(() => {
		log('SMP/E Language Server client started successfully');
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
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}
