package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"scrapify/pkg/models"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiService struct {
	client *genai.Client
	ctx    context.Context
}

func NewGeminiService(apiKey string) (*GeminiService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &GeminiService{client: client, ctx: ctx}, nil
}

func (s *GeminiService) TranscribeAudio(audioPath string) ([]models.TimelineEntry, error) {
	model := s.client.GenerativeModel("gemini-2.5-flash-lite")

	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, err
	}

	prompt := []genai.Part{
		genai.Text("Transcribe this audio file into a JSON array of objects with 'timestamp' (MM:SS) and 'event' (the transcribed text) fields. Return ONLY the JSON array."),
		genai.Blob{MIMEType: "audio/mpeg", Data: audioData},
	}

	resp, err := model.GenerateContent(s.ctx, prompt...)
	if err != nil {
		return nil, err
	}

	var transcription []models.TimelineEntry
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				cleanJSON := strings.TrimSpace(string(txt))
				cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
				cleanJSON = strings.TrimSuffix(cleanJSON, "```")
				err := json.Unmarshal([]byte(cleanJSON), &transcription)
				if err != nil {
					return nil, fmt.Errorf("failed to parse transcription JSON: %v, content: %s", err, cleanJSON)
				}
			}
		}
	}

	return transcription, nil
}

func (s *GeminiService) GenerateStudyGuide(transcription []models.TimelineEntry, gridImagePaths []string) (*models.StudyGuide, error) {
	model := s.client.GenerativeModel("gemini-2.5-flash-lite")

	transcriptionJSON, _ := json.MarshalIndent(transcription, "", "  ")

	prompt := []genai.Part{
		genai.Text(fmt.Sprintf(`As an elite educational architect, analyze the attached video frames and the provided transcription to create an ultra-advanced study guide.

TRANSCRIPTION & TIMELINE:
%s

OUTPUT REQUIREMENTS (Strict JSON):
1. **Title & Summary**: High-level context of the material.
2. **Visual Bit-by-Bit Analysis**: For each image grid (5x4 layout, 20 items per grid), explain the key visual information and process shown in EACH snapshot. Reference them by grid index and item index (1-20).
3. **External Resources**: Based on the content, provide 3-5 high-quality YouTube tutorial links or web documentation links that cover similar concepts.
4. **Pop Quiz**: Generate a 5-question interactive quiz (multiple choice) based on the video content to test comprehension.
5. **Full Guide Synthesis**: Integrate transcribed concepts with visual proof.

JSON STRUCTURE:
{
  "title": "...",
  "summary": "...",
  "key_points": ["..."],
  "glossary": [{"term": "...", "definition": "..."}],
  "timeline": [{"timestamp": "...", "event": "..."}],
  "visual_analysis": [{"grid_index": 1, "item_index": 1, "timestamp": "...", "explanation": "..."}],
  "external_resources": [{"type": "video|article", "title": "...", "url": "..."}],
  "quiz": [{"question": "...", "options": ["...", "..."], "answer": "..."}]
}
`, string(transcriptionJSON))),
	}

	// Add all grid images to the prompt
	for _, path := range gridImagePaths {
		gridData, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read grid image %s: %v", path, err)
		}
		prompt = append(prompt, genai.Blob{MIMEType: "image/jpeg", Data: gridData})
	}

	resp, err := model.GenerateContent(s.ctx, prompt...)
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %v", err)
	}

	var guide models.StudyGuide
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				cleanJSON := strings.TrimSpace(string(txt))
				cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
				cleanJSON = strings.TrimSuffix(cleanJSON, "```")
				err := json.Unmarshal([]byte(cleanJSON), &guide)
				if err != nil {
					return nil, fmt.Errorf("failed to parse study guide JSON: %v, content: %s", err, cleanJSON)
				}
			}
		}
	}

	return &guide, nil
}
