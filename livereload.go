// Copyright 2020 The Verbis Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package livereload

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

// Handler is a HandlerFunc handling the livereload
// Websocket interaction.
func Handler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	wsHub.register <- c
	defer func() { wsHub.unregister <- c }()
	go c.writer()
	c.reader()
}

// Initialize starts the Websocket Hub handling live reloads.
func Initialize() {
	go wsHub.run()
}

// ForceRefresh tells livereload to force a hard refresh.
func ForceRefresh() {
	RefreshPath("/x.js")
}

// NavigateToPath tells livereload to navigate to the given path.
// This translates to `window.location.href = path` in the client.
func NavigateToPath(path string) {
	RefreshPath(path)
}

// RefreshPath tells livereload to refresh only the given path.
// If that path points to a CSS stylesheet or an image, only the changes
// will be updated in the browser, not the entire page.
func RefreshPath(s string) {
	refreshPathForPort(s, -1)
}

func refreshPathForPort(s string, port int) {
	// Tell livereload a file has changed - will force a hard refresh if not CSS or an image
	urlPath := filepath.ToSlash(s)
	portStr := ""
	if port > 0 {
		portStr = fmt.Sprintf(`, "overrideURL": %d`, port)
	}
	msg := fmt.Sprintf(`{"command":"reload","path":%q,"originalPath":"","liveCSS":true,"liveImg":true%s}`, urlPath, portStr)
	wsHub.broadcast <- []byte(msg)
}

// ServeJS serves the liverreload.js who's reference is injected into the page.
func ServeJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(javascript))
}
