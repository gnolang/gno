package acl

// Helper function to append without duplicates
func appendGroupsWithoutDuplicates(slice []string, items []string) []string {
	seen := make(map[string]struct{})
	for _, v := range slice {
		seen[v] = struct{}{}
	}
	for _, item := range items {
		if _, exists := seen[item]; !exists {
			slice = append(slice, item)
			seen[item] = struct{}{}
		}
	}
	return slice
}
