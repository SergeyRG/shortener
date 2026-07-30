package middleware

import (
	"net"
	"net/http"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/logging"
	"go.uber.org/zap"
)

func ForTrustedSubnet(cfg config.Config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			l := logging.Logger
			if cfg.TrustedSubnet == "" {
				rw.WriteHeader(http.StatusForbidden)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				rw.WriteHeader(http.StatusForbidden)
				return
			}

			_, trustedSubnet, err := net.ParseCIDR(cfg.TrustedSubnet)
			if err != nil {
				l.Error("ошибка разбора ip адреса переданного в заголовке 'X-real-ip'", zap.Error(err))
			}

			if !trustedSubnet.Contains(net.ParseIP(realIP)) {
				rw.WriteHeader(http.StatusForbidden)
				return
			}

			h.ServeHTTP(rw, r)
		})
	}
}
