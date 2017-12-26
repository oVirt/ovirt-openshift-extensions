package internal

type Status string

const (
	Success      Status = "Success"
	Failure      Status = "Failure"
	NotSupported Status = "Not supported"
)

var FailedResponse = Response{Status: Failure}
var SuccessfulResponse = Response{Status: Success}
var NotSupportedResponse = Response{Status: NotSupported}
var SuccesfullResonseJson = `{ "status": "Success", "message": ""}`
var FailedResponseJson = `{ "status": "Failure", "message": ""}`

type Response struct {
	Status     Status //`json:"status"`     //"status": "<Success/Failure/Not supported>",
	Message    string //`json:"message"`    //"message": "<Reason for success/failure>",
	Device     string //`json:"device"`     //"device": "<Path to the device attached. This field is valid only for attach & waitforattach call-outs>"
	VolumeName string //`json:"volumeName"` //"volumeName": "<Cluster wide unique name of the volume. Valid only for getvolumename call-out>"
	Attached   string //`json:"attached"`   //"attached": <True/False (Return true if volume is attached on the node. Valid only for isattached call-out)>
	Capabilities
}

type Capabilities struct {
	//"capabilities": <Only included as part of the Init response>
	Attach string //: <True/False (Return true if the driver implements attach and detach)>
}

type AttchResponse struct {
	Response
}

type AttachRequest struct {
	StorageDomain string `json:"oVirtStorageDomain"`
	VolumeName    string `json:"oVirtVolumeName"`
	FsType        string `json:"kubernetes.io/fsType"`
	Mode          string `json:""kubernetes.io/readwrite"`
	// TODO use k8s secret?
	Secret string `json:""kubernetes.io/secret"`
}
