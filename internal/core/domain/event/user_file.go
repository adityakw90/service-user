package event

// EventUserFileCreatedData is emitted when a user file is created.
type EventUserFileCreatedData struct {
	UserUID  string
	FileName string
}

// EventUserFileUpdatedData is emitted when a user file is updated.
type EventUserFileUpdatedData struct {
	UserUID string
}

// EventUserFileDeletedData is emitted when a user file is deleted.
type EventUserFileDeletedData struct {
	UserUID string
}
