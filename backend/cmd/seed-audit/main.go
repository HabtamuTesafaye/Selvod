package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Priority: Environment Variable -> Auto-discovery -> Default
	dataDir := os.Getenv("SV_DATA_DIR")
	if dataDir == "" {
		if _, err := os.Stat("data"); err == nil {
			dataDir = "data"
		} else if _, err := os.Stat("../data"); err == nil {
			dataDir = "../data"
		} else {
			dataDir = "data"
		}
	}

	dbPath := filepath.Join(dataDir, "selvod.db")
	absDbPath, _ := filepath.Abs(dbPath)
	
	fmt.Printf("Seeding database at: %s\n", absDbPath)

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Println("Failed to create data dir:", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Println("Failed to open DB:", err)
		os.Exit(1)
	}
	defer db.Close()

	query := `
	CREATE TABLE IF NOT EXISTS videos (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		original_ext TEXT NOT NULL,
		status TEXT NOT NULL,
		upload_size_bytes BIGINT NOT NULL,
		total_size_bytes BIGINT DEFAULT 0,
		duration INTEGER DEFAULT 0,
		transcoding_started_at DATETIME,
		transcoding_finished_at DATETIME,
		error_message TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Failed to create table:", err)
		os.Exit(1)
	}

	_, err = db.Exec(`INSERT OR REPLACE INTO videos (id, title, original_ext, status, upload_size_bytes, created_at, updated_at) 
		VALUES ('audit-vid', 'Audit Test Video', '.mp4', 'completed', 1024, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if err != nil {
		fmt.Println("Failed to insert video:", err)
		os.Exit(1)
	}

	hlsBase := filepath.Join(dataDir, "hls/audit-vid")
	hlsDir := filepath.Join(hlsBase, "0")
	os.MkdirAll(hlsDir, 0755)
	
	os.WriteFile(filepath.Join(hlsBase, "master.m3u8"), []byte("#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1280000\n0/index.m3u8\n"), 0644)
	os.WriteFile(filepath.Join(hlsDir, "index.m3u8"), []byte("#EXTM3U\n#EXT-X-TARGETDURATION:10\n#EXTINF:10,\n001.ts\n#EXT-X-ENDLIST\n"), 0644)
	os.WriteFile(filepath.Join(hlsDir, "001.ts"), []byte("dummy binary segment data"), 0644)

	fmt.Println("Audit seed completed successfully.")
}
