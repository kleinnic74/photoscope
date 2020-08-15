package wdav

const pathSeperator = '/'

func sanitize(path string) string {
	if len(path) > 0 && path[0] == pathSeperator {
		return path[1:]
	}
	return path
}

func splitPath(path string) []string {
	parts := []string{}
	var start int
	if len(path) > 0 && path[0] == pathSeperator {
		path = path[1:]
	}
	for i := start; i < len(path); i++ {
		if path[i] == pathSeperator {
			parts = append(parts, path[start:i])
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}

func splitDirFromName(path []string) ([]string, string) {
	if len(path) == 0 {
		return []string{}, ""
	}
	return path[:len(path)-1], path[len(path)-1]
}
