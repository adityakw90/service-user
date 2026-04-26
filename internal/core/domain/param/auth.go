package param

type AuthParams struct {
	Identifier        string
	IdentifierType    string
	Password          string
	DeviceFingerprint *string
	DeviceName        *string
	DeviceIP          *string
	Extra             *map[string]any
}
