import * as fs from 'fs';
import * as vscode from 'vscode';

type FileMap = string | { de: string; en: string };

interface ModuleEntry {
    labelDe: string;
    labelEn: string;
    file: FileMap;
}

const MODULES: ModuleEntry[] = [
    { labelDe: '00 — Index',                labelEn: '00 — Index',                file: '00-index.md' },
    { labelDe: '01 — Installation',          labelEn: '01 — Installation',         file: '01-installation.md' },
    { labelDe: '02 — Erste Schritte',        labelEn: '02 — First Steps',          file: { de: '02-erste-schritte.md', en: '02-first-steps.md' } },
    { labelDe: '03 — Syntax Highlighting',   labelEn: '03 — Syntax Highlighting',  file: '03-syntax-highlighting.md' },
    { labelDe: '04 — Code Completion',       labelEn: '04 — Code Completion',      file: '04-code-completion.md' },
    { labelDe: '05 — Diagnostics',           labelEn: '05 — Diagnostics',          file: '05-diagnostics.md' },
    { labelDe: '06 — Hover und Navigation',  labelEn: '06 — Hover and Navigation', file: { de: '06-hover-und-navigation.md', en: '06-hover-and-navigation.md' } },
    { labelDe: '07 — Formatierung',          labelEn: '07 — Formatting',           file: '07-formatting.md' },
    { labelDe: '08 — z/OSMF Integration',    labelEn: '08 — z/OSMF Integration',   file: '08-zosmf-integration.md' },
    { labelDe: '09 — Bonus: smpe_lint',      labelEn: '09 — Bonus: smpe_lint',     file: '09-bonus-smpe-lint.md' },
    { labelDe: '10 — Bonus: smpe_outl',      labelEn: '10 — Bonus: smpe_outl',     file: '10-bonus-smpe-outl.md' },
];

function resolveLanguage(): 'de' | 'en' {
    return vscode.env.language.toLowerCase().startsWith('de') ? 'de' : 'en';
}

function resolveFilename(entry: ModuleEntry, lang: 'de' | 'en'): string {
    return typeof entry.file === 'string' ? entry.file : entry.file[lang];
}

export async function openTraining(context: vscode.ExtensionContext): Promise<void> {
    const lang = resolveLanguage();

    const items = MODULES.map((entry) => ({
        label: lang === 'de' ? entry.labelDe : entry.labelEn,
        entry,
    }));

    const picked = await vscode.window.showQuickPick(items, {
        placeHolder: lang === 'de' ? 'Trainings-Modul auswählen' : 'Select training module',
        matchOnDescription: false,
    });

    if (!picked) {
        return;
    }

    const filename = resolveFilename(picked.entry, lang);
    const uri = resolveTrainingUri(context, lang, filename);

    if (!uri) {
        vscode.window.showErrorMessage(
            `Training module not found: ${filename}`,
        );
        return;
    }

    try {
        await vscode.commands.executeCommand('markdown.showPreview', uri);
    } catch (err) {
        vscode.window.showErrorMessage(
            `Training module could not be opened: ${filename}`,
        );
    }
}

function resolveTrainingUri(
    context: vscode.ExtensionContext,
    lang: 'de' | 'en',
    filename: string,
): vscode.Uri | undefined {
    const candidates = [
        vscode.Uri.joinPath(context.extensionUri, 'training', lang, filename),
        vscode.Uri.joinPath(context.extensionUri, '..', '..', 'docs', 'training', lang, filename),
    ];

    for (const candidate of candidates) {
        if (fs.existsSync(candidate.fsPath)) {
            return candidate;
        }
    }

    return undefined;
}
