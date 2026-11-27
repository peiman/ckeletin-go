package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// appendLog is a rotating file opened with O_APPEND. Two processes can write
// the same path, including while the file is created or rotated, without one
// overwriting the other's bytes. lumberjack opens a new file without append
// and then writes at its own offset, which drops and tears lines.
type appendLog struct {
	path       string
	maxBytes   int64
	maxBackups int
	maxAge     time.Duration
	compress   bool

	mu     sync.Mutex
	file   *os.File
	size   int64
	closed bool
}

func openAppendLog(path string, maxBytes int64, maxBackups, maxAgeDays int, compress bool) (*appendLog, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	logFile := &appendLog{
		path:       path,
		maxBytes:   maxBytes,
		maxBackups: maxBackups,
		compress:   compress,
	}
	if maxAgeDays > 0 {
		logFile.maxAge = time.Duration(maxAgeDays) * 24 * time.Hour
	}
	if err := logFile.open(); err != nil {
		return nil, err
	}
	return logFile, nil
}

func (l *appendLog) open() error {
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}
	l.file = file
	l.size = info.Size()
	return nil
}

func (l *appendLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return len(p), nil
	}
	if l.file == nil {
		if err := l.open(); err != nil {
			return 0, err
		}
	}
	if l.maxBytes > 0 && l.size > 0 && l.size+int64(len(p)) > l.maxBytes {
		if err := l.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := l.file.Write(p)
	l.size += int64(n)
	return n, err
}

func (l *appendLog) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed = true
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}

func (l *appendLog) rotate() error {
	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
	}
	if _, err := os.Stat(l.path); err == nil {
		backup := fmt.Sprintf("%s.%s", l.path, time.Now().UTC().Format("20060102T150405.000000000Z"))
		if err := os.Rename(l.path, backup); err != nil && !os.IsNotExist(err) {
			return err
		}
		if l.compress {
			go compressBackup(backup)
		}
	}
	l.prune()
	return l.open()
}

func (l *appendLog) prune() {
	matches, err := filepath.Glob(l.path + ".*")
	if err != nil {
		return
	}
	if l.maxAge > 0 {
		cutoff := time.Now().Add(-l.maxAge)
		kept := matches[:0]
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				_ = os.Remove(match)
				continue
			}
			kept = append(kept, match)
		}
		matches = kept
	}
	if l.maxBackups > 0 && len(matches) > l.maxBackups {
		for _, match := range matches[:len(matches)-l.maxBackups] {
			_ = os.Remove(match)
		}
	}
}

func compressBackup(path string) {
	in, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.OpenFile(path+".gz", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return
	}
	zw := gzip.NewWriter(out)
	_, copyErr := io.Copy(zw, in)
	closeErr := zw.Close()
	if err := out.Close(); err != nil {
		closeErr = err
	}
	if copyErr == nil && closeErr == nil {
		_ = os.Remove(path)
	}
}
