package signal

import "testing"

func TestDomain_Signal_UserFileSignal_Creation(t *testing.T) {
	uid := "file-123"
	fileName := "profile.jpg"
	fileSize := int64(1024)

	sig := UserFileSignal{
		UID:       &uid,
		FileName:  &fileName,
		FileSize:  &fileSize,
		Operation: "get",
	}

	if sig.UID == nil {
		t.Error("UID should not be nil")
	}
	if *sig.UID != uid {
		t.Errorf("UID = %s, want %s", *sig.UID, uid)
	}
}

func TestDomain_Signal_UserFileSignal_AllFieldsNil(t *testing.T) {
	sig := UserFileSignal{
		Operation: "list",
	}

	if sig.UID != nil {
		t.Error("UID should be nil")
	}
}
