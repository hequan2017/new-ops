// Package service 白泽 Agent 通道：K8s 集群内接入 YAML 渲染（K4 一键接入）
// SA + 只读 ClusterRole/Binding + Deployment（自动挂载 SA token），serverAddr/token 由调用方注入。
package service

import (
	"fmt"
	"strings"
)

// DefaultK8sAgentImage 集群内 agent 默认镜像
const DefaultK8sAgentImage = "hequan2017/baize-agent:0.2.0"

// RenderK8sAgentDeployment 渲染集群内 agent 一键接入 YAML（纯逻辑）
// serverAddr 为 server 基地址（http(s)://）；image 空取默认；token 为本次签发的接入令牌明文。
func RenderK8sAgentDeployment(serverAddr, image, token string) (string, error) {
	sa := strings.TrimRight(strings.TrimSpace(serverAddr), "/")
	if !strings.HasPrefix(sa, "http://") && !strings.HasPrefix(sa, "https://") {
		return "", newAgentErr(ErrCodeAgentParamInvalid, "serverAddr 必须以 http(s):// 开头")
	}
	if strings.TrimSpace(token) == "" {
		return "", newAgentErr(ErrCodeAgentParamInvalid, "token 不能为空")
	}
	if strings.TrimSpace(image) == "" {
		image = DefaultK8sAgentImage
	}
	return fmt.Sprintf(`# 白泽集群内 Agent 一键接入（K4）：kubectl apply -f baize-agent.yaml
# agent 以 ServiceAccount（只读）身份把 K8s API 请求经 WS 通道代理回 server
apiVersion: v1
kind: ServiceAccount
metadata:
  name: baize-agent
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: baize-agent-readonly
rules:
  - apiGroups: ["*"]
    resources: ["*"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: baize-agent-readonly
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: baize-agent-readonly
subjects:
  - kind: ServiceAccount
    name: baize-agent
    namespace: default
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: baize-agent
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: baize-agent
  template:
    metadata:
      labels:
        app: baize-agent
    spec:
      serviceAccountName: baize-agent
      containers:
        - name: agent
          image: %s
          args:
            - -server
            - %s
            - -token
            - %s
`, image, sa, token), nil
}
