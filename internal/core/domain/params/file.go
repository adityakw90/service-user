package params

type FileListFilterParam struct {
	Uids       []string // file uids
	UserUid    *string
	FileType   *string
	Visibility *string
}

type FileListParam struct {
	Pagination *PaginationParam
	Filter     *FileListFilterParam
}
