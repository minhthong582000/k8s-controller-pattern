package common

const (
	MetadataPrefix = "thongdepzai.cloud"
	ControllerName = "gitops-controller"
)

var (
	LabelKeyAppInstance = MetadataPrefix + "/app-instance"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Tunnel is synced
	SuccessSynced = "Synced"

	// MessageResourceSynced is the message used for an Event fired when a Tunnel
	// is synced successfully
	MessageResourceSynced = "Tunnel synced successfully"
)
