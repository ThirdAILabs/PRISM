package triangulation

type Author struct {
	ID                int    `gorm:"primaryKey;column:id;autoIncrement"`
	FundCodesID       *int   `gorm:"column:fundcode_id"`
	AuthorName        string `gorm:"column:authorname;type:text"`
	NumPapersByAuthor int    `gorm:"column:numpapersbyauthor"`
}

type FundCode struct {
	ID        int    `gorm:"primaryKey;column:id;autoIncrement"`
	FundCodes string `gorm:"column:fundcodes;type:text"`
	NumPapers int    `gorm:"column:numpapers"`
}

type AuthorFundCodeResult struct {
	NumPapersByAuthor int `gorm:"column:numpapersbyauthor"`
	NumPapers         int `gorm:"column:numpapers"`
}
