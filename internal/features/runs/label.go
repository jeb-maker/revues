package runs

import "github.com/jeb-maker/revues/internal/store"

// RunDisplayLabel formats "{Modèle} · {Sujet} · {date}" with optional " · #id".
func RunDisplayLabel(templateName, subjectName, createdAt string, runID int64) string {
	return store.RunDisplayLabel(templateName, subjectName, createdAt, runID)
}
