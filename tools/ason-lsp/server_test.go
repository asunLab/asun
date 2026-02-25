package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════════════
// LSP Server Integration Tests
// ═══════════════════════════════════════════════════════════════════════════════

// lspClient is a test helper that drives the LSP server through pipes.
type lspClient struct {
	in  *bytes.Buffer // server reads from here
	out *bytes.Buffer // server writes to here
	srv *Server
}

func newTestClient() *lspClient {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	return &lspClient{
		in:  in,
		out: out,
		srv: NewServer(in, out),
	}
}

func (c *lspClient) send(method string, id *int, params interface{}) {
	body, _ := json.Marshal(params)
	msg := jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  json.RawMessage(body),
	}
	if id != nil {
		raw := json.RawMessage([]byte(fmt.Sprintf("%d", *id)))
		msg.ID = &raw
	}
	data, _ := json.Marshal(msg)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	c.in.WriteString(header)
	c.in.Write(data)
}

func (c *lspClient) readResponse() *jsonRPCMessage {
	raw := c.out.String()
	// Find Content-Length header
	for {
		idx := strings.Index(raw, "Content-Length: ")
		if idx < 0 {
			return nil
		}
		raw = raw[idx:]
		endHeader := strings.Index(raw, "\r\n\r\n")
		if endHeader < 0 {
			return nil
		}
		// parse length
		lenStr := raw[len("Content-Length: "):strings.Index(raw, "\r\n")]
		var clen int
		fmt.Sscanf(lenStr, "%d", &clen)

		bodyStart := endHeader + 4
		if len(raw) < bodyStart+clen {
			return nil
		}
		body := raw[bodyStart : bodyStart+clen]

		var msg jsonRPCMessage
		json.Unmarshal([]byte(body), &msg)
		return &msg
	}
}

func (c *lspClient) readAllResponses() []*jsonRPCMessage {
	raw := c.out.String()
	var msgs []*jsonRPCMessage

	for len(raw) > 0 {
		idx := strings.Index(raw, "Content-Length: ")
		if idx < 0 {
			break
		}
		raw = raw[idx:]
		endHeader := strings.Index(raw, "\r\n\r\n")
		if endHeader < 0 {
			break
		}
		lenStr := raw[len("Content-Length: "):strings.Index(raw, "\r\n")]
		var clen int
		fmt.Sscanf(lenStr, "%d", &clen)
		bodyStart := endHeader + 4
		if len(raw) < bodyStart+clen {
			break
		}
		body := raw[bodyStart : bodyStart+clen]
		var msg jsonRPCMessage
		json.Unmarshal([]byte(body), &msg)
		msgs = append(msgs, &msg)
		raw = raw[bodyStart+clen:]
	}
	return msgs
}

func intPtr(v int) *int { return &v }

// initAndOpen initializes the server and opens a document.
func (c *lspClient) initAndOpen(uri, text string) {
	c.send("initialize", intPtr(1), InitializeParams{ProcessID: 1})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Text: text},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()
}

// ──────────── Initialize ────────────

func TestLSPInitialize(t *testing.T) {
	c := newTestClient()

	c.send("initialize", intPtr(1), InitializeParams{ProcessID: 1234})

	msg, err := c.srv.readMessage()
	if err != nil {
		t.Fatalf("readMessage: %v", err)
	}
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("no response")
	}
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}

	var result InitializeResult
	json.Unmarshal(resp.Result, &result)
	if result.Capabilities.TextDocumentSync != 1 {
		t.Errorf("TextDocumentSync = %d, want 1", result.Capabilities.TextDocumentSync)
	}
	if !result.Capabilities.HoverProvider {
		t.Error("HoverProvider should be true")
	}
	if !result.Capabilities.DocumentFormattingProvider {
		t.Error("DocumentFormattingProvider should be true")
	}
	if result.Capabilities.CompletionProvider == nil {
		t.Error("CompletionProvider should be set")
	}
	if result.Capabilities.SemanticTokensProvider == nil {
		t.Error("SemanticTokensProvider should be set")
	}
	if result.ServerInfo == nil || result.ServerInfo.Name != "ason-lsp" {
		t.Error("ServerInfo name should be ason-lsp")
	}
}

