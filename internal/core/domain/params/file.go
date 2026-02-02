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

type FileCreateParam struct {
	UserUID    string
	FileType   string
	FileName   string
	FilePath   string
	MimeType   string
	SizeBytes  int64
	Visibility string
}

type FileUpdateParam struct {
	FileName   *string
	FilePath   *string
	MimeType   *string
	SizeBytes  *int64
	Visibility *string
}
