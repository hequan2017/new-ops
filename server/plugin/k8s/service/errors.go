package service

// k8s 插件错误码段：1500-1599（DEV_PLAN 3.6）
const (
	ErrCodeClusterNameRequired = 1501
	ErrCodeClusterDuplicate    = 1502
	ErrCodeKubeconfigInvalid   = 1503
	ErrCodeClusterNotFound     = 1504
	// 1505-1507 见 k8s_write.go（ReplicasInvalid/NamespaceReq/WorkloadNotFound）
	ErrCodeNodeNotFound         = 1508
	ErrCodeKindUnsupported      = 1509
	ErrCodeYAMLIdentityMismatch = 1510
	ErrCodeYAMLEmpty            = 1511
	ErrCodeRepoURLInvalid       = 1512
)

// K8sError k8s 插件错误
type K8sError struct {
	Code int
	Msg  string
}

func (e *K8sError) Error() string { return e.Msg }

func newK8sErr(code int, msg string) *K8sError {
	return &K8sError{Code: code, Msg: msg}
}
