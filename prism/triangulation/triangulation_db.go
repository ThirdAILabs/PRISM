package triangulation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

var triangulationDB *gorm.DB

func SetTriangulationDB(db *gorm.DB) {
	triangulationDB = db
}

func GetTriangulationDB() *gorm.DB {
	return triangulationDB
}

func GetAuthorFundCodeResult(db *gorm.DB, authorName, fundCode string) (*AuthorFundCodeResult, error) {
	hash := sha256.New()
	hash.Write([]byte(fundCode))
	fundCodeHash := hex.EncodeToString(hash.Sum(nil))

	var result AuthorFundCodeResult

	err := db.Raw(`
        SELECT a.numpapersbyauthor, f.numpapers
        FROM authors a
        JOIN fundcodes f 
        ON a.fundcodes_id = f.id
        WHERE a.authorname = ? AND a.fund_code_hash = ?
        LIMIT 1
    `, authorName, fundCodeHash).Scan(&result).Error
	if err != nil {
		slog.Error("error executing fundcode triangulation query", "error", err)
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	if result.NumPapersByAuthor == 0 && result.NumPapers == 0 {
		slog.Info("not found author and number of papers with triangulation query", "author_name", authorName, "fund_code", fundCode)
		return nil, nil
	}

	return &result, nil
}