// ──────────── didOpen + diagnostics ────────────

func TestLSPDidOpenValidDoc(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.ason",
			LanguageID: "ason",
			Version:    1,
			Text:       `{name:str,age:int}:(Alice,30)`,
		},
	})

	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)

	msgs := c.readAllResponses()
	// Should get publishDiagnostics notification with 0 diagnostics
	found := false
	for _, m := range msgs {
		if m.Method == "textDocument/publishDiagnostics" {
			var params PublishDiagnosticsParams
			json.Unmarshal(m.Params, &params)
			if len(params.Diagnostics) != 0 {
				t.Errorf("expected 0 diagnostics, got %d", len(params.Diagnostics))
			}
			found = true
		}
	}
	if !found {
		t.Error("no publishDiagnostics notification sent")
	}
}

func TestLSPDidOpenInvalidDoc(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.ason",
			LanguageID: "ason",
			Version:    1,
			Text:       `{a:int,b:int}:(1,2,3)`,
		},
	})

	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)

	msgs := c.readAllResponses()
	found := false
	for _, m := range msgs {
		if m.Method == "textDocument/publishDiagnostics" {
			var params PublishDiagnosticsParams
			json.Unmarshal(m.Params, &params)
			if len(params.Diagnostics) == 0 {
				t.Error("expected at least 1 diagnostic for field mismatch")
			}
			found = true
		}
	}
	if !found {
		t.Error("no publishDiagnostics notification sent")
	}
}

// ──────────── didChange ────────────

func TestLSPDidChange(t *testing.T) {
	c := newTestClient()

	// Open
	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.ason",
			Text: `{a:int}:(1)`,
		},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	// Change to invalid
	c.send("textDocument/didChange", nil, DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{URI: "file:///test.ason", Version: 2},
		ContentChanges: []TextDocumentContentChangeEvent{
			{Text: `{a:int,b:int}:(1,2,3)`},
		},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	msgs := c.readAllResponses()
	found := false
	for _, m := range msgs {
		if m.Method == "textDocument/publishDiagnostics" {
			var params PublishDiagnosticsParams
			json.Unmarshal(m.Params, &params)
			if len(params.Diagnostics) == 0 {
				t.Error("expected diagnostic after change")
			}
			found = true
		}
	}
	if !found {
		t.Error("no diagnostics after change")
	}
}

// ──────────── Hover ────────────

func TestLSPHover(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.ason",
			Text: `{name:str,age:int}:(Alice,30)`,
		},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	// Hover over 'name' (col 1)
	c.send("textDocument/hover", intPtr(2), TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.ason"},
		Position:     LSPPosition{Line: 0, Character: 1},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("no hover response")
	}
	if resp.Result == nil || string(resp.Result) == "null" {
		t.Fatal("hover returned null for field position")
	}
	var hover LSPHover
	json.Unmarshal(resp.Result, &hover)
	if !strings.Contains(hover.Contents.Value, "name") {
		t.Errorf("hover should mention 'name', got: %q", hover.Contents.Value)
	}
}

// ──────────── Completion ────────────

func TestLSPCompletion(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.ason",
			Text: ``,
		},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	c.send("textDocument/completion", intPtr(3), TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.ason"},
		Position:     LSPPosition{Line: 0, Character: 0},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("no completion response")
	}
	var list LSPCompletionList
	json.Unmarshal(resp.Result, &list)
	if len(list.Items) == 0 {
		t.Error("expected completions for empty doc")
	}
}

// ──────────── Formatting ────────────

func TestLSPFormatting(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.ason",
			Text: `{name:str,  age:int}:(Alice,  30)`,
		},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	c.send("textDocument/formatting", intPtr(4), DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.ason"},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("no formatting response")
	}
	var edits []LSPTextEdit
	json.Unmarshal(resp.Result, &edits)
	if len(edits) == 0 {
		t.Error("expected formatting edits for document with extra spaces")
	}
	if len(edits) > 0 && !strings.Contains(edits[0].NewText, "{name:str, age:int}") {
		t.Errorf("formatted result: %q", edits[0].NewText)
	}
}

// ──────────── Semantic Tokens ────────────

