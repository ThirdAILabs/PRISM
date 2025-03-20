package versions

import (
	"encoding/json"
	"fmt"
	"prism/prism/api"
	"prism/prism/schema"

	"gorm.io/gorm"
)

func Migration3(db *gorm.DB) error {
	if err := db.Exec(`
        CREATE TABLE IF NOT EXISTS flag_feedbacks (
            id UUID PRIMARY KEY,
            user_id UUID NOT NULL,
            report_id UUID NOT NULL,
            flag_hash CHAR(64) NOT NULL,
            timestamp TIMESTAMPTZ,
            data BYTEA
        )
    `).Error; err != nil {
		return fmt.Errorf("failed to create FlagFeedback table: %w", err)
	}

	if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_flag_feedbacks_user_id ON flag_feedbacks(user_id)
    `).Error; err != nil {
		return fmt.Errorf("failed to create index on FlagFeedback.user_id: %w", err)
	}

	if err := db.Exec(`
        ALTER TABLE flag_feedbacks 
        ADD CONSTRAINT fk_flag_feedbacks_flag 
        FOREIGN KEY (report_id, flag_hash) 
        REFERENCES author_flags(report_id, flag_hash) 
        ON DELETE CASCADE
    `).Error; err != nil {
		return fmt.Errorf("failed to add foreign key constraint to FlagFeedback: %w", err)
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
			return fmt.Errorf("failed to update flag data for report %s, flag hash %s: %w", flag.ReportId, flag.FlagHash, err)
		}
	}

	return nil
}

func Rollback3(db *gorm.DB) error {
	err := db.Migrator().DropTable("feedback")
	if err != nil {
		return fmt.Errorf("failed to drop table feedback: %w", err)
	}

	return nil
}
