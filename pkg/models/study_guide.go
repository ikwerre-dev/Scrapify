package models

type StudyGuide struct {
	Title             string              `json:"title"`
	Summary           string              `json:"summary"`
	KeyPoints         []string            `json:"key_points"`
	Glossary          []map[string]string `json:"glossary"`
	Timeline          []TimelineEntry     `json:"timeline"`
	VisualAnalysis    []VisualBit         `json:"visual_analysis"`
	ExternalResources []ExternalResource  `json:"external_resources"`
	Quiz              []QuizQuestion      `json:"quiz"`
}

type VisualBit struct {
	GridIndex   int    `json:"grid_index"`
	ItemIndex   int    `json:"item_index"`
	Timestamp   string `json:"timestamp"`
	Explanation string `json:"explanation"`
}

type ExternalResource struct {
	Type  string `json:"type"` // "video", "article", "documentation"
	Title string `json:"title"`
	URL   string `json:"url"`
}

type QuizQuestion struct {
	Question string   `json:"question"`
	Options  []string `json:"options,omitempty"`
	Answer   string   `json:"answer"`
}

type TimelineEntry struct {
	Timestamp string `json:"timestamp"`
	Event     string `json:"event"`
}

type AnalysisResult struct {
	Transcription []TimelineEntry `json:"transcription"`
	StudyGuide    StudyGuide      `json:"study_guide"`
	ImagePaths    []string        `json:"image_paths"`
}
