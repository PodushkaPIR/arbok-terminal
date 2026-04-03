package pty

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
)

type Winsize struct {
	Rows uint16
	Cols uint16
	X    uint16
	Y    uint16
}

type Manager struct {
	Cmd      *exec.Cmd
	PTY      *os.File
	OutputCh chan []byte
	InputCh  chan []byte
	SizeCh   chan Winsize
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
		InputCh:  make(chan []byte, 256),
		SizeCh:   make(chan Winsize, 10),
		DoneCh:   make(chan struct{}),
	}

	go m.readLoop()
	go m.handleSignals()

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
			return
		}
	}
}

func (m *Manager) handleSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)

	for {
		select {
		case <-ch:
			rows, cols, err := pty.Getsize(m.PTY)
			if err != nil {
				continue
			}
			m.SizeCh <- Winsize{
				Rows: uint16(rows),
				Cols: uint16(cols),
			}
		case <-m.DoneCh:
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
		m.Cmd.Process.Kill()
	}
	m.Cmd.Wait()
	return nil
}
