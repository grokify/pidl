import * as vscode from 'vscode';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

let diagnosticCollection: vscode.DiagnosticCollection;

export function activate(context: vscode.ExtensionContext) {
    console.log('PIDL extension activated');

    // Create diagnostic collection
    diagnosticCollection = vscode.languages.createDiagnosticCollection('pidl');
    context.subscriptions.push(diagnosticCollection);

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('pidl.validate', validateCurrentFile),
        vscode.commands.registerCommand('pidl.preview', previewDiagram),
        vscode.commands.registerCommand('pidl.exportMermaid', exportMermaid),
        vscode.commands.registerCommand('pidl.exportSvg', exportSvg),
        vscode.commands.registerCommand('pidl.runSecurity', runSecurityAnalysis)
    );

    // Auto-validate on save
    const config = vscode.workspace.getConfiguration('pidl');
    if (config.get('validateOnSave')) {
        context.subscriptions.push(
            vscode.workspace.onDidSaveTextDocument((document) => {
                if (document.languageId === 'pidl') {
                    validateDocument(document);
                }
            })
        );
    }

    // Validate on open
    context.subscriptions.push(
        vscode.workspace.onDidOpenTextDocument((document) => {
            if (document.languageId === 'pidl') {
                validateDocument(document);
            }
        })
    );
}

export function deactivate() {
    if (diagnosticCollection) {
        diagnosticCollection.dispose();
    }
}

async function validateCurrentFile() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
    }

    if (editor.document.languageId !== 'pidl') {
        vscode.window.showErrorMessage('Current file is not a PIDL file');
        return;
    }

    await validateDocument(editor.document);
}

async function validateDocument(document: vscode.TextDocument) {
    const config = vscode.workspace.getConfiguration('pidl');
    const cliPath = config.get<string>('cliPath', 'pidl');
    const filePath = document.uri.fsPath;

    try {
        await execAsync(`${cliPath} validate "${filePath}"`);
        diagnosticCollection.set(document.uri, []);
        vscode.window.showInformationMessage('PIDL validation passed');
    } catch (error: unknown) {
        const errorMessage = error instanceof Error ? error.message : String(error);
        const diagnostics = parseValidationErrors(document, errorMessage);
        diagnosticCollection.set(document.uri, diagnostics);
        vscode.window.showErrorMessage('PIDL validation failed');
    }
}

function parseValidationErrors(document: vscode.TextDocument, errorOutput: string): vscode.Diagnostic[] {
    const diagnostics: vscode.Diagnostic[] = [];

    // Parse error output - format varies but commonly includes line numbers
    const lines = errorOutput.split('\n');
    for (const line of lines) {
        if (line.includes('error') || line.includes('Error')) {
            // Try to extract line number
            const lineMatch = line.match(/line (\d+)/i);
            const lineNumber = lineMatch ? parseInt(lineMatch[1], 10) - 1 : 0;

            const range = new vscode.Range(
                new vscode.Position(lineNumber, 0),
                new vscode.Position(lineNumber, document.lineAt(Math.min(lineNumber, document.lineCount - 1)).text.length)
            );

            diagnostics.push(new vscode.Diagnostic(
                range,
                line.trim(),
                vscode.DiagnosticSeverity.Error
            ));
        }
    }

    // If no specific errors found, add a general error
    if (diagnostics.length === 0 && errorOutput.trim()) {
        diagnostics.push(new vscode.Diagnostic(
            new vscode.Range(0, 0, 0, 0),
            errorOutput.trim(),
            vscode.DiagnosticSeverity.Error
        ));
    }

    return diagnostics;
}

