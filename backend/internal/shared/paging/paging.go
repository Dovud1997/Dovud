package paging

func Normalize(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}

func Offset(page, perPage int) int {
	return (page - 1) * perPage
}
