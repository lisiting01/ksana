package auth

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type Manager struct {
	keyFile string
	keys    map[string]struct{}
	mu      sync.RWMutex
	logger  *slog.Logger
}

func NewManager(keyFile string, logger *slog.Logger) (*Manager, error) {
	m := &Manager{
		keyFile: keyFile,
		keys:    make(map[string]struct{}),
		logger:  logger,
	}

	if err := m.loadKeys(); err != nil {
		return nil, fmt.Errorf("failed to load API keys: %w", err)
	}

	return m, nil
}

func (m *Manager) loadKeys() error {
	file, err := os.Open(m.keyFile)
	if err != nil {
		if os.IsNotExist(err) {
			m.logger.Warn("API keys file not found, authentication will reject all requests", "file", m.keyFile)
			return nil
		}
		return fmt.Errorf("failed to open API keys file %s: %w", m.keyFile, err)
	}
	defer file.Close()

	m.mu.Lock()
	defer m.mu.Unlock()

	newKeys := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if line != "" {
			newKeys[line] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading API keys file %s: %w", m.keyFile, err)
	}

	m.keys = newKeys
	m.logger.Info("Loaded API keys", "count", len(m.keys), "file", m.keyFile)

	return nil
}

func (m *Manager) Validate(key string) bool {
	if key == "" {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.keys[key]
	return exists
}

func (m *Manager) Reload() error {
	m.logger.Info("Reloading API keys", "file", m.keyFile)

	if err := m.loadKeys(); err != nil {
		m.logger.Error("Failed to reload API keys", "error", err)
		return err
	}

	m.logger.Info("Successfully reloaded API keys")
	return nil
}