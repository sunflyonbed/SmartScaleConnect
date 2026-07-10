package internal

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
	"github.com/stretchr/testify/require"
)

func TestStoreWeightsDedupBySyncID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scaleconnect.db")
	date := time.Date(2026, 7, 9, 8, 0, 0, 0, time.UTC)
	weights := []*core.Weight{
		{Date: date, Weight: 70, BMI: 22},
	}

	stats, err := StoreWeights(path, "sync-a", weights)
	require.NoError(t, err)
	require.Equal(t, 1, stats.Synced)
	require.Equal(t, 1, stats.New)
	require.Len(t, stats.NewWeights, 1)

	stats, err = StoreWeights(path, "sync-a", weights)
	require.NoError(t, err)
	require.Equal(t, 1, stats.Synced)
	require.Equal(t, 0, stats.New)
	require.Empty(t, stats.NewWeights)

	stats, err = StoreWeights(path, "sync-b", weights)
	require.NoError(t, err)
	require.Equal(t, 1, stats.Synced)
	require.Equal(t, 1, stats.New)
	require.Len(t, stats.NewWeights, 1)

	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	defer db.Close()

	var total int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM weights").Scan(&total))
	require.Equal(t, 2, total)

	var countA int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM weights WHERE sync_id = ?", "sync-a").Scan(&countA))
	require.Equal(t, 1, countA)
}