func TestLSPSemanticTokens(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.ason",
			Text: `{name:str}:(Alice)`,
		},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	c.send("textDocument/semanticTokens/full", intPtr(5), SemanticTokensParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.ason"},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("no semantic tokens response")
	}
	var result SemanticTokensResult
	json.Unmarshal(resp.Result, &result)
	if len(result.Data) == 0 {
		t.Error("expected semantic token data")
	}
	// Data is encoded as groups of 5: deltaLine, deltaStartChar, length, tokenType, tokenModifiers
	if len(result.Data)%5 != 0 {
		t.Errorf("semantic token data length %d is not multiple of 5", len(result.Data))
	}
}

// ──────────── didClose ────────────

func TestLSPDidClose(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.ason",
			Text: `{a:int}:(1)`,
		},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	c.send("textDocument/didClose", nil, DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.ason"},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	// Should send empty diagnostics
	msgs := c.readAllResponses()
	found := false
	for _, m := range msgs {
		if m.Method == "textDocument/publishDiagnostics" {
			var params PublishDiagnosticsParams
			json.Unmarshal(m.Params, &params)
			if len(params.Diagnostics) != 0 {
				t.Error("expected empty diagnostics on close")
			}
			found = true
		}
	}
	if !found {
		t.Error("no clear-diagnostics notification on close")
	}

	// Verify doc is removed
	if _, ok := c.srv.docs["file:///test.ason"]; ok {
		t.Error("document should be removed from docs map")
	}
}

// ──────────── Inlay Hints ────────────

func TestLSPInlayHint(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.ason",
			Text: `{name:str,age:int}:(Alice,30)`,
		},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	c.send("textDocument/inlayHint", intPtr(10), InlayHintParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.ason"},
		Range: LSPRange{
			Start: LSPPosition{Line: 0, Character: 0},
			End:   LSPPosition{Line: 0, Character: 100},
		},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("no inlay hint response")
	}
	var hints []LSPInlayHint
	json.Unmarshal(resp.Result, &hints)
	if len(hints) != 2 {
		t.Fatalf("expected 2 inlay hints, got %d", len(hints))
	}
	if hints[0].Label != "name:" {
		t.Errorf("hint[0] = %q, want 'name:'", hints[0].Label)
	}
	if hints[1].Label != "age:" {
		t.Errorf("hint[1] = %q, want 'age:'", hints[1].Label)
	}
}

func TestLSPInlayHintObjectArray(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.ason",
			Text: `[{id:int,name:str}]:(1,Alice),(2,Bob)`,
		},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	c.send("textDocument/inlayHint", intPtr(11), InlayHintParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.ason"},
		Range: LSPRange{
			Start: LSPPosition{Line: 0, Character: 0},
			End:   LSPPosition{Line: 0, Character: 200},
		},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("no inlay hint response")
	}
	var hints []LSPInlayHint
	json.Unmarshal(resp.Result, &hints)
	if len(hints) != 4 {
		t.Fatalf("expected 4 inlay hints, got %d", len(hints))
	}
}

// ──────────── Execute Command (Compress) ────────────

func TestLSPExecuteCommandCompress(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.ason",
			Text: "{name:str, age:int}:\n  (Alice, 30)",
		},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	uriArg, _ := json.Marshal("file:///test.ason")
	c.send("workspace/executeCommand", intPtr(20), ExecuteCommandParams{
		Command:   "ason.compress",
		Arguments: []json.RawMessage{json.RawMessage(uriArg)},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("expected response for compress command")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
	var compressed string
	json.Unmarshal(resp.Result, &compressed)
	expected := "{name:str,age:int}:(Alice,30)"
	if compressed != expected {
		t.Errorf("compress result = %q, want %q", compressed, expected)
	}
}

// ──────────── Unknown Method: executeCommand errors ────────────

func TestLSPExecuteCommandUnknown(t *testing.T) {
	c := newTestClient()

	c.send("workspace/executeCommand", intPtr(30), ExecuteCommandParams{
		Command: "ason.unknownCommand",
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("expected response for unknown command")
	}
	if resp.Error == nil {
		t.Error("expected error for unknown command")
	}
}

// ──────────── Unknown Method ────────────

func TestLSPUnknownMethod(t *testing.T) {
	c := newTestClient()

	c.send("textDocument/unknown", intPtr(99), nil)
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("expected error response for unknown method")
	}
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("error code = %d, want -32601", resp.Error.Code)
	}
}