async function previewDiagram() {
    const editor = vscode.window.activeTextEditor;
    if (!editor || editor.document.languageId !== 'pidl') {
        vscode.window.showErrorMessage('No PIDL file open');
        return;
    }

    const config = vscode.workspace.getConfiguration('pidl');
    const cliPath = config.get<string>('cliPath', 'pidl');
    const filePath = editor.document.uri.fsPath;

    try {
        const { stdout } = await execAsync(`${cliPath} render -f mermaid "${filePath}"`);

        // Create webview panel with Mermaid diagram
        const panel = vscode.window.createWebviewPanel(
            'pidlPreview',
            `PIDL Preview: ${editor.document.fileName}`,
            vscode.ViewColumn.Beside,
            { enableScripts: true }
        );

        panel.webview.html = getMermaidHtml(stdout);
    } catch (error: unknown) {
        const errorMessage = error instanceof Error ? error.message : String(error);
        vscode.window.showErrorMessage(`Failed to generate preview: ${errorMessage}`);
    }
}

function getMermaidHtml(mermaidCode: string): string {
    return `<!DOCTYPE html>
<html>
<head>
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <style>
        body {
            padding: 20px;
            background: var(--vscode-editor-background);
            color: var(--vscode-editor-foreground);
        }
        .mermaid { text-align: center; }
    </style>
</head>
<body>
    <div class="mermaid">
${mermaidCode}
    </div>
    <script>
        mermaid.initialize({ startOnLoad: true, theme: 'neutral' });
    </script>
</body>
</html>`;
}

async function exportMermaid() {
    const editor = vscode.window.activeTextEditor;
    if (!editor || editor.document.languageId !== 'pidl') {
        vscode.window.showErrorMessage('No PIDL file open');
        return;
    }

    const config = vscode.workspace.getConfiguration('pidl');
    const cliPath = config.get<string>('cliPath', 'pidl');
    const filePath = editor.document.uri.fsPath;

    try {
        const { stdout } = await execAsync(`${cliPath} render -f mermaid "${filePath}"`);

        // Create new document with mermaid content
        const doc = await vscode.workspace.openTextDocument({
            content: stdout,
            language: 'mermaid'
        });
        await vscode.window.showTextDocument(doc);
    } catch (error: unknown) {
        const errorMessage = error instanceof Error ? error.message : String(error);
        vscode.window.showErrorMessage(`Failed to export Mermaid: ${errorMessage}`);
    }
}

async function exportSvg() {
    const editor = vscode.window.activeTextEditor;
    if (!editor || editor.document.languageId !== 'pidl') {
        vscode.window.showErrorMessage('No PIDL file open');
        return;
    }

    const config = vscode.workspace.getConfiguration('pidl');
    const cliPath = config.get<string>('cliPath', 'pidl');
    const filePath = editor.document.uri.fsPath;

    try {
        const { stdout } = await execAsync(`${cliPath} render -f svg "${filePath}"`);

        // Save SVG file
        const saveUri = await vscode.window.showSaveDialog({
            defaultUri: vscode.Uri.file(filePath.replace('.pidl.json', '.svg')),
            filters: { 'SVG': ['svg'] }
        });

        if (saveUri) {
            await vscode.workspace.fs.writeFile(saveUri, Buffer.from(stdout, 'utf8'));
            vscode.window.showInformationMessage(`SVG saved to ${saveUri.fsPath}`);
        }
    } catch (error: unknown) {
        const errorMessage = error instanceof Error ? error.message : String(error);
        vscode.window.showErrorMessage(`Failed to export SVG: ${errorMessage}`);
    }
}

async function runSecurityAnalysis() {
    const editor = vscode.window.activeTextEditor;
    if (!editor || editor.document.languageId !== 'pidl') {
        vscode.window.showErrorMessage('No PIDL file open');
        return;
    }

    const config = vscode.workspace.getConfiguration('pidl');
    const cliPath = config.get<string>('cliPath', 'pidl');
    const filePath = editor.document.uri.fsPath;

    try {
        const { stdout } = await execAsync(`${cliPath} security "${filePath}"`);

        // Show security report in output channel
        const outputChannel = vscode.window.createOutputChannel('PIDL Security');
        outputChannel.clear();
        outputChannel.appendLine('Security Analysis Report');
        outputChannel.appendLine('========================\n');
        outputChannel.appendLine(stdout);
        outputChannel.show();
    } catch (error: unknown) {
        const errorMessage = error instanceof Error ? error.message : String(error);
        vscode.window.showErrorMessage(`Security analysis failed: ${errorMessage}`);
    }
}
