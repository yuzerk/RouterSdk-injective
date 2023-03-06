package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/tools/crypto"
	"github.com/anyswap/RouterSDK-injective/config"

	"github.com/gorilla/mux"
)

func addAuthenticationMiddleware(router *mux.Router) {
	tokCount := len(config.GetServerConfig().SessionTokens)
	if tokCount == 0 {
		return
	}
	amw := authenticationMiddleware{authedTokens: make(map[string]*config.SessionToken)}
	amw.Populate()
	router.Use(amw.Middleware)
	log.Info("enable auth session token", "tokens", tokCount)
}

type authenticationMiddleware struct {
	authedTokens map[string]*config.SessionToken
}

func (amw *authenticationMiddleware) Populate() {
	cfg := config.GetServerConfig()
	for _, tok := range cfg.SessionTokens {
		amw.authedTokens[tok.Token] = tok
	}
}

// Middleware function, which will be called for each request
func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessToken := r.Header.Get("X-Session-Token")
		parts := strings.Split(sessToken, ":")
		if len(parts) != 3 {
			log.Debug("rpc call with wrong token", "token", sessToken)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		token := parts[0]
		timestamp := parts[1]
		signature := parts[2]
		log.Debug("rpc call with sig", "token", token, "timestamp", timestamp, "signature", signature)

		statTotalCalls(token)

		tokinfo, ok := amw.authedTokens[token]
		if !ok {
			log.Debug("rpc call with unauth token", "token", token)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if err := verifySignature(tokinfo, timestamp, signature); err != nil {
			log.Debug("rpc call verify token failed", "token", token, "err", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		statSucessCalls(token)

		next.ServeHTTP(w, r)
	})
}

// use milli seconds unit, accept recent two minutes
func verifyTimestamp(timestamp string) error {
	ts, err := common.GetUint64FromStr(timestamp)
	if err != nil {
		return fmt.Errorf("wrong timestamp format. %w", err)
	}

	const future = uint64(10 * 1000)
	const delay = uint64(2 * 60 * 1000)

	now := uint64(common.NowMilli())

	if ts > now+future {
		return fmt.Errorf("future timestamp")
	}
	if ts+delay < now {
		return fmt.Errorf("expired timestamp")
	}
	return nil
}

type signMessage struct {
	User string `json:"user"`
	Salt string `json:"salt"`
	Time string `json:"time"`
}

func verifySignature(tok *config.SessionToken, timestamp, sig string) error {
	if err := verifyTimestamp(timestamp); err != nil {
		return err
	}

	signature := common.FromHex(sig)
	if len(signature) == crypto.SignatureLength {
		signature = signature[:crypto.SignatureLength-1]
	} else if len(signature) != crypto.SignatureLength-1 {
		return fmt.Errorf("wrong signature length")
	}

	msg, err := json.Marshal(&signMessage{
		User: tok.User,
		Salt: tok.Salt,
		Time: timestamp,
	})
	if err != nil {
		return fmt.Errorf("marshal sign message failed. %w", err)
	}
	hash := common.Keccak256Hash(msg).Bytes()

	pubkey := common.FromHex(tok.Token)
	if !crypto.VerifySignature(pubkey, hash, signature) {
		return fmt.Errorf("verify signature failed")
	}
	return nil
}
