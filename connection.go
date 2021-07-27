// Copyright 2020 The Verbis Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package livereload

import (
	"bytes"
	"github.com/gorilla/websocket"
	"sync"
)

// Connection defines a singular connection to the ws.
// Including the []byte chan for sending websocket.Message's
// to the client.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	send chan []byte
	// Potential data race.
	closer sync.Once
}

// close closes the connection once.
func (c *connection) close() {
	c.closer.Do(func() {
		close(c.send)
	})
}

// reader reads incoming messages and sends the hello
// command to the client.
func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		if bytes.Contains(message, []byte(`"command":"hello"`)) {
			c.send <- []byte(`{
				"command": "hello",
				"protocols": [ "http://livereload.com/protocols/official-7" ],
				"serverName": "Verbis"
			}`)
		}
	}
	c.ws.Close()
}

// writer writes the websocket.TextMessage to the websocket.
func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}
