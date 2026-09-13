package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/oauth2"
)

type StateStore struct {
	mu    sync.Mutex
	dir   string
	file  string
	state PersistedState
}

func NewStateStore(dir string) *StateStore {
	return &StateStore{dir: dir, file: filepath.Join(dir, "state.json"), state: PersistedState{Chats: map[string]*ChatState{}}}
}

func (s *StateStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(s.file)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &s.state); err != nil {
		return err
	}
	if s.state.Chats == nil {
		s.state.Chats = map[string]*ChatState{}
	}
	return nil
}

func (s *StateStore) Chat(id string) ChatState {
	s.mu.Lock()
	defer s.mu.Unlock()
	if chat := s.state.Chats[id]; chat != nil {
		return *chat
	}
	return ChatState{}
}

func (s *StateStore) PutChat(id string, chat ChatState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chats[id] = &chat
	return s.saveLocked()
}

func (s *StateStore) OAuth() (*oauth2.Config, *oauth2.Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.OAuth == nil {
		return nil, nil
	}
	return s.state.OAuth.Config, s.state.OAuth.Token
}

func (s *StateStore) PutOAuth(config *oauth2.Config, token *oauth2.Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.OAuth = &OAuthState{Config: config, Token: token}
	return s.saveLocked()
}

func (s *StateStore) saveLocked() error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, append(b, '\n'), 0o600)
}
