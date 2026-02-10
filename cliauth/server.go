package cliauth

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type CallbackResult struct {
	Code  string
	State string
	Error string
}

type CallbackServer struct {
	listener net.Listener
	server   *http.Server
	result   chan CallbackResult
}

const callbackPort = 8080

func NewCallbackServer() (*CallbackServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", callbackPort))
	if err != nil {
		return nil, fmt.Errorf("listen on localhost:%d: %w", callbackPort, err)
	}

	cs := &CallbackServer{
		listener: listener,
		result:   make(chan CallbackResult, 1),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", cs.handleCallback)

	cs.server = &http.Server{Handler: mux}

	return cs, nil
}

func (cs *CallbackServer) RedirectURL() string {
	return fmt.Sprintf("http://localhost:%d", callbackPort)
}

func (cs *CallbackServer) Start() {
	go cs.server.Serve(cs.listener)
}

func (cs *CallbackServer) WaitForCallback(ctx context.Context) (CallbackResult, error) {
	select {
	case result := <-cs.result:
		return result, nil
	case <-ctx.Done():
		return CallbackResult{}, ctx.Err()
	}
}

func (cs *CallbackServer) Shutdown(ctx context.Context) error {
	return cs.server.Shutdown(ctx)
}

func (cs *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	result := CallbackResult{
		Code:  query.Get("code"),
		State: query.Get("state"),
		Error: query.Get("error"),
	}

	if result.Error != "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authentication Failed</title></head>
<body>
<h1>Authentication Failed</h1>
<p>Error: %s</p>
<p>%s</p>
<p>You can close this window.</p>
</body>
</html>`, result.Error, query.Get("error_description"))
	} else {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Authentication Successful</title></head>
<body>
<h1>Authentication Successful</h1>
<p>You have successfully logged in to Polar Signals CLI.</p>
<p>You can close this window and return to your terminal.</p>
</body>
</html>`)
	}

	cs.result <- result
}
