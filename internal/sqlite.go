package internal

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
	_ "modernc.org/sqlite"
)

type StoreStats struct {
	Synced     int
	New        int
	NewWeights []*core.Weight
}

func StoreWeights(path, syncID string, weights []*core.Weight) (StoreStats, error) {
	var stats StoreStats

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return stats, err
	}
	defer db.Close()

	if _, err = db.Exec(weightsSchema); err != nil {
		return stats, err
	}

	tx, err := db.Begin()
	if err != nil {
		return stats, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(upsertWeightSQL)
	if err != nil {
		return stats, err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, weight := range weights {
		if weight == nil || weight.Date.IsZero() {
			continue
		}

		data, err := json.Marshal(weight)
		if err != nil {
			return stats, err
		}
		sum := sha256.Sum256(data)
		ts := weight.Date.Unix()

		var exists bool
		if err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM weights WHERE sync_id = ? AND ts = ?)", syncID, ts).Scan(&exists); err != nil {
			return stats, err
		}
		if !exists {
			stats.New++
			stats.NewWeights = append(stats.NewWeights, weight)
		}
		stats.Synced++

		_, err = stmt.Exec(
			syncID,
			ts,
			weight.Date.UTC().Format(time.RFC3339Nano),
			weight.Weight,
			weight.BMI,
			weight.BodyFat,
			weight.BodyWater,
			weight.BoneMass,
			weight.MetabolicAge,
			weight.MuscleMass,
			weight.PhysiqueRating,
			weight.ProteinMass,
			weight.VisceralFat,
			weight.BasalMetabolism,
			weight.BodyScore,
			weight.HeartRate,
			weight.Height,
			weight.SkeletalMuscleMass,
			weight.User,
			weight.Source,
			string(data),
			hex.EncodeToString(sum[:]),
			now,
			now,
		)
		if err != nil {
			return stats, err
		}
	}

	return stats, tx.Commit()
}

const weightsSchema = `
CREATE TABLE IF NOT EXISTS weights (
	sync_id TEXT NOT NULL,
	ts INTEGER NOT NULL,
	date TEXT NOT NULL,
	weight REAL NOT NULL,
	bmi REAL,
	body_fat REAL,
	body_water REAL,
	bone_mass REAL,
	metabolic_age INTEGER,
	muscle_mass REAL,
	physique_rating INTEGER,
	protein_mass REAL,
	visceral_fat INTEGER,
	basal_metabolism INTEGER,
	body_score INTEGER,
	heart_rate INTEGER,
	height REAL,
	skeletal_muscle_mass REAL,
	user TEXT,
	source TEXT,
	data TEXT NOT NULL,
	data_hash TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	PRIMARY KEY (sync_id, ts)
);

CREATE INDEX IF NOT EXISTS idx_weights_sync_date ON weights (sync_id, date);
`

const upsertWeightSQL = `
INSERT INTO weights (
	sync_id,
	ts,
	date,
	weight,
	bmi,
	body_fat,
	body_water,
	bone_mass,
	metabolic_age,
	muscle_mass,
	physique_rating,
	protein_mass,
	visceral_fat,
	basal_metabolism,
	body_score,
	heart_rate,
	height,
	skeletal_muscle_mass,
	user,
	source,
	data,
	data_hash,
	created_at,
	updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(sync_id, ts) DO UPDATE SET
	date = excluded.date,
	weight = excluded.weight,
	bmi = excluded.bmi,
	body_fat = excluded.body_fat,
	body_water = excluded.body_water,
	bone_mass = excluded.bone_mass,
	metabolic_age = excluded.metabolic_age,
	muscle_mass = excluded.muscle_mass,
	physique_rating = excluded.physique_rating,
	protein_mass = excluded.protein_mass,
	visceral_fat = excluded.visceral_fat,
	basal_metabolism = excluded.basal_metabolism,
	body_score = excluded.body_score,
	heart_rate = excluded.heart_rate,
	height = excluded.height,
	skeletal_muscle_mass = excluded.skeletal_muscle_mass,
	user = excluded.user,
	source = excluded.source,
	data = excluded.data,
	data_hash = excluded.data_hash,
	updated_at = excluded.updated_at
WHERE weights.data_hash != excluded.data_hash;
`
