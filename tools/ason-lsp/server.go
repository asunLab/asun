package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

// ──────────────────────────────────────────────────────────────────────────────
// JSON-RPC types
// ──────────────────────────────────────────────────────────────────────────────

type jsonRPCMessage struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *jsonRPCError    `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ──────────────────────────────────────────────────────────────────────────────
// LSP types (minimal subset)
// ──────────────────────────────────────────────────────────────────────────────

type InitializeParams struct {
	ProcessID int `json:"processId"`
}

type ServerCapabilities struct {
	TextDocumentSync           int                    `json:"textDocumentSync"` // 1=Full
	CompletionProvider         *CompletionOptions     `json:"completionProvider,omitempty"`
	HoverProvider              bool                   `json:"hoverProvider"`
	DocumentFormattingProvider bool                   `json:"documentFormattingProvider"`
	DiagnosticProvider         *DiagnosticOptions     `json:"diagnosticProvider,omitempty"`
	SemanticTokensProvider     *SemanticTokensOptions `json:"semanticTokensProvider,omitempty"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type DiagnosticOptions struct {
	InterFileDependencies bool `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool `json:"workspaceDiagnostics"`
}

type SemanticTokensOptions struct {
	Full   bool           `json:"full"`
	Legend SemanticLegend `json:"legend"`
}

type SemanticLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     LSPPosition            `json:"position"`
}

type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type LSPPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type LSPRange struct {
	Start LSPPosition `json:"start"`
	End   LSPPosition `json:"end"`
}

type LSPDiagnostic struct {
	Range    LSPRange `json:"range"`
	Severity int      `json:"severity"`
	Source   string   `json:"source"`
	Message  string   `json:"message"`
}

type PublishDiagnosticsParams struct {
	URI         string          `json:"uri"`
	Diagnostics []LSPDiagnostic `json:"diagnostics"`
}

type LSPHover struct {
	Contents LSPMarkupContent `json:"contents"`
	Range    *LSPRange        `json:"range,omitempty"`
}

type LSPMarkupContent struct {
	Kind  string `json:"kind"` // "markdown" or "plaintext"
	Value string `json:"value"`
}

type LSPCompletionItem struct {
	Label      string `json:"label"`
	Kind       int    `json:"kind"`
	Detail     string `json:"detail,omitempty"`
	InsertText string `json:"insertText,omitempty"`
}

type LSPCompletionList struct {
	IsIncomplete bool                `json:"isIncomplete"`
	Items        []LSPCompletionItem `json:"items"`
}

type LSPTextEdit struct {
	Range   LSPRange `json:"range"`
	NewText string   `json:"newText"`
}

type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type SemanticTokensResult struct {
	Data []int `json:"data"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Server
// ──────────────────────────────────────────────────────────────────────────────

// Server is the ASON LSP server.
type Server struct {
	reader      *bufio.Reader
	writer      io.Writer
	mu          sync.Mutex
	docs        map[string]string // uri → content
	trees       map[string]*Node  // uri → parsed AST
	logger      *log.Logger
	shutdown    bool
	initialized bool
}

// NewServer creates a new Server reading from r and writing to w.
func NewServer(r io.Reader, w io.Writer) *Server {
	return &Server{
		reader: bufio.NewReaderSize(r, 65536),
		writer: w,
		docs:   make(map[string]string),
		trees:  make(map[string]*Node),
		logger: log.New(os.Stderr, "[ason-lsp] ", log.Ltime),
	}
}

// Run starts the main message loop.
func (s *Server) Run() error {
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		s.handleMessage(msg)
		if s.shutdown {
			return nil
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Transport (base protocol)
// ──────────────────────────────────────────────────────────────────────────────

func (s *Server) readMessage() (*jsonRPCMessage, error) {
	// Read headers
	contentLen := 0
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length: ") {
			n, _ := strconv.Atoi(strings.TrimPrefix(line, "Content-Length: "))
			contentLen = n
		}
	}
	if contentLen == 0 {
		return nil, fmt.Errorf("missing Content-Length")
	}

	// Read body
	body := make([]byte, contentLen)
	_, err := io.ReadFull(s.reader, body)
	if err != nil {
		return nil, err
	}

	var msg jsonRPCMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (s *Server) sendResponse(id *json.RawMessage, result interface{}) {
	data, _ := json.Marshal(result)
	msg := jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  json.RawMessage(data),
	}
	s.writeMessage(&msg)
}

func (s *Server) sendError(id *json.RawMessage, code int, message string) {
	msg := jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonRPCError{Code: code, Message: message},
	}
	s.writeMessage(&msg)
}

