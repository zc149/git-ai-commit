package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Cache는 커밋 메시지를 캐싱하는 구조체입니다.
type Cache struct {
	DiffHash  string    `json:"diff_hash"` // diff의 SHA256 해시 (식별용)
	Message   string    `json:"message"`   // 선택한 커밋 메시지
	Timestamp time.Time `json:"timestamp"` // 저장 시간
}

// CacheManager는 캐시를 관리합니다.
type CacheManager struct {
	cacheFile string
}

// NewCacheManager는 새로운 CacheManager 인스턴스를 생성합니다.
func NewCacheManager() (*CacheManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".git-ai-commit")
	cacheFile := filepath.Join(cacheDir, "cache.json")

	// 캐시 디렉토리 생성
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &CacheManager{
		cacheFile: cacheFile,
	}, nil
}

// Save는 캐시를 저장합니다.
func (cm *CacheManager) Save(diffHash string, message string) error {
	cache := Cache{
		DiffHash:  diffHash,
		Message:   message,
		Timestamp: time.Now(),
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cm.cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Load은 캐시를 로드합니다.
func (cm *CacheManager) Load(diffHash string) (*Cache, error) {
	data, err := os.ReadFile(cm.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 캐시가 없으면 nil 반환 (에러 아님)
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	// diff hash가 같은지 확인
	if cache.DiffHash != diffHash {
		return nil, nil // hash가 다르면 nil 반환
	}

	return &cache, nil
}

// Clear는 캐시를 삭제합니다.
func (cm *CacheManager) Clear() error {
	if err := os.Remove(cm.cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %w", err)
	}
	return nil
}

// CalculateHash는 문자열의 SHA256 해시를 계산합니다.
func CalculateHash(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}
