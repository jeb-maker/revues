package auth

const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleReader = "reader"
)

var roleRank = map[string]int{
	RoleReader: 1,
	RoleEditor: 2,
	RoleAdmin:  3,
}

// HasMinRole reports whether userRole meets or exceeds minRole in the global hierarchy.
func HasMinRole(userRole, minRole string) bool {
	userRank, ok := roleRank[userRole]
	if !ok {
		return false
	}
	minRank, ok := roleRank[minRole]
	if !ok {
		return false
	}
	return userRank >= minRank
}

// ValidRole reports whether role is a known global role.
func ValidRole(role string) bool {
	_, ok := roleRank[role]
	return ok
}
