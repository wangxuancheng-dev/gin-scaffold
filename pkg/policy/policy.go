// Package policy 提供与 RBAC 正交的轻量策略辅助（如资源归属）。
package policy

// SameUser 判断操作者是否为资源所有者（用于「只能改自己」等场景）。
func SameUser(actorID, resourceOwnerID int64) bool {
	return actorID > 0 && resourceOwnerID > 0 && actorID == resourceOwnerID
}