func (s *Server) sendNotification(method string, params interface{}) {
	data, _ := json.Marshal(params)
	msg := jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  json.RawMessage(data),
	}
	s.writeMessage(&msg)
}

func (s *Server) writeMessage(msg *jsonRPCMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	body, _ := json.Marshal(msg)
	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n", len(body))
	s.writer.Write([]byte(header))
	s.writer.Write(body)
}

// ──────────────────────────────────────────────────────────────────────────────
// Message dispatch
// ──────────────────────────────────────────────────────────────────────────────

func (s *Server) handleMessage(msg *jsonRPCMessage) {
	s.logger.Printf("← %s", msg.Method)

	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg)
	case "initialized":
		s.initialized = true
	case "shutdown":
		s.shutdown = true
		s.sendResponse(msg.ID, nil)
	case "exit":
		os.Exit(0)

	case "textDocument/didOpen":
		s.handleDidOpen(msg)
	case "textDocument/didChange":
		s.handleDidChange(msg)
	case "textDocument/didClose":
		s.handleDidClose(msg)
	case "textDocument/hover":
		s.handleHover(msg)
	case "textDocument/completion":
		s.handleCompletion(msg)
	case "textDocument/formatting":
		s.handleFormatting(msg)
	case "textDocument/semanticTokens/full":
		s.handleSemanticTokens(msg)

	default:
		if msg.ID != nil {
			// Unknown request → method not found
			s.sendError(msg.ID, -32601, "method not found: "+msg.Method)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Handlers
// ──────────────────────────────────────────────────────────────────────────────

func (s *Server) handleInitialize(msg *jsonRPCMessage) {
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: 1, // Full sync
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{":", "{", "[", "(", ","},
			},
			HoverProvider:              true,
			DocumentFormattingProvider: true,
			SemanticTokensProvider: &SemanticTokensOptions{
				Full: true,
				Legend: SemanticLegend{
					TokenTypes: []string{
						"keyword",   // 0: structural chars
						"type",      // 1: type hints
						"variable",  // 2: field names
						"string",    // 3: string values
						"number",    // 4: numbers
						"comment",   // 5: comments
						"operator",  // 6: colon, comma
						"parameter", // 7: booleans
					},
					TokenModifiers: []string{},
				},
			},
		},
		ServerInfo: &ServerInfo{
			Name:    "ason-lsp",
			Version: "0.1.0",
		},
	}
	s.sendResponse(msg.ID, result)
}

func (s *Server) handleDidOpen(msg *jsonRPCMessage) {
	var params DidOpenTextDocumentParams
	json.Unmarshal(msg.Params, &params)
	uri := params.TextDocument.URI
	s.docs[uri] = params.TextDocument.Text
	s.updateDiagnostics(uri)
}

func (s *Server) handleDidChange(msg *jsonRPCMessage) {
	var params DidChangeTextDocumentParams
	json.Unmarshal(msg.Params, &params)
	uri := params.TextDocument.URI
	if len(params.ContentChanges) > 0 {
		s.docs[uri] = params.ContentChanges[len(params.ContentChanges)-1].Text
	}
	s.updateDiagnostics(uri)
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

func (s *Server) handleDidClose(msg *jsonRPCMessage) {
	var params DidCloseTextDocumentParams
	json.Unmarshal(msg.Params, &params)
	uri := params.TextDocument.URI
	delete(s.docs, uri)
	delete(s.trees, uri)
	// Clear diagnostics
	s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: []LSPDiagnostic{},
	})
}

func (s *Server) handleHover(msg *jsonRPCMessage) {
	var params TextDocumentPositionParams
	json.Unmarshal(msg.Params, &params)
	uri := params.TextDocument.URI
	root := s.getTree(uri)
	if root == nil {
		s.sendResponse(msg.ID, nil)
		return
	}
	text := HoverInfo(root, params.Position.Line, params.Position.Character)
	if text == "" {
		s.sendResponse(msg.ID, nil)
		return
	}
	s.sendResponse(msg.ID, LSPHover{
		Contents: LSPMarkupContent{Kind: "markdown", Value: text},
	})
}

func (s *Server) handleCompletion(msg *jsonRPCMessage) {
	var params TextDocumentPositionParams
	json.Unmarshal(msg.Params, &params)
	uri := params.TextDocument.URI
	src := s.docs[uri]
	root := s.getTree(uri)

	items := Complete(root, src, params.Position.Line, params.Position.Character)
	var lspItems []LSPCompletionItem
	for _, it := range items {
		lspItems = append(lspItems, LSPCompletionItem{
			Label:      it.Label,
			Kind:       it.Kind,
			Detail:     it.Detail,
			InsertText: it.InsertText,
		})
	}
	s.sendResponse(msg.ID, LSPCompletionList{
		IsIncomplete: false,
		Items:        lspItems,
	})
}

