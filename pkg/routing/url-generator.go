package routing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func TemporarySignedURL(baseURL string, params map[string]string, secret string, expiry time.Duration) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	expires := time.Now().Add(expiry).Unix()
	q.Set("expires", strconv.FormatInt(expires, 10))

	for k, v := range params {
		q.Set(k, v)
	}

	// Build raw string to sign
	raw := u.Path + "?" + q.Encode()

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(raw))
	signature := hex.EncodeToString(mac.Sum(nil))

	q.Set("signature", signature)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func VerifyRequest(r *http.Request, secret string) bool {
	u := r.URL
	q := u.Query()

	expiresStr := q.Get("expires")
	signature := q.Get("signature")

	if expiresStr == "" || signature == "" {
		return false
	}

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || time.Now().Unix() > expires {
		log.Println("signature is expired", r)
		return false
	}

	qCopy := q
	qCopy.Del("signature")
	raw := u.Path + "?" + qCopy.Encode()

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(raw))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}
