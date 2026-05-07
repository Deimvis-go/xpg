package pg

import (
	"fmt"
	"strings"

	"github.com/Deimvis/go-ext/go1.25/ext"
	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
)

func BuildConnUrl(cfg *ConnConfig) string {
	var addr string
	if len(cfg.Servers) > 0 {
		addr = strings.Join(ext.Map(cfg.Servers, func(s ServerLocation) string {
			return fmt.Sprintf("%s:%d", s.Host, s.Port)
		}), ",")
	} else {
		// deprecated
		xmust.True(len(cfg.Host) > 0)
		xmust.True(cfg.Port > 0)
		addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}
	return fmt.Sprintf("postgres://%s:%s@%s/%s", cfg.User, cfg.Password, addr, cfg.Database)
}
