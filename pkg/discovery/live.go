package discovery

import (
	"context"
	"net"
	"time"
)

type LivePortScanner struct {
	Timeout time.Duration
}

func NewLivePortScanner() *LivePortScanner {
	return &LivePortScanner{Timeout: 2 * time.Second}
}

// CheckEndpoint verifies if an MCP Server or microservice port is listening.
func (l *LivePortScanner) CheckEndpoint(ctx context.Context, hostPort string) bool {
	d := net.Dialer{Timeout: l.Timeout}
	conn, err := d.DialContext(ctx, "tcp", hostPort)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
