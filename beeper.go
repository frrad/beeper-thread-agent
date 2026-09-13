package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"golang.org/x/oauth2"
)

type BeeperClient struct {
	Endpoint string
	Token    string
	Store    *StateStore
	session  *mcp.ClientSession
	server   *http.Server
}

func (b *BeeperClient) Connect(ctx context.Context) error {
	transport := &mcp.StreamableClientTransport{Endpoint: b.Endpoint, DisableStandaloneSSE: true}
	if b.Token != "" {
		transport.HTTPClient = &http.Client{Transport: bearerTransport{token: b.Token, base: http.DefaultTransport}}
	} else {
		handler, err := b.oauthHandler(ctx)
		if err != nil {
			return err
		}
		transport.OAuthHandler = handler
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "beeper-thread-agent", Version: "0.2.0"}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("connect to Beeper MCP: %w", err)
	}
	b.session = session
	return nil
}

func (b *BeeperClient) oauthHandler(ctx context.Context) (*auth.AuthorizationCodeHandler, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start OAuth callback listener: %w", err)
	}
	callbackURL := "http://" + listener.Addr().String() + "/callback"
	results := make(chan *auth.AuthorizationResult, 1)
	errs := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if oauthErr := q.Get("error"); oauthErr != "" {
			errs <- fmt.Errorf("OAuth authorization failed: %s", oauthErr)
			http.Error(w, "Authorization failed. Return to the terminal.", http.StatusBadRequest)
			return
		}
		results <- &auth.AuthorizationResult{Code: q.Get("code"), State: q.Get("state"), Iss: q.Get("iss")}
		fmt.Fprint(w, "Beeper authorized. You can close this tab and return to the terminal.")
	})
	b.server = &http.Server{Handler: mux}
	go func() {
		if err := b.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	var initial oauth2.TokenSource
	if config, token := b.Store.OAuth(); config != nil && token != nil {
		initial = &savingTokenSource{base: config.TokenSource(ctx, token), config: config, store: b.Store}
	}
	handler, err := auth.NewAuthorizationCodeHandler(&auth.AuthorizationCodeHandlerConfig{
		RedirectURL: callbackURL,
		DynamicClientRegistrationConfig: &auth.DynamicClientRegistrationConfig{Metadata: &oauthex.ClientRegistrationMetadata{
			RedirectURIs: []string{callbackURL}, ClientName: "Beeper Thread Agent",
			TokenEndpointAuthMethod: "none", GrantTypes: []string{"authorization_code", "refresh_token"}, ResponseTypes: []string{"code"},
		}},
		RequestRefreshToken: true,
		InitialTokenSource:  initial,
		NewTokenSource: func(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (oauth2.TokenSource, error) {
			if err := b.Store.PutOAuth(config, token); err != nil {
				return nil, err
			}
			return &savingTokenSource{base: config.TokenSource(ctx, token), config: config, store: b.Store}, nil
		},
		AuthorizationCodeFetcher: func(ctx context.Context, args *auth.AuthorizationArgs) (*auth.AuthorizationResult, error) {
			fmt.Printf("Authorize Beeper in your browser: %s\n", args.URL)
			if err := openBrowser(args.URL); err != nil {
				fmt.Printf("Could not open browser automatically: %v\n", err)
			}
			select {
			case result := <-results:
				return result, nil
			case err := <-errs:
				return nil, err
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	})
	if err != nil {
		_ = b.server.Close()
		return nil, err
	}
	return handler, nil
}

func (b *BeeperClient) Close() error {
	if b.server != nil {
		_ = b.server.Close()
	}
	if b.session != nil {
		return b.session.Close()
	}
	return nil
}

func (b *BeeperClient) call(ctx context.Context, name string, arguments map[string]any) (string, error) {
	if b.session == nil {
		return "", errors.New("Beeper is not connected")
	}
	result, err := b.session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: arguments})
	if err != nil {
		return "", err
	}
	for _, content := range result.Content {
		if text, ok := content.(*mcp.TextContent); ok {
			if result.IsError {
				return "", errors.New(text.Text)
			}
			return text.Text, nil
		}
	}
	return "", fmt.Errorf("%s returned no text result", name)
}

func (b *BeeperClient) ListMessages(ctx context.Context, chatID string) (MessagePage, error) {
	text, err := b.call(ctx, "list_messages", map[string]any{"chatID": chatID})
	if err != nil {
		return MessagePage{}, err
	}
	var page MessagePage
	if err := json.Unmarshal([]byte(text), &page); err != nil {
		return MessagePage{}, fmt.Errorf("decode list_messages: %w", err)
	}
	return page, nil
}

func (b *BeeperClient) SearchChats(ctx context.Context, query string) (string, error) {
	return b.call(ctx, "search", map[string]any{"query": query})
}

func (b *BeeperClient) SendMessage(ctx context.Context, chatID, text string) (string, error) {
	return b.call(ctx, "send_message", map[string]any{"chatID": chatID, "text": text})
}

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t bearerTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	clone := request.Clone(request.Context())
	clone.Header = request.Header.Clone()
	clone.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(clone)
}

type savingTokenSource struct {
	base   oauth2.TokenSource
	config *oauth2.Config
	store  *StateStore
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.base.Token()
	if err == nil {
		err = s.store.PutOAuth(s.config, token)
	}
	return token, err
}

func openBrowser(rawURL string) error {
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return err
	}
	var command string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		command, args = "open", []string{rawURL}
	case "windows":
		command, args = "rundll32", []string{"url.dll,FileProtocolHandler", rawURL}
	default:
		command, args = "xdg-open", []string{rawURL}
	}
	return exec.Command(command, args...).Start()
}
