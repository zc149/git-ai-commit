package worker

import (
	"bufio"
	"path/filepath"
	"strings"
	"sync"
)

// 파일 타입 분류용 Map (패키지 초기화 시 한 번만 생성)
var (
	testFileExts   map[string]bool
	docFileExts    map[string]bool
	configFileExts map[string]bool
	classifyOnce   sync.Once
)

// FileDiff는 단일 파일의 diff 정보를 담습니다.
type FileDiff struct {
	Header string // diff 헤더 라인
	Body   string // 파일의 diff 내용
}

// ParsedFile는 파싱된 파일 정보를 담습니다.
type ParsedFile struct {
	Path      string // 파일 경로
	FileType  int    // 파일 타입 (0: source, 1: test, 2: doc, 3: config)
	IsNew     bool   // 새 파일 여부
	IsDeleted bool   // 삭제된 파일 여부
	Changes   string // 변경된 내용
}

// WorkerPool은 병렬로 diff를 파싱하는 worker pool입니다.
type WorkerPool struct {
	workers    int
	input      chan FileDiff
	output     chan ParsedFile
	wg         sync.WaitGroup
	fileMap    map[int]ParsedFile // 순서 보장을 위한 맵
	resultLock sync.Mutex
	fileCount  int
}

// NewWorkerPool은 새로운 WorkerPool을 생성합니다.
func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		input:   make(chan FileDiff, workers*2),
		output:  make(chan ParsedFile, workers*2),
		fileMap: make(map[int]ParsedFile),
	}
}

// Start는 worker들을 시작합니다.
func (p *WorkerPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker는 파일 diff를 파싱하는 작업자입니다.
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	for fileDiff := range p.input {
		result := p.parseFileDiff(fileDiff)
		p.output <- result
	}
}

// parseFileDiff는 단일 파일의 diff를 파싱합니다.
func (p *WorkerPool) parseFileDiff(fileDiff FileDiff) ParsedFile {
	// 헤더에서 경로 추출
	path, isNew, isDeleted := p.parseHeader(fileDiff.Header)

	return ParsedFile{
		Path:      path,
		FileType:  classifyFileType(path),
		IsNew:     isNew,
		IsDeleted: isDeleted,
		Changes:   fileDiff.Body,
	}
}

// parseHeader는 diff 헤더를 파싱하여 경로와 상태를 추출합니다.
func (p *WorkerPool) parseHeader(header string) (path string, isNew, isDeleted bool) {
	lines := strings.Split(header, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 파일 경로 추출
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				path = strings.TrimPrefix(parts[3], "b/")
			}
		}

		// 새 파일 확인
		if strings.HasPrefix(line, "new file mode") {
			isNew = true
		}

		// 삭제된 파일 확인
		if strings.HasPrefix(line, "deleted file mode") {
			isDeleted = true
		}
	}

	return path, isNew, isDeleted
}

// Submit은 작업을 제출합니다.
func (p *WorkerPool) Submit(fileDiff FileDiff, index int) {
	p.fileCount = index + 1
	p.input <- fileDiff
}

// Close는 input 채널을 닫습니다.
func (p *WorkerPool) Close() {
	close(p.input)
}

// Wait는 모든 작업이 완료될 때까지 대기합니다.
func (p *WorkerPool) Wait() {
	p.wg.Wait()
	close(p.output)
}

// Results는 모든 결과를 수집하여 반환합니다.
func (p *WorkerPool) Results() []ParsedFile {
	results := make([]ParsedFile, 0, p.fileCount)

	for result := range p.output {
		results = append(results, result)
	}

	return results
}

// ParseDiffParallel은 worker pool을 사용하여 diff를 병렬로 파싱합니다.
func ParseDiffParallel(diff string, workers int) ([]ParsedFile, error) {
	// 빈 diff 처리
	if strings.TrimSpace(diff) == "" {
		return []ParsedFile{}, nil
	}

	// 파일 단위로 분리
	fileDiffs := splitFileDiffs(diff)

	// 파일 수가 적으면 nil 반환 (호출자가 순차 처리)
	if len(fileDiffs) <= 3 {
		return nil, nil
	}

	// Worker Pool 생성 및 시작
	pool := NewWorkerPool(workers)
	pool.Start()

	// 결과를 수집할 채널 (별도 goroutine에서 수집 시작) - 먼저 시작!
	resultChan := make(chan []ParsedFile, 1)
	go func() {
		results := pool.Results()
		resultChan <- results
	}()

	// 작업 제출
	for i, fileDiff := range fileDiffs {
		pool.Submit(fileDiff, i)
	}

	// 작업 완료 대기 (결과 수집은 별도 goroutine에서 진행 중)
	pool.Close()
	pool.Wait()

	// 결과 수집
	results := <-resultChan

	return results, nil
}

