package main

import (
    "github.com/gorilla/sessions"
)

var (
    // Key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
    key   = []byte("super-secret-key")
    store = sessions.NewCookieStore(key)
)

func init() {
    store.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   86400 * 7, // 7 days
        HttpOnly: true,
    }
}
