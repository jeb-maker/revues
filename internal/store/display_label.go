package store

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// RunDisplayLabel formats a review label from template, subject, and creation time.
// When runID is non-zero, appends " · #id" for disambiguation.
func RunDisplayLabel(templateName, subjectName, createdAt string, runID int64) string {
	templateName = strings.TrimSpace(templateName)
	subjectName = strings.TrimSpace(subjectName)
	dateLabel := formatRunDisplayDate(createdAt)

	parts := make([]string, 0, 4)
	if templateName != "" {
		parts = append(parts, templateName)
	}
	if subjectName != "" {
		parts = append(parts, subjectName)
	}
	if dateLabel != "" {
		parts = append(parts, dateLabel)
	}
	label := strings.Join(parts, " · ")
	if label == "" && runID > 0 {
		return fmt.Sprintf("#%d", runID)
	}
	if runID > 0 {
		label += fmt.Sprintf(" · #%d", runID)
	}
	return label
}

func formatRunDisplayDate(createdAt string) string {
	createdAt = strings.TrimSpace(createdAt)
	if createdAt == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		t, err = time.Parse("2006-01-02", createdAt)
		if err != nil {
			return createdAt
		}
	}
	return t.UTC().Format("02/01/2006")
}

// RunDisplayLabelForRun loads subject and template metadata for a run label.
func (s *Store) RunDisplayLabelForRun(ctx context.Context, run *ChecklistRun) (string, error) {
	subject, err := s.SubjectByID(ctx, run.SubjectID)
	if err != nil {
		return "", fmt.Errorf("load subject for run label: %w", err)
	}
	versionInfo, err := s.TemplateVersionInfo(ctx, run.TemplateVersionID)
	if err != nil {
		return RunDisplayLabel("", subject.Name, run.CreatedAt, run.ID), nil
	}
	return RunDisplayLabel(versionInfo.Name, subject.Name, run.CreatedAt, run.ID), nil
}
