import * as path from 'path';
import * as vscode from 'vscode';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    Executable
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
    console.log('SMP/E Language Server extension is now active');

    // Get configuration
    const config = vscode.workspace.getConfiguration('smpe');
    const serverPath = config.get<string>('serverPath') || 'smpe_ls';

    // FORCE debug mode to true during development
    const debug = true;

    console.log(`DEBUG: debug config value = ${config.get<boolean>('debug')}, FORCED final debug = ${debug}`);

    // Determine the full path to the server
    let executable = serverPath;

    // If it's not an absolute path, try to find it
    if (!path.isAbsolute(serverPath)) {
        // First try ~/.local/bin
        const homeDir = process.env.HOME || process.env.USERPROFILE;
        if (homeDir) {
            const localBinPath = path.join(homeDir, '.local', 'bin', serverPath);
            executable = localBinPath;
        }
    }

    // Server options
    // Server will use default data path: ~/.local/share/smpe_ls/smpe.json
    const args = debug ? ['--debug'] : [];

    const serverExecutable: Executable = {
        command: executable,
        args: args,
        options: {
            env: process.env
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
    client.start();

    console.log('SMP/E Language Server client started');
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}
