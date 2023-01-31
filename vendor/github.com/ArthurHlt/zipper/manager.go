package zipper

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

var fManager *Manager = mustNewManager(
	NewGitHandler(),
	&HttpHandler{},
	&LocalHandler{},
)

type Manager struct {
	handlers   map[string]Handler
	httpClient *http.Client
}

func mustNewManager(handlers ...Handler) *Manager {
	m, err := NewManager(handlers...)
	if err != nil {
		panic(err)
	}
	return m
}

// Create new manager with given zip handlers
func NewManager(handlers ...Handler) (*Manager, error) {
	m := &Manager{
		handlers: make(map[string]Handler),
		httpClient: &http.Client{
			Timeout: 0,
		},
	}
	err := m.AddHandlers(handlers...)
	return m, err
}

// Set a custom http client for zip handlers which need it
func (m *Manager) SetHttpClient(httpClient *http.Client) {
	m.httpClient = httpClient
	m.httpClient.Timeout = time.Duration(0)
}

// For default manager
//
// Set a custom http client for zip handlers which need it
func SetHttpClient(httpClient *http.Client) {
	fManager.SetHttpClient(httpClient)
}

// For default manager
//
// Create a session for a given path with given handler type.
// Omitting handler type will use auto detection)
func CreateSession(path string, handlerNames ...string) (*Session, error) {
	return fManager.CreateSession(path, handlerNames...)
}

// Create a session for a given path with given handler type.
// Omitting handler type will use auto detection)
func (m *Manager) CreateSession(path string, handlerNames ...string) (*Session, error) {
	handlerName := ""
	if len(handlerNames) > 0 {
		handlerName = handlerNames[0]
	}
	h, err := m.FindHandler(path, handlerName)
	if err != nil {
		return nil, err
	}
	src := NewSource(path)
	SetCtxHttpClient(src, m.httpClient)
	return NewSession(src, h), nil
}

// For default manager
//
// Add new zip handlers to manager
func AddHandlers(handlers ...Handler) error {
	return fManager.AddHandlers(handlers...)
}

// Add new zip handlers to manager
func (m *Manager) AddHandlers(handlers ...Handler) error {
	for _, handler := range handlers {
		err := m.AddHandler(handler)
		if err != nil {
			return err
		}
	}
	return nil
}

// For default manager
//
// Add new zip handler to manager
func AddHandler(handler Handler) error {
	return fManager.AddHandler(handler)
}

// Add new zip handler to manager
func (m *Manager) AddHandler(handler Handler) error {
	name := strings.ToLower(handler.Name())
	if _, ok := m.handlers[name]; ok {
		return fmt.Errorf("Handler %s already exists", name)
	}
	m.handlers[name] = handler
	return nil
}

// For default manager
//
// Find zip handler by its type
// if type is empty string this will use auto-detection
func FindHandler(path string, handlerName string) (Handler, error) {
	return fManager.FindHandler(path, handlerName)
}

// Find zip handler by its type
// if type is empty string this will use auto-detection
func (m *Manager) FindHandler(path string, handlerName string) (Handler, error) {
	src := NewSource(path)
	handlerName = strings.ToLower(handlerName)
	if handlerName == "" {
		for _, h := range m.handlers {
			if h.Detect(src) {
				return h, nil
			}
		}
	}
	if h, ok := m.handlers[handlerName]; ok {
		return h, nil
	}
	return nil, fmt.Errorf("Handler for path '%s' cannot be found.", src.Path)
}
