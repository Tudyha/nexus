package session

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/proto"
)

func init() {
	// Minimal init so messageHandlers and managerInstance exist
	messageHandlers = make(map[proto.MessageType]conn.MessageHandler)
	managerInstance = &manager{}
}

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		status int32
		want   int32
	}{
		{StatusInit, 0},
		{StatusPending, 1},
		{StatusReady, 2},
		{StatusClosed, 3},
	}
	for _, tt := range tests {
		if tt.status != tt.want {
			t.Errorf("status constant = %d, want %d", tt.status, tt.want)
		}
	}
}

func TestNewSessionInitState(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Start checkAuth in background; it blocks on ReadMessage
	s := newSession(server)

	if s.Id == "" {
		t.Error("session Id should not be empty")
	}
	if s.status.Load() != StatusPending {
		t.Errorf("initial status = %d, want %d", s.status.Load(), StatusPending)
	}
	if s.closed {
		t.Error("session should not be marked closed initially")
	}

	// Give checkAuth time to start
	time.Sleep(10 * time.Millisecond)

	// Close client side to unblock checkAuth
	client.Close()
	time.Sleep(20 * time.Millisecond)

	// After ReadMessage fails, checkAuth should close the session
	if s.status.Load() != StatusClosed {
		t.Errorf("status after auth failure = %d, want %d", s.status.Load(), StatusClosed)
	}
}

func TestConcurrentClose(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	s := newSession(server)

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Close()
		}()
	}
	wg.Wait()

	// Verify session is closed
	if s.status.Load() != StatusClosed {
		t.Errorf("status after close = %d, want %d", s.status.Load(), StatusClosed)
	}

	// Verify closeCh is closed (should not block)
	select {
	case <-s.closeCh:
	default:
		t.Error("closeCh should be closed")
	}
}

func TestCloseIdempotent(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	s := newSession(server)
	s.Close()
	s.Close() // second close should not panic
}

func TestOpenTunnelBeforeReady(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	s := newSession(server)
	defer s.Close()

	_, err := s.OpenTunnel(proto.TunnelType_TCP, "127.0.0.1:8080")
	if err != errcode.ErrClientNotReady {
		t.Errorf("OpenTunnel before ready: got %v, want %v", err, errcode.ErrClientNotReady)
	}
}

func TestOpenTunnelClosedSession(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	s := newSession(server)
	s.Close()

	_, err := s.OpenTunnel(proto.TunnelType_TCP, "127.0.0.1:8080")
	if err != errcode.ErrClientNotReady {
		t.Errorf("OpenTunnel on closed session: got %v, want %v", err, errcode.ErrClientNotReady)
	}
}

func TestExitBeforeReady(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	s := newSession(server)
	defer s.Close()

	err := s.Exit()
	if err != errcode.ErrClientNotReady {
		t.Errorf("Exit before ready: got %v, want %v", err, errcode.ErrClientNotReady)
	}
}

func TestManagerNewAndGetSession(t *testing.T) {
	// Direct setup without Init() to avoid handler registration that needs MQ
	messageHandlers = make(map[proto.MessageType]conn.MessageHandler)
	managerInstance = &manager{}

	server, client := net.Pipe()
	defer client.Close()

	mgr := GetManager().(*manager)

	err := mgr.NewSession(server)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	// We need to find the session ID. iterate sync.Map
	var sessionID string
	mgr.sessions.Range(func(key, _ any) bool {
		sessionID = key.(string)
		return false
	})
	if sessionID == "" {
		t.Fatal("no session stored after NewSession")
	}

	s, err := mgr.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if s.Id != sessionID {
		t.Errorf("session ID = %s, want %s", s.Id, sessionID)
	}
}

func TestManagerGetSessionNotFound(t *testing.T) {
	messageHandlers = make(map[proto.MessageType]conn.MessageHandler)
	managerInstance = &manager{}
	mgr := GetManager().(*manager)

	_, err := mgr.GetSession("nonexistent")
	if err != errcode.ErrClientDisconnect {
		t.Errorf("GetSession nonexistent: got %v, want %v", err, errcode.ErrClientDisconnect)
	}
}

func TestManagerGetSessionAfterRemoval(t *testing.T) {
	messageHandlers = make(map[proto.MessageType]conn.MessageHandler)
	managerInstance = &manager{}
	mgr := GetManager().(*manager)

	server, client := net.Pipe()
	defer client.Close()

	mgr.NewSession(server)

	var sessionID string
	mgr.sessions.Range(func(key, _ any) bool {
		sessionID = key.(string)
		return false
	})
	if sessionID == "" {
		t.Fatal("no session stored")
	}

	// Manually delete
	mgr.sessions.Delete(sessionID)

	_, err := mgr.GetSession(sessionID)
	if err != errcode.ErrClientDisconnect {
		t.Errorf("GetSession after removal: got %v, want %v", err, errcode.ErrClientDisconnect)
	}
}

func TestSessionCloseRemovesFromManager(t *testing.T) {
	messageHandlers = make(map[proto.MessageType]conn.MessageHandler)
	managerInstance = &manager{}
	mgr := GetManager().(*manager)

	server, client := net.Pipe()
	defer client.Close()

	mgr.NewSession(server)

	var sessionID string
	mgr.sessions.Range(func(key, _ any) bool {
		sessionID = key.(string)
		return false
	})
	if sessionID == "" {
		t.Fatal("no session stored")
	}

	s, _ := mgr.GetSession(sessionID)
	s.Close()

	// Verify session is removed from manager
	_, err := mgr.GetSession(sessionID)
	if err != errcode.ErrClientDisconnect {
		t.Errorf("session should be removed from manager after Close")
	}
}

func TestSendTaskBeforeReady(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	s := newSession(server)
	defer s.Close()

	err := s.SendTask(1, proto.TaskType(0), nil, false, nil)
	if err != errcode.ErrClientNotReady {
		t.Errorf("SendTask before ready: got %v, want %v", err, errcode.ErrClientNotReady)
	}
}

func TestSession_CreateWithPipe(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	s := newSession(server)
	defer s.Close()

	// Verify session ID format
	if len(s.Id) != 36 { // UUID length
		t.Errorf("session Id length = %d, want 36", len(s.Id))
	}
}

func TestConcurrentNewSession(t *testing.T) {
	messageHandlers = make(map[proto.MessageType]conn.MessageHandler)
	managerInstance = &manager{}
	mgr := GetManager().(*manager)

	// Keep client ends open so checkAuth doesn't fail and remove sessions
	clients := make([]net.Conn, 10)
	var wg sync.WaitGroup
		for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			server, client := net.Pipe()
			clients[n] = client
			mgr.NewSession(server)
		}(i)
	}
	wg.Wait()

	// All client connections are still open, so sessions should be present
	count := 0
	mgr.sessions.Range(func(_, _ any) bool {
		count++
		return true
	})
	if count != 10 {
		t.Errorf("expected 10 sessions, got %d", count)
	}

	// Cleanup: close client ends to unblock checkAuth goroutines
	for _, c := range clients {
		c.Close()
	}
}
