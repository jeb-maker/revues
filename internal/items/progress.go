package items

import "github.com/jeb-maker/revues/internal/store"

// Progress counts completed run items (ok or na) and the total item count.
func Progress(runItems []store.RunItem) (done, total int) {
	total = len(runItems)
	for _, item := range runItems {
		if item.Status == StatusOK || item.Status == StatusNA {
			done++
		}
	}
	return done, total
}
