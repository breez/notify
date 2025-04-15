package notification

type Notification struct {
	Template         string
	DisplayMessage   string
	Type             string
	TargetIdentifier string
	AppData          *string
	Data             map[string]interface{}
}
