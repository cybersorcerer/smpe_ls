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
	outputChannel.show(true);

	log('SMP/E Language Server extension activating...');
	log(`Platform: ${process.platform}`);
	log(`Extension path: ${context.extensionPath}`);

	// Get configuration
	const config = vscode.workspace.getConfiguration('smpe');
	const configuredServerPath = config.get<string>('serverPath') || '';
	const debug = config.get<boolean>('debug') || true;

	log(`Configured serverPath: "${configuredServerPath}"`);
	log(`Debug mode: ${debug}`);

	// Determine the full path to the server
	let executable = '';
	let dataPathArgs: string[] = [];

	// Check if user configured a custom path
	if (configuredServerPath && fs.existsSync(configuredServerPath)) {
		log(`Using configured server path: ${configuredServerPath}`);
		executable = configuredServerPath;
	} else {
		// Check if we have a bundled binary in the extension folder
		const bundledBinaryName = process.platform === 'win32' ? 'smpe_ls.exe' : 'smpe_ls';
		const bundledBinaryPath = path.join(context.extensionPath, bundledBinaryName);
		const bundledDataPath = path.join(context.extensionPath, 'smpe.json');

		log(`Looking for bundled binary at: ${bundledBinaryPath}`);

		if (fs.existsSync(bundledBinaryPath)) {
			log(`Found bundled binary`);
			executable = bundledBinaryPath;

			// Check for bundled data too
			log(`Looking for bundled data at: ${bundledDataPath}`);
			if (fs.existsSync(bundledDataPath)) {
				log(`Found bundled data`);
				dataPathArgs = ['--data', bundledDataPath];
			} else {
				log(`WARNING: Bundled data file not found`);
			}
		} else {
			log(`Bundled binary NOT found`);

			// Try ~/.local/bin as fallback (Linux/macOS)
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
	}

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

	// Client options
	const clientOptions: LanguageClientOptions = {
		documentSelector: [
			{ scheme: 'file', language: 'smpe' }
		],
		synchronize: {
			fileEvents: vscode.workspace.createFileSystemWatcher('**/*.{smpe,mcs,smp}')
		},
		outputChannel: outputChannel
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
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}
