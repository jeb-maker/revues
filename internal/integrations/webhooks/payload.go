package webhooks

type Envelope struct {
	EventID    string `json:"event_id"`
	EventType  string `json:"event_type"`
	OccurredAt string `json:"occurred_at"`
	Data       any    `json:"data"`
}

type ReviewCompletedData struct {
	Review ReviewRef    `json:"review"`
	Items  ItemsSummary `json:"items"`
}

type ReviewItemNOKData struct {
	Review ReviewRef `json:"review"`
	Item   ItemRef   `json:"item"`
}

type TestData struct {
	Message string `json:"message"`
}

type ReviewRef struct {
	ID           int64  `json:"id"`
	DisplayLabel string `json:"display_label"`
	Status       string `json:"status"`
	SubjectID    int64  `json:"subject_id"`
	SubjectName  string `json:"subject_name"`
	ClosingNote  string `json:"closing_note,omitempty"`
	CompletedAt  string `json:"completed_at,omitempty"`
}

type ItemRef struct {
	ID      int64  `json:"id"`
	Section string `json:"section"`
	Label   string `json:"label"`
	Status  string `json:"status"`
	Comment string `json:"comment"`
}

type ItemsSummary struct {
	Total   int `json:"total"`
	OK      int `json:"ok"`
	NOK     int `json:"nok"`
	NA      int `json:"na"`
	Pending int `json:"pending"`
}