func (s *Server) handleFormatting(msg *jsonRPCMessage) {
	var params DocumentFormattingParams
	json.Unmarshal(msg.Params, &params)
	uri := params.TextDocument.URI
	src, ok := s.docs[uri]
	if !ok {
		s.sendResponse(msg.ID, nil)
		return
	}

	formatted := Format(src)
	if formatted == src {
		s.sendResponse(msg.ID, []LSPTextEdit{})
		return
	}

	// Replace entire document
	lines := strings.Count(src, "\n")
	lastLineLen := len(src)
	if idx := strings.LastIndex(src, "\n"); idx >= 0 {
		lastLineLen = len(src) - idx - 1
	}

	edits := []LSPTextEdit{{
		Range: LSPRange{
			Start: LSPPosition{Line: 0, Character: 0},
			End:   LSPPosition{Line: lines, Character: lastLineLen},
		},
		NewText: formatted,
	}}
	s.sendResponse(msg.ID, edits)
}

func (s *Server) handleSemanticTokens(msg *jsonRPCMessage) {
	var params SemanticTokensParams
	json.Unmarshal(msg.Params, &params)
	uri := params.TextDocument.URI
	src, ok := s.docs[uri]
	if !ok {
		s.sendResponse(msg.ID, SemanticTokensResult{Data: []int{}})
		return
	}

	lex := NewLexer(src)
	tokens := lex.All()

	var data []int
	prevLine := 0
	prevCol := 0

	for _, tok := range tokens {
		tokenType := -1
		switch tok.Type {
		case TokenLBrace, TokenRBrace, TokenLParen, TokenRParen, TokenLBracket, TokenRBracket:
			tokenType = 0 // keyword
		case TokenTypeHint, TokenMapKw:
			tokenType = 1 // type
		case TokenIdent:
			tokenType = 2 // variable
		case TokenString, TokenPlainStr:
			tokenType = 3 // string
		case TokenNumber:
			tokenType = 4 // number
		case TokenComment:
			tokenType = 5 // comment
		case TokenColon, TokenComma:
			tokenType = 6 // operator
		case TokenBool:
			tokenType = 7 // parameter (booleans)
		}
		if tokenType < 0 {
			continue
		}

		length := tok.EndOff - tok.Offset
		if length <= 0 {
			continue
		}

		deltaLine := tok.Line - prevLine
		deltaCol := tok.Col
		if deltaLine == 0 {
			deltaCol = tok.Col - prevCol
		}

		data = append(data, deltaLine, deltaCol, length, tokenType, 0)
		prevLine = tok.Line
		prevCol = tok.Col
	}

	s.sendResponse(msg.ID, SemanticTokensResult{Data: data})
}

// ──────────────────────────────────────────────────────────────────────────────
// Diagnostics pipeline
// ──────────────────────────────────────────────────────────────────────────────

func (s *Server) updateDiagnostics(uri string) {
	src, ok := s.docs[uri]
	if !ok {
		return
	}

	root, parseDiags := Parse(src)
	s.trees[uri] = root

	analysisDiags := Analyze(root, src)
	allDiags := append(parseDiags, analysisDiags...)

	var lspDiags []LSPDiagnostic
	for _, d := range allDiags {
		lspDiags = append(lspDiags, LSPDiagnostic{
			Range: LSPRange{
				Start: LSPPosition{Line: d.StartLine, Character: d.StartCol},
				End:   LSPPosition{Line: d.EndLine, Character: d.EndCol},
			},
			Severity: int(d.Severity),
			Source:   "ason-lsp",
			Message:  d.Message,
		})
	}
	if lspDiags == nil {
		lspDiags = []LSPDiagnostic{}
	}

	s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: lspDiags,
	})
}

func (s *Server) getTree(uri string) *Node {
	if tree, ok := s.trees[uri]; ok {
		return tree
	}
	src, ok := s.docs[uri]
	if !ok {
		return nil
	}
	tree, _ := Parse(src)
	s.trees[uri] = tree
	return tree
}

// ──────────────────────────────────────────────────────────────────────────────
// HTTP debug mode (optional)
// ──────────────────────────────────────────────────────────────────────────────

func startHTTPDebug(addr string) {
	http.HandleFunc("/parse", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		root, diags := Parse(string(body))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ast":         root,
			"diagnostics": diags,
		})
	})
	go http.ListenAndServe(addr, nil)
}
