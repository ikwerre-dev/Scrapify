package service

import (
	"log"
	"path/filepath"
	"scrapify/pkg/models"
	"time"
)

type WorkerService struct {
	taskChan      chan *models.Task
	mediaService  *MediaService
	imageService  *ImageService
	geminiService *GeminiService
	storage       *StorageService
}

func NewWorkerService(ms *MediaService, is *ImageService, gs *GeminiService, ss *StorageService, bufferSize int) *WorkerService {
	return &WorkerService{
		taskChan:      make(chan *models.Task, bufferSize),
		mediaService:  ms,
		imageService:  is,
		geminiService: gs,
		storage:       ss,
	}
}

func (s *WorkerService) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go s.worker(i)
	}
}

func (s *WorkerService) Enqueue(task *models.Task) {
	s.taskChan <- task
}

func (s *WorkerService) worker(id int) {
	for task := range s.taskChan {
		log.Printf("Worker %d processing task: %s", id, task.ID)
		s.processTask(task)
	}
}

func (s *WorkerService) processTask(task *models.Task) {
	task.Status = models.StatusProcessing
	task.UpdatedAt = time.Now()
	startTime := time.Now()
	s.storage.SaveTask(task)

	addStage := func(name string) *models.Stage {
		stage := &models.Stage{
			Name:      name,
			Status:    models.StatusProcessing,
			StartTime: time.Now(),
		}
		task.Stages = append(task.Stages, stage)
		s.storage.SaveTask(task)
		return stage
	}

	finishStage := func(stage *models.Stage, err error) {
		stage.Duration = time.Since(stage.StartTime)
		if err != nil {
			stage.Status = models.StatusFailed
			stage.Error = err.Error()
			task.Status = models.StatusFailed
		} else {
			stage.Status = models.StatusCompleted
		}
		task.UpdatedAt = time.Now()
		s.storage.SaveTask(task)
	}

	uploadPath := filepath.Join("uploads", task.StoredName)

	// 1. Media Processing (Parallel Snapshots & Audio)
	mediaStage := addStage("Media Processing")

	errChan := make(chan error, 2)
	var snapshots []string
	snapshotDir := filepath.Join("snapshots", task.ID)

	go func() {
		var err error
		snapshots, err = s.mediaService.GenerateSnapshots(uploadPath, snapshotDir)
		errChan <- err
	}()

	audioPath := filepath.Join("processed", task.ID+".mp3")
	go func() {
		errChan <- s.mediaService.ExtractAudio(uploadPath, audioPath)
	}()

	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			finishStage(mediaStage, err)
			return
		}
	}
	finishStage(mediaStage, nil)

	// 2. Create Image Grids
	gridStage := addStage("Grid Generation")
	gridPrefix := filepath.Join("processed", task.ID)
	grids, err := s.imageService.CreateGrids(snapshots, gridPrefix)
	if err != nil {
		finishStage(gridStage, err)
		return
	}
	finishStage(gridStage, nil)

	// 3. Transcription
	transcribeStage := addStage("Transcription")
	transcription, err := s.geminiService.TranscribeAudio(audioPath)
	if err != nil {
		finishStage(transcribeStage, err)
		return
	}
	finishStage(transcribeStage, nil)

	// 4. Study Guide
	guideStage := addStage("Study Guide Generation")
	// Use the first grid for overview, or adjust GeminiService to handle multiple
	guide, err := s.geminiService.GenerateStudyGuide(transcription, grids[0])
	if err != nil {
		finishStage(guideStage, err)
		return
	}
	finishStage(guideStage, nil)

	// Finalize
	task.Status = models.StatusCompleted
	task.TotalTime = time.Since(startTime)
	task.Result = &models.AnalysisResult{
		Transcription: transcription,
		StudyGuide:    *guide,
		ImagePath:     grids[0], // Primary grid
	}
	task.UpdatedAt = time.Now()
	s.storage.SaveTask(task)
}
