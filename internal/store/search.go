package store

import (
	"context"
	"fmt"
	"strings"
)

// Search result kinds — alignés sur OpenAPI SearchResultKind.
const (
	SearchKindSubject  = "subject"
	SearchKindRun      = "run"
	SearchKindTemplate = "template"
	SearchKindTask     = "task"
)

const (
	DefaultSearchLimitPerKind = 10
	MaxSearchLimitPerKind     = 25
)

// SearchHit is one row of GET /api/v1/search.
type SearchHit struct {
	Kind     string
	ID       int64
	Title    string
	Subtitle string
	Href     string
}

// GlobalSearchOpts configures GlobalSearch.
type GlobalSearchOpts struct {
	UserID           int64
	Admin            bool
	Query            string
	LimitPerKind     int
	IncludeTemplates bool
	IncludeTasks     bool
}

// GlobalSearch aggregates subjects, runs, templates and assigned tasks with the
// same visibility rules as the corresponding list endpoints (≤ 4 queries).
func (s *Store) GlobalSearch(ctx context.Context, opts GlobalSearchOpts) ([]SearchHit, map[string]int, error) {
	q := strings.TrimSpace(opts.Query)
	if q == "" {
		return nil, nil, fmt.Errorf("search query required")
	}
	limit := opts.LimitPerKind
	if limit <= 0 {
		limit = DefaultSearchLimitPerKind
	}
	if limit > MaxSearchLimitPerKind {
		limit = MaxSearchLimitPerKind
	}

	totals := map[string]int{}
	var results []SearchHit

	subjects, err := s.ListSubjects(ctx, opts.UserID, opts.Admin, q)
	if err != nil {
		return nil, nil, fmt.Errorf("search subjects: %w", err)
	}
	totals[SearchKindSubject] = len(subjects)
	for i, sub := range subjects {
		if i >= limit {
			break
		}
		hit := SearchHit{
			Kind:  SearchKindSubject,
			ID:    sub.ID,
			Title: sub.Name,
			Href:  fmt.Sprintf("/subjects/%d", sub.ID),
		}
		if desc := strings.TrimSpace(sub.Description); desc != "" {
			hit.Subtitle = desc
		}
		results = append(results, hit)
	}

	runs, totalRuns, err := s.ListFilteredRunSummaries(ctx, opts.UserID, opts.Admin, "", q, limit, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("search runs: %w", err)
	}
	totals[SearchKindRun] = totalRuns
	for _, run := range runs {
		hit := SearchHit{
			Kind:  SearchKindRun,
			ID:    run.RunID,
			Title: run.Title,
			Href:  fmt.Sprintf("/runs/%d", run.RunID),
		}
		if name := strings.TrimSpace(run.SubjectName); name != "" {
			hit.Subtitle = name
		}
		results = append(results, hit)
	}

	if opts.IncludeTemplates {
		templates, tplErr := s.ListTemplateIndex(ctx, opts.UserID, opts.Admin, q)
		if tplErr != nil {
			return nil, nil, fmt.Errorf("search templates: %w", tplErr)
		}
		totals[SearchKindTemplate] = len(templates)
		for i, tpl := range templates {
			if i >= limit {
				break
			}
			hit := SearchHit{
				Kind:  SearchKindTemplate,
				ID:    tpl.ID,
				Title: tpl.Name,
				Href:  fmt.Sprintf("/modeles/%d", tpl.ID),
			}
			if len(tpl.Tags) > 0 {
				hit.Subtitle = strings.Join(tpl.Tags, ", ")
			}
			results = append(results, hit)
		}
	}

	if opts.IncludeTasks {
		tasks, taskErr := s.ListAssignedRunItems(ctx, opts.UserID, "", q)
		if taskErr != nil {
			return nil, nil, fmt.Errorf("search tasks: %w", taskErr)
		}
		totals[SearchKindTask] = len(tasks)
		for i, task := range tasks {
			if i >= limit {
				break
			}
			hit := SearchHit{
				Kind:  SearchKindTask,
				ID:    task.ID,
				Title: task.Label,
				Href:  fmt.Sprintf("/runs/%d/items/%d", task.RunID, task.ID),
			}
			parts := make([]string, 0, 2)
			if name := strings.TrimSpace(task.SubjectName); name != "" {
				parts = append(parts, name)
			}
			if sec := strings.TrimSpace(task.Section); sec != "" {
				parts = append(parts, sec)
			}
			if len(parts) > 0 {
				hit.Subtitle = strings.Join(parts, " · ")
			}
			results = append(results, hit)
		}
	}

	if results == nil {
		results = []SearchHit{}
	}
	return results, totals, nil
}

func searchTerms(query string) []string {
	raw := strings.Fields(strings.TrimSpace(query))
	if len(raw) == 0 {
		return nil
	}
	return raw
}

func likeContainsPattern(term string) string {
	term = strings.ReplaceAll(term, `\`, `\\`)
	term = strings.ReplaceAll(term, `%`, `\%`)
	term = strings.ReplaceAll(term, `_`, `\_`)
	return "%" + term + "%"
}
