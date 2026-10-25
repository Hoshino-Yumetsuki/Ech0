// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package server

import (
	"context"
	"net"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
)

func TestServerStopWithoutHTTPServer(t *testing.T) {
	s := &Server{}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}

func TestServerStartFailsWhenPortAlreadyBound(t *testing.T) {
	config.Config().Server.Port = "6288"

	ln, err := net.Listen("tcp", ":6288")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer func() { _ = ln.Close() }()

	s := New(gin.New())
	if err := s.Start(context.Background()); err == nil {
		t.Fatalf("expected start failure when port is already bound")
	}
}

func TestListenAddr(t *testing.T) {
	tests := []struct {
		host string
		port string
		want string
	}{
		{host: "", port: "6277", want: ":6277"},
		{host: "0.0.0.0", port: "6277", want: ":6277"},
		{host: "::", port: "6277", want: ":6277"},
		{host: "127.0.0.1", port: "6277", want: "127.0.0.1:6277"},
		{host: "::1", port: "6277", want: "[::1]:6277"},
		{host: "localhost", port: "6277", want: "localhost:6277"},
	}
	for _, tt := range tests {
		if got := listenAddr(tt.host, tt.port); got != tt.want {
			t.Errorf("listenAddr(%q, %q) = %q, want %q", tt.host, tt.port, got, tt.want)
		}
	}
}
