package triangulation

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

type TriangulationDB struct {
	db *gorm.DB
}

func (t *TriangulationDB) GetAuthorFundCodeResult(authorName, fundCode string) (*AuthorFundCodeResult, error) {
	var result AuthorFundCodeResult

	err := t.db.Table("authors a").
		Select("a.numpapersbyauthor, f.numpapers").
		Joins("JOIN fundcodes f ON a.fundcode_id = f.id").
		Where("a.authorname = ? AND f.fundcode = ?", authorName, fundCode).
		Limit(1).
		Scan(&result).Error

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

func CreateTriangulationDB(db *gorm.DB) *TriangulationDB {
	return &TriangulationDB{db: db}
}
