package session

import (
	"net/http"
)

// SessionID 是 session 在 cookie 中的 name.
const SessionID = "RecoitSessionID"

// Manager 是 session manager.
type Manager struct {
	store  map[string]bool
	name   string
	maxAge int
}

// NewManager .
func NewManager(maxAge int) *Manager {
	return &Manager{
		store:  make(map[string]bool),
		name:   SessionID,
		maxAge: maxAge,
	}
}

func (manager *Manager) newSession(sid string) http.Cookie {
	return http.Cookie{
		Name:     manager.name,
		Value:    sid,
		Path:     "/", // important
		MaxAge:   manager.maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// Add adds a new sid into manager.store.
func (manager *Manager) Add(w http.ResponseWriter, sid string) {
	session := manager.newSession(sid)
	http.SetCookie(w, &session)
	manager.store[sid] = true
}

// Check 检查 session 的有效性，有效时返回 true.
func (manager *Manager) Check(r *http.Request) bool {
	cookie, err := r.Cookie(manager.name)
	if err != nil || cookie.Value == "" || !manager.store[cookie.Value] {
		return false
	}
	return true
}

// DeleteSID 同时删除 manager.store 中的 sid 并使 session 过期。
func (manager *Manager) DeleteSID(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.name)
	if err == nil {
		manager.store[cookie.Value] = false
	}
	session := http.Cookie{
		Name:     manager.name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &session)
}
