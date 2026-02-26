import * as path from "path";
import * as fs from "fs";
import {
  workspace,
  ExtensionContext,
  commands,
  window,
  TextEditor,
  TextEdit,
  Range,
  Position,
} from "vscode";

import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind,
  ExecuteCommandRequest,
} from "vscode-languageclient/node";

let client: LanguageClient;

export function activate(context: ExtensionContext) {
  // Register commands FIRST so they are always available,
  // even if the LSP binary is not found yet.
  context.subscriptions.push(
    commands.registerCommand("ason.format", () => formatDocument()),
    commands.registerCommand("ason.compress", () => compressDocument()),
    commands.registerCommand("ason.toJSON", () => asonToJSON()),
    commands.registerCommand("ason.fromJSON", () => jsonToASON()),
  );

  const serverPath = resolveServerPath(context);
  if (!serverPath) {
    window.showErrorMessage(
      "ASON LSP binary not found. Please set ason.lspPath in settings or ensure ason-lsp is in PATH.",
    );
    return;
  }

  const serverOptions: ServerOptions = {
    run: {
      command: serverPath,
      args: ["-stdio"],
      transport: TransportKind.stdio,
    },
    debug: {
      command: serverPath,
      args: ["-stdio"],
      transport: TransportKind.stdio,
    },
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: "file", language: "ason" }],
    synchronize: {
      fileEvents: workspace.createFileSystemWatcher("**/*.ason"),
    },
  };

  client = new LanguageClient(
    "ason-lsp",
    "ASON Language Server",
    serverOptions,
    clientOptions,
  );

  client.start();
}

export function deactivate(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}

/**
 * Resolve the ason-lsp binary path.
 * Priority: settings > bundled (server/ inside extension) > PATH
 */
function resolveServerPath(context: ExtensionContext): string | undefined {
  // 1. Check user setting
  const config = workspace.getConfiguration("ason");
  const configPath = config.get<string>("lspPath");
  if (configPath && configPath.trim() !== "") {
    if (fs.existsSync(configPath)) {
      return configPath;
    }
    window.showWarningMessage(
      `ason.lspPath "${configPath}" not found, trying defaults.`,
    );
  }

  // 2. Check bundled binary inside extension (like rust-analyzer)
  const isWindows = process.platform === "win32";
  const binaryName = isWindows ? "ason-lsp.exe" : "ason-lsp";
  const bundledPaths = [
    // Primary: server/ directory inside extension
    path.resolve(context.extensionPath, "server", binaryName),
    // Fallback: sibling ason-lsp directory (development mode)
    path.resolve(context.extensionPath, "..", "ason-lsp", binaryName),
  ];
  for (const bp of bundledPaths) {
    if (fs.existsSync(bp)) {
      // Ensure executable permission on Unix
      if (!isWindows) {
        try {
          fs.chmodSync(bp, 0o755);
        } catch {
          // ignore
        }
      }
      return bp;
    }
  }

  // 3. Check PATH
  const { execSync } = require("child_process");
  try {
    const cmd = isWindows ? "where ason-lsp" : "which ason-lsp";
    const which = execSync(cmd, { encoding: "utf8" }).trim();
    if (which && fs.existsSync(which)) {
      return which;
    }
  } catch {
    // not in PATH
  }

  return undefined;
}

/**
 * Format (beautify) the active ASON document.
 */
async function formatDocument() {
  const editor = window.activeTextEditor;
  if (!editor || editor.document.languageId !== "ason") {
    window.showInformationMessage("Open an ASON file first.");
    return;
  }
  await commands.executeCommand("editor.action.formatDocument");
}

/**
 * Compress (minify) the active ASON document via LSP command.
 */
async function compressDocument() {
  const editor = window.activeTextEditor;
  if (!editor || editor.document.languageId !== "ason") {
    window.showInformationMessage("Open an ASON file first.");
    return;
  }

  if (!client) {
    window.showErrorMessage("ASON language server not running.");
    return;
  }

  try {
    const compressed: string = await client.sendRequest(
      ExecuteCommandRequest.type,
      {
        command: "ason.compress",
        arguments: [editor.document.uri.toString()],
      },
    );
    if (compressed) {
      const fullRange = new Range(
        new Position(0, 0),
        editor.document.lineAt(editor.document.lineCount - 1).range.end,
      );
      await editor.edit((editBuilder) => {
        editBuilder.replace(fullRange, compressed);
      });
    }
  } catch (err: any) {
    window.showErrorMessage(`Compress failed: ${err.message || err}`);
  }
}

/**
 * Convert the active ASON document to JSON and open in a new editor.
 */
async function asonToJSON() {
  const editor = window.activeTextEditor;
  if (!editor || editor.document.languageId !== "ason") {
    window.showInformationMessage("Open an ASON file first.");
    return;
  }

  if (!client) {
    window.showErrorMessage("ASON language server not running.");
    return;
  }

  try {
    const result: string = await client.sendRequest(
      ExecuteCommandRequest.type,
      {
        command: "ason.toJSON",
        arguments: [editor.document.uri.toString()],
      },
    );
    if (result) {
      const doc = await workspace.openTextDocument({
        content: result,
        language: "json",
      });
      await window.showTextDocument(doc, { preview: false });
    }
  } catch (err: any) {
    window.showErrorMessage(`Convert to JSON failed: ${err.message || err}`);
  }
}

/**
 * Convert JSON content to ASON. Works on active JSON file or prompts for input.
 */
async function jsonToASON() {
  const editor = window.activeTextEditor;
  let jsonText = "";

  if (editor && editor.document.languageId === "json") {
    jsonText = editor.document.getText();
  } else {
    const input = await window.showInputBox({
      prompt: "Paste JSON content to convert to ASON",
      placeHolder: '{"key": "value"}',
    });
    if (!input) {
      return;
    }
    jsonText = input;
  }

  if (!client) {
    window.showErrorMessage("ASON language server not running.");
    return;
  }

  try {
    const result: string = await client.sendRequest(
      ExecuteCommandRequest.type,
      {
        command: "ason.fromJSON",
        arguments: [jsonText],
      },
    );
    if (result) {
      const doc = await workspace.openTextDocument({
        content: result,
        language: "ason",
      });
      await window.showTextDocument(doc, { preview: false });
    }
  } catch (err: any) {
    window.showErrorMessage(`Convert to ASON failed: ${err.message || err}`);
  }
}
