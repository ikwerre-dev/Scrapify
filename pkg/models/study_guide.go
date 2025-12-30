package models

type StudyGuide struct {
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	KeyPoints   []string `json:"key_points"`
	Glossary    []map[string]string `json:"glossary"`
	Timeline    []TimelineEntry `json:"timeline"`
}

type TimelineEntry struct {
	Timestamp string `json:"timestamp"`
	Event     string `json:"event"`
}

type AnalysisResult struct {
	Transcription []TimelineEntry `json:"transcription"`
	StudyGuide    StudyGuide      `json:"study_guide"`
	ImagePath     string          `json:"image_path"`
}
