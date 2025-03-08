package triangulation

type Author struct {
	ID                int    `gorm:"primaryKey;column:id;autoIncrement"`
	FundCodeID        *int   `gorm:"column:fundcode_id"`
	AuthorName        string `gorm:"column:authorname;type:text"`
	NumPapersByAuthor int    `gorm:"column:numpapersbyauthor"`
}

func (Author) TableName() string {
	return "authors"
}

type FundCode struct {
	ID        int    `gorm:"primaryKey;column:id;autoIncrement"`
	FundCode  string `gorm:"column:fundcode;type:text"`
	NumPapers int    `gorm:"column:numpapers"`
}

func (FundCode) TableName() string {
	return "fundcodes"
}

type AuthorFundCodeResult struct {
	NumPapersByAuthor int `gorm:"column:numpapersbyauthor"`
	NumPapers         int `gorm:"column:numpapers"`
}
