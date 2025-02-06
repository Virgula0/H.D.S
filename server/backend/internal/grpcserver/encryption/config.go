package encryption

import "sync"

// ClientConfig represents the configuration for a single client.
type ClientConfig struct {
	ID                string
	EncryptionEnabled bool
}

// ClientConfigStore manages client configurations.
type ClientConfigStore struct {
	mu      sync.RWMutex
	clients map[string]*ClientConfig
}

func NewClientCertStore() *ClientConfigStore {
	return &ClientConfigStore{
		clients: make(map[string]*ClientConfig),
	}
}

func (store *ClientConfigStore) GetClientConfig(clientID string) (*ClientConfig, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	client, exists := store.clients[clientID]
	return client, exists
}

func (store *ClientConfigStore) UpdateClientConfig(clientID string, config *ClientConfig) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.clients[clientID] = config
}
