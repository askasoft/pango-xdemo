package admin

var pagerLimits = []int{20, 50, 100}

type ArgIDs struct {
	IDs []int64 `form:"id[]"`
}
