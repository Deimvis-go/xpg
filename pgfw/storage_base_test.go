package pgfw

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/Deimvis-go/logs/logs"
	"github.com/Deimvis-go/xpg/pg"
	"github.com/Deimvis-go/xpg/pg/pgconn"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestStorageBase(t *testing.T) {
	t.Run("pg-conn-fallback", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		z := zap.New(core, zap.AddCaller()).Sugar()
		lg := logs.ZapAsKVCtxLogger(z)

		pm := NewMockPoolManager(t)
		pm.EXPECT().GetPool(mock.Anything).Return(nil).Maybe()
		b := NewStorageBase(pm, lg)

		t.Run("log-fallback", func(t *testing.T) {
			defer recorded.TakeAll()

			_ = b.RWConn(context.Background())
			z.Sync()

			logRecords := recorded.All()
			require.Len(t, logRecords, 1)
			log0 := logRecords[0]
			require.Equal(t, log0.Message, pgConnAcquireFallbackLogMsg)
		})
		t.Run("no-log-when-no-fallback", func(t *testing.T) {
			defer recorded.TakeAll()

			ctx := context.Background()
			ctx = context.WithValue(ctx, pg.CtxConnKey(pgconn.RW), pg.Conn(NewMockConn(t)))
			_ = b.RWConn(ctx)
			z.Sync()

			logRecords := recorded.All()
			require.Len(t, logRecords, 0)
		})

	})
}