// splitFileDiffs는 diff를 파일 단위로 분리합니다.
func splitFileDiffs(diff string) []FileDiff {
	scanner := bufio.NewScanner(strings.NewReader(diff))

	var fileDiffs []FileDiff
	var currentDiff FileDiff
	var bodyLines []string
	var inDiff bool

	for scanner.Scan() {
		line := scanner.Text()

		// 새 파일 헤더 감지
		if strings.HasPrefix(line, "diff --git") {
			// 이전 파일이 있다면 저장
			if inDiff {
				currentDiff.Body = strings.Join(bodyLines, "\n")
				fileDiffs = append(fileDiffs, currentDiff)
				bodyLines = []string{}
			}

			currentDiff = FileDiff{Header: line}
			inDiff = true
		} else if inDiff {
			// 변경 라인 수집
			bodyLines = append(bodyLines, line)
		}
	}

	// 마지막 파일 저장
	if inDiff {
		currentDiff.Body = strings.Join(bodyLines, "\n")
		fileDiffs = append(fileDiffs, currentDiff)
	}

	return fileDiffs
}

// init는 패키지 초기화 시 파일 타입 분류용 Map을 생성합니다.
func init() {
	classifyOnce.Do(func() {
		// 테스트 파일 확장자
		testFileExts = make(map[string]bool)
		testFileExts["_test.go"] = true
		testFileExts[".spec.js"] = true
		testFileExts[".test.ts"] = true
		testFileExts[".test.jsx"] = true
		testFileExts[".spec.tsx"] = true

		// 문서 파일 확장자 및 파일명
		docFileExts = make(map[string]bool)
		docFileExts["README.md"] = true
		docFileExts["CHANGELOG.md"] = true
		docFileExts["CONTRIBUTING.md"] = true
		docFileExts[".md"] = true
		docFileExts[".txt"] = true
		docFileExts[".rst"] = true

		// 설정 파일 확장자 및 파일명
		configFileExts = make(map[string]bool)
		configFileExts["package.json"] = true
		configFileExts["package-lock.json"] = true
		configFileExts["go.mod"] = true
		configFileExts["go.sum"] = true
		configFileExts["Cargo.toml"] = true
		configFileExts["pom.xml"] = true
		configFileExts["build.gradle"] = true
		configFileExts["requirements.txt"] = true
		configFileExts["Makefile"] = true
		configFileExts["Dockerfile"] = true
		configFileExts[".yml"] = true
		configFileExts[".yaml"] = true
		configFileExts[".toml"] = true
		configFileExts[".json"] = true
		configFileExts[".xml"] = true
		configFileExts[".ini"] = true
		configFileExts[".conf"] = true
		configFileExts[".cfg"] = true
	})
}

// classifyFileType은 파일 경로에서 파일 타입을 결정합니다.
// O(1) 성능: Map 기반 룩업 사용
func classifyFileType(path string) int {
	base := filepath.Base(path)
	ext := filepath.Ext(path)

	// 테스트 파일 Map 룩업
	if testFileExts[base] || testFileExts[ext] {
		return 1 // FileTypeTest
	}

	// 문서 파일 Map 룩업
	if docFileExts[base] || docFileExts[ext] {
		return 2 // FileTypeDoc
	}

	// 설정 파일 Map 룩업
	if configFileExts[base] || configFileExts[ext] {
		return 3 // FileTypeConfig
	}

	// 기본적으로 소스 파일
	return 0 // FileTypeSource
}

// GetOptimalWorkerCount는 시스템에 최적화된 worker 수를 반환합니다.
func GetOptimalWorkerCount(fileCount int) int {
	// CPU 코어 수를 기본으로
	const maxWorkers = 8
	const minWorkers = 2

	// 파일 수에 따라 동적으로 조절
	if fileCount <= 10 {
		return minWorkers
	} else if fileCount <= 50 {
		return 4
	} else if fileCount <= 100 {
		return 6
	}

	return maxWorkers
}
