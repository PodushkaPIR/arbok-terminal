package pty

import (
	"io"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

type Manager struct {
	Cmd      *exec.Cmd
	PTY      *os.File
	OutputCh chan []byte
	DoneCh   chan struct{}
}

func New(shell string, cols, rows int) (*Manager, error) {
	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}
	}

	cmd := exec.Command(shell)
	cmd.Env = os.Environ()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})

	m := &Manager{
		Cmd:      cmd,
		PTY:      ptmx,
		OutputCh: make(chan []byte, 4096),
		DoneCh:   make(chan struct{}),
	}

	slog.Info("pty: created", "shell", shell, "pid", cmd.Process.Pid, "cols", cols, "rows", rows)

	go m.readLoop()

	return m, nil
}

func (m *Manager) readLoop() {
	defer close(m.OutputCh)

	buf := make([]byte, 8192)
	for {
		n, err := m.PTY.Read(buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			m.OutputCh <- data
		}
		if err != nil {
			if err == io.EOF {
				return
			}
			slog.Error("pty: read error", "err", err)
			return
		}
	}
}

func (m *Manager) Write(data []byte) error {
	_, err := m.PTY.Write(data)
	return err
}

func (m *Manager) Resize(cols, rows int) error {
	return pty.Setsize(m.PTY, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

func (m *Manager) Close() error {
	close(m.DoneCh)
	m.PTY.Close()

	if m.Cmd.Process != nil {
		done := make(chan error, 1)
		go func() { done <- m.Cmd.Wait() }()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			m.Cmd.Process.Kill()
			<-done
		}
	}
	return nil
}
