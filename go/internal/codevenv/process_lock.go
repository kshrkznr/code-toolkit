package codevenv

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const processLockOwner = "owner.json"

var incompleteLockGrace = 10 * time.Second

type lockOwner struct {
	PID     int       `json:"pid"`
	Token   string    `json:"token"`
	Created time.Time `json:"created"`
}

func acquireProcessLock(path, busyMessage string) (func(), error) {
	for attempt := 0; attempt < 2; attempt++ {
		if err := os.Mkdir(path, 0o700); err == nil {
			owner, err := newLockOwner()
			if err != nil {
				_ = os.Remove(path)
				return nil, err
			}
			data, err := json.Marshal(owner)
			if err == nil {
				err = os.WriteFile(filepath.Join(path, processLockOwner), data, 0o600)
			}
			if err != nil {
				_ = os.RemoveAll(path)
				return nil, fmt.Errorf("record process lock owner: %w", err)
			}
			return func() { releaseProcessLock(path, owner.Token) }, nil
		} else if !os.IsExist(err) {
			return nil, err
		}

		stale, err := processLockIsStale(path, time.Now())
		if err != nil {
			return nil, err
		}
		if !stale || attempt > 0 {
			return nil, fmt.Errorf("%s", busyMessage)
		}
		if err := os.RemoveAll(path); err != nil {
			return nil, fmt.Errorf("remove stale process lock: %w", err)
		}
	}
	return nil, fmt.Errorf("%s", busyMessage)
}

func newLockOwner() (lockOwner, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return lockOwner{}, fmt.Errorf("create process lock token: %w", err)
	}
	return lockOwner{PID: os.Getpid(), Token: hex.EncodeToString(value), Created: time.Now().UTC()}, nil
}

func processLockIsStale(path string, now time.Time) (bool, error) {
	data, err := os.ReadFile(filepath.Join(path, processLockOwner))
	if err == nil {
		var owner lockOwner
		if json.Unmarshal(data, &owner) == nil && owner.PID > 0 && owner.Token != "" {
			return !processAlive(owner.PID), nil
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read process lock owner: %w", err)
	}
	info, statErr := os.Stat(path)
	if statErr != nil {
		return false, statErr
	}
	// A creator may exist between mkdir and owner publication. Old empty or
	// malformed directories are legacy/interrupted locks and can be reclaimed.
	return now.Sub(info.ModTime()) >= incompleteLockGrace, nil
}

func releaseProcessLock(path, token string) {
	data, err := os.ReadFile(filepath.Join(path, processLockOwner))
	if err != nil {
		return
	}
	var owner lockOwner
	if json.Unmarshal(data, &owner) != nil || owner.Token != token {
		return
	}
	_ = os.RemoveAll(path)
}
