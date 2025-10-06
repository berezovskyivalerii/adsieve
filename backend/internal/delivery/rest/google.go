package rest

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func randBytes(n int) []byte { b := make([]byte, n); _, _ = rand.Read(b); return b }
func pkceVerifier() string {
	v := base64.RawURLEncoding.EncodeToString(randBytes(32))
	if len(v) < 43 {
		v += strings.Repeat("x", 43-len(v))
	}
	return v
}
func pkceChallengeS256(v string) string {
	sum := sha256.Sum256([]byte(v))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// POST /integrations/google/connect
func (h *Handler) googleConnect(c *gin.Context) {
	uidVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no user"})
		return
	}
	userID := uidVal.(int64)

	verifier := pkceVerifier()
	challenge := pkceChallengeS256(verifier)
	state := fmt.Sprintf("gads_%d_%d", userID, time.Now().UnixNano())

	if err := h.oauthStates.Save(c.Request.Context(), state, verifier, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save state"})
		return
	}

	url := h.oauthCfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	c.JSON(http.StatusOK, gin.H{"redirect_url": url})
}

// GET /integrations/google/callback
func (h *Handler) googleCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")
	if state == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state/code missing"})
		return
	}
	userID, verifier, err := h.oauthStates.Consume(c, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	tok, err := h.oauthCfg.Exchange(context.Background(), code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "exchange failed"})
		return
	}
	if tok.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no refresh_token (prompt=consent, offline)"})
		return
	}

	var googleUserID string
	if raw, ok := tok.Extra("id_token").(string); ok {
		googleUserID = raw
	} // разберёшь в репозитории

	if err := h.tokenVault.SaveGoogleRefreshToken(c, userID, googleUserID, tok.RefreshToken, "adwords"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save refresh failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GET /integrations/google/accounts
func (h *Handler) googleAccounts(c *gin.Context) {
	uidVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no user"})
		return
	}
	userID := uidVal.(int64)

	ids, err := h.gadsClient.ListAccessibleCustomers(c, userID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "list failed"})
		return
	}

	type item struct {
		CustomerID string `json:"customer_id"`
	}
	out := make([]item, 0, len(ids))
	for _, id := range ids {
		out = append(out, item{CustomerID: id})
	}
	c.JSON(http.StatusOK, gin.H{"accounts": out})
}

// POST /integrations/google/link-accounts
type linkReq struct {
	CustomerIDs []string `json:"customer_ids"`
}

func (h *Handler) googleLinkAccounts(c *gin.Context) {
	uidVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no user"})
		return
	}
	userID := uidVal.(int64)

	var req linkReq
	if err := c.ShouldBindJSON(&req); err != nil || len(req.CustomerIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_ids required"})
		return
	}
	if err := h.gadsClient.LinkAccounts(c, userID, req.CustomerIDs); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "link failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}
