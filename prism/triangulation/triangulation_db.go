package triangulation

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

type TriangulationDB struct {
	db *gorm.DB
}

func CreateTriangulationDB(db *gorm.DB) *TriangulationDB {
	return &TriangulationDB{db: db}
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
		return nil, fmt.Errorf("error executing fund codd triangulation query: %w", err)
	}

	if result.NumPapersByAuthor == 0 && result.NumPapers == 0 {
		return nil, nil
	}

	return &result, nil
}

func (t *TriangulationDB) IsAuthorGrantRecipient(authorName string, grantNumber string) (bool, error) {
	result, err := t.GetAuthorFundCodeResult(authorName, grantNumber)

	if err != nil {
		return false, err
	}

	if result != nil {
		if float64(result.NumPapersByAuthor)/float64(result.NumPapers) >= 0.4 {
			return true, nil
		}
	}

	return false, nil
}
