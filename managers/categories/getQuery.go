package categories

type GetQuery struct {
	ReqId      string
	Id         string
	Name       string
	SearchName string
	FromDate   int64
	ToDate     int64
	Limit      int
}
