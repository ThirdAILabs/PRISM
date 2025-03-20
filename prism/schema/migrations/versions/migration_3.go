package versions

import (
	"encoding/json"
	"fmt"
	"prism/prism/api"
	"prism/prism/schema"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Migration3(db *gorm.DB) error {
	type Feedback struct {
		Id        uuid.UUID         `gorm:"type:uuid;primaryKey"`
		UserId    uuid.UUID         `gorm:"type:uuid;not null;index"`
		Flag      schema.AuthorFlag `gorm:"foreignKey:ReportId;constraint:OnDelete:CASCADE"`
		Timestamp time.Time
		Data      []byte
	}

	err := db.AutoMigrate(&Feedback{})
	if err != nil {
		return fmt.Errorf("failed to migrate Feedback: %w", err)
	}

	// port the new data byte of AuthorFlag to the DB
	var flags []schema.AuthorFlag
	if err := db.Find(&flags).Error; err != nil {
		return fmt.Errorf("failed to find AuthorFlag: %w", err)
	}

	for _, flag := range flags {
		existingFlag, err := api.ParseFlag(flag.FlagType, flag.Data)
		if err != nil {
			return fmt.Errorf("failed to parse flag: %w", err)
		}

		existingFlag.UpdateFlagHash()

		updatedData, err := json.Marshal(existingFlag)
		if err != nil {
			return fmt.Errorf("failed to serialize updated flag data for flag hash %s: %w", flag.FlagHash, err)
		}

		// Update the record in the database
		if err := db.Model(&schema.AuthorFlag{}).
			Where("report_id = ? AND flag_hash = ?", flag.ReportId, flag.FlagHash).
			Update("data", updatedData).Error; err != nil {
			return fmt.Errorf("failed to update flag data for flag hash %s: %w", flag.FlagHash, err)
		}
	}

	return nil
}

func Roolback3(db *gorm.DB) error {
	err := db.Migrator().DropTable("feedback")
	if err != nil {
		return fmt.Errorf("failed to drop table feedback: %w", err)
	}

	return nil
}
