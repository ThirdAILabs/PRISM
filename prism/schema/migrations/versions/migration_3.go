package versions

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"prism/prism/api"
	"prism/prism/schema"

	"gorm.io/gorm"
)

func UpdateHash(ftype string, flag api.Flag) (api.Flag, error) {
	switch ftype {
	case api.TalentContractType:
		temp, ok := flag.(*api.TalentContractFlag)
		if !ok {
			return nil, fmt.Errorf("invalid flag type for %s", ftype)
		}
		hashBytes := temp.CalculateHash()
		temp.Hash = hex.EncodeToString(hashBytes[:])
		return temp, nil

	case api.AssociationsWithDeniedEntityType:
		temp, ok := flag.(*api.AssociationWithDeniedEntityFlag)
		if !ok {
			return nil, fmt.Errorf("invalid flag type for %s", ftype)
		}
		hashBytes := temp.CalculateHash()
		temp.Hash = hex.EncodeToString(hashBytes[:])
		return temp, nil

	case api.HighRiskFunderType:
		temp, ok := flag.(*api.HighRiskFunderFlag)
		if !ok {
			return nil, fmt.Errorf("invalid flag type for %s", ftype)
		}
		hashBytes := temp.CalculateHash()
		temp.Hash = hex.EncodeToString(hashBytes[:])
		return temp, nil

	case api.AuthorAffiliationType:
		temp, ok := flag.(*api.AuthorAffiliationFlag)
		if !ok {
			return nil, fmt.Errorf("invalid flag type for %s", ftype)
		}
		hashBytes := temp.CalculateHash()
		temp.Hash = hex.EncodeToString(hashBytes[:])
		return temp, nil

	case api.PotentialAuthorAffiliationType:
		temp, ok := flag.(*api.PotentialAuthorAffiliationFlag)
		if !ok {
			return nil, fmt.Errorf("invalid flag type for %s", ftype)
		}
		hashBytes := temp.CalculateHash()
		temp.Hash = hex.EncodeToString(hashBytes[:])
		return temp, nil

	case api.MiscHighRiskAssociationType:
		temp, ok := flag.(*api.MiscHighRiskAssociationFlag)
		if !ok {
			return nil, fmt.Errorf("invalid flag type for %s", ftype)
		}
		hashBytes := temp.CalculateHash()
		temp.Hash = hex.EncodeToString(hashBytes[:])
		return temp, nil

	case api.CoauthorAffiliationType:
		temp, ok := flag.(*api.CoauthorAffiliationFlag)
		if !ok {
			return nil, fmt.Errorf("invalid flag type for %s", ftype)
		}
		hashBytes := temp.CalculateHash()
		temp.Hash = hex.EncodeToString(hashBytes[:])
		return temp, nil

	case api.MultipleAffiliationType:
		temp, ok := flag.(*api.MultipleAffiliationFlag)
		if !ok {
			return nil, fmt.Errorf("invalid flag type for %s", ftype)
		}
		hashBytes := temp.CalculateHash()
		temp.Hash = hex.EncodeToString(hashBytes[:])
		return temp, nil

	case api.HighRiskPublisherType:
		temp, ok := flag.(*api.HighRiskPublisherFlag)
		if !ok {
			return nil, fmt.Errorf("invalid flag type for %s", ftype)
		}
		hashBytes := temp.CalculateHash()
		temp.Hash = hex.EncodeToString(hashBytes[:])
		return temp, nil

	case api.HighRiskCoauthorType:
		temp, ok := flag.(*api.HighRiskCoauthorFlag)
		if !ok {
			return nil, fmt.Errorf("invalid flag type for %s", ftype)
		}
		hashBytes := temp.CalculateHash()
		temp.Hash = hex.EncodeToString(hashBytes[:])
		return temp, nil
	}

	return nil, fmt.Errorf("unknown flag type %s", ftype)
}

func Migration3(db *gorm.DB) error {
	err := db.AutoMigrate(&schema.FlagFeedback{}, &schema.AuthorFlag{})
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

		updatedFlag, err := UpdateHash(flag.FlagType, existingFlag)
		if err != nil {
			return fmt.Errorf("failed to update flag hash: %w", err)
		}
		updatedData, err := json.Marshal(updatedFlag)
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
