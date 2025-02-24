package triangulation

type Author struct {
    NumPapersByAuthor int `gorm:"column:numpapersbyauthor"`
}

type FundCode struct {
    NumPapers int `gorm:"column:numpapers"`
}

type AuthorFundCodeResult struct {
    NumPapersByAuthor int `gorm:"column:numpapersbyauthor"`
    NumPapers         int `gorm:"column:numpapers"`
}
