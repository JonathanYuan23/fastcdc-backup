package sqlitechunks

import (
	"database/sql"
	"encoding/json"
	"io"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type ChunkSchema struct {
	checksum          string
	id, instanceCount int64
}

func OpenDB(dbFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func PopulateDB(db *sql.DB, dataSource string) error {
	fi, err := os.Open(dataSource)
	if err != nil {
		return err
	}
	defer fi.Close()

	bytes, err := io.ReadAll(fi)
	if err != nil {
		return err
	}

	rows := []ChunkSchema{}
	json.Unmarshal(bytes, &rows)

	const drop string = `
	DROP TABLE IF EXISTS chunks;
	`

	if _, err := db.Exec(drop); err != nil {
		return err
	}

	const create string = `
	CREATE TABLE IF NOT EXISTS chunks (
		id INTEGER NOT NULL PRIMARY KEY, 
		checksum TEXT NOT NULL UNIQUE,
		instance_count INTEGER NOT NULL DEFAULT 1
	);`

	if _, err := db.Exec(create); err != nil {
		return err
	}

	const insert string = `
	INSERT INTO chunks (id, checksum, instance_count) VALUES (?, ?, ?);
	`

	for _, row := range rows {
		if _, err := db.Exec(insert, row.id, row.checksum, row.instanceCount); err != nil {
			return err
		}
	}

	return nil
}

func Exists(db *sql.DB, checksum string) bool {
	const query string = `
	SELECT id FROM chunks WHERE checksum = ?;
	`
	var id int64
	err := db.QueryRow(query, checksum).Scan(&id)

	return err != sql.ErrNoRows && err == nil
}

func InsertChunk(db *sql.DB, checksum string) {
	const insert = `
	INSERT INTO chunks (instance_count) VALUES (?);
	`

	db.Exec(insert, checksum)
}

func GetCount(db *sql.DB, checksum string) int64 {
	const query = `
	SELECT instance_count FROM chunks WHERE checksum = ?;
	`

	var instances int64
	_ = db.QueryRow(query, checksum).Scan(&instances)

	return instances
}

func IncreaseCount(db *sql.DB, checksum string) {
	const update = `
	UPDATE chunks
	SET instance_count = instance_count + 1
	WHERE checksum = ?;
	`

	db.Exec(update, checksum)
}

func DecreaseCount(db *sql.DB, checksum string) {
	const update = `
	UPDATE chunks
	SET instance_count = instance_count - 1
	WHERE checksum = ?;
	`

	db.Exec(update, checksum)
}

func Delete(db *sql.DB, checksum string) {
	const delete = `
	DELETE FROM chunk WHERE checksum = ?;
	`

	db.Exec(delete, checksum)
}