// ──────────── Shutdown ────────────

func TestLSPShutdown(t *testing.T) {
	c := newTestClient()

	c.send("shutdown", intPtr(100), nil)
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)

	if !c.srv.shutdown {
		t.Error("server should be in shutdown state")
	}
}

// ──────────── End-to-end flow ────────────

func TestLSPEndToEndFlow(t *testing.T) {
	c := newTestClient()

	// 1. Initialize
	c.send("initialize", intPtr(1), InitializeParams{ProcessID: 1})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	// 2. Initialized
	c.send("initialized", nil, struct{}{})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	// 3. Open document
	c.send("textDocument/didOpen", nil, DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///project/test.ason",
			Text: "[{id:int,name:str}]:\n  (1,Alice),\n  (2,Bob)",
		},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	msgs := c.readAllResponses()
	diagOK := false
	for _, m := range msgs {
		if m.Method == "textDocument/publishDiagnostics" {
			var params PublishDiagnosticsParams
			json.Unmarshal(m.Params, &params)
			if len(params.Diagnostics) == 0 {
				diagOK = true
			}
		}
	}
	if !diagOK {
		t.Error("expected clean diagnostics for valid doc")
	}
	c.out.Reset()

	// 4. Hover
	c.send("textDocument/hover", intPtr(2), TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///project/test.ason"},
		Position:     LSPPosition{Line: 0, Character: 2},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)
	c.out.Reset()

	// 5. Semantic tokens
	c.send("textDocument/semanticTokens/full", intPtr(3), SemanticTokensParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///project/test.ason"},
	})
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	resp := c.readResponse()
	if resp == nil {
		t.Fatal("no semantic tokens response in e2e")
	}
	c.out.Reset()

	// 6. Shutdown
	c.send("shutdown", intPtr(4), nil)
	msg, _ = c.srv.readMessage()
	c.srv.handleMessage(msg)

	if !c.srv.shutdown {
		t.Error("server should shut down")
	}
}

// ──────────── ASON ↔ JSON commands ────────────

func TestLSPToJSON(t *testing.T) {
	c := newTestClient()
	c.initAndOpen("file:///conv.ason", `{name:str,age:int}:(Alice,30)`)

	c.send("workspace/executeCommand", intPtr(10), ExecuteCommandParams{
		Command:   "ason.toJSON",
		Arguments: []json.RawMessage{json.RawMessage(`"file:///conv.ason"`)},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	resp := c.readResponse()
	if resp == nil {
		t.Fatal("no response for toJSON")
	}
	var result string
	json.Unmarshal(resp.Result, &result)
	if !strings.Contains(result, "Alice") || !strings.Contains(result, "30") {
		t.Errorf("toJSON result = %q, missing expected values", result)
	}
	// Validate it's proper JSON
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Errorf("toJSON result is not valid JSON: %v", err)
	}
}

func TestLSPFromJSON(t *testing.T) {
	c := newTestClient()
	c.initAndOpen("file:///dummy.ason", `{}:()`)

	jsonSrc := `{"age": 30, "name": "Alice"}`
	c.send("workspace/executeCommand", intPtr(11), ExecuteCommandParams{
		Command:   "ason.fromJSON",
		Arguments: []json.RawMessage{json.RawMessage(`"` + strings.ReplaceAll(jsonSrc, `"`, `\"`) + `"`)},
	})
	msg, _ := c.srv.readMessage()
	c.srv.handleMessage(msg)
	resp := c.readResponse()
	if resp == nil {
		t.Fatal("no response for fromJSON")
	}
	var result string
	json.Unmarshal(resp.Result, &result)
	if !strings.Contains(result, "age") || !strings.Contains(result, "name") {
		t.Errorf("fromJSON result = %q, missing schema fields", result)
	}
	if !strings.Contains(result, "30") || !strings.Contains(result, "Alice") {
		t.Errorf("fromJSON result = %q, missing data values", result)
	}
}
