package param

type UserFileListFilterParam struct {
	Uids       []string // file uids
	UserUid    []string
	FileType   *string
	Visibility *string
}

type UserFileListParam struct {
	Pagination *PaginationParam
	Filter     *UserFileListFilterParam
}

type UserFileCreateParam struct {
	UserUID    string
	FileType   string
	FileName   string
	FilePath   string
	MimeType   string
	SizeBytes  int64
	Visibility string
}

type UserFileUpdateParam struct {
	FileName   *string
	FilePath   *string
	MimeType   *string
	SizeBytes  *int64
	Visibility *string
}
