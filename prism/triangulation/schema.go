package triangulation

type Author struct {
	UniqueID          int    `gorm:"primaryKey;column:uniqueid;autoIncrement"`
	FundCodesID       *int   `gorm:"column:fundcodes_id"`
	FundCodeHash      string `gorm:"column:fund_code_hash;type:varchar(64)"`
	AuthorName        string `gorm:"column:authorname;type:text"`
	NumPapersByAuthor int    `gorm:"column:numpapersbyauthor"`
}

func (Author) TableName() string {
	return "authors"
}

type FundCode struct {
	ID        int    `gorm:"primaryKey;column:id;autoIncrement"`
	FundCodes string `gorm:"column:fundcodes;type:text"`
	NumPapers int    `gorm:"column:numpapers"`
}

func (FundCode) TableName() string {
	return "fundcodes"
}

type AuthorFundCodeResult struct {
	NumPapersByAuthor int `gorm:"column:numpapersbyauthor"`
	NumPapers         int `gorm:"column:numpapers"`
}
