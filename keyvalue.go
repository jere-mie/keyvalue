package keyvalue

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// a key-value pair, with optional delete flag.
type Entry struct {
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
}

type Store struct {
	mu           sync.RWMutex
	data         map[string]string // Optional in-memory storage
	useMemory    bool              // Whether to store in memory
	filename     string
	file         *os.File
	maxKeys      int // Maximum number of entries
	maxKeySize   int // Max key size
	maxValueSize int // Max value size
}

type StoreConfig struct {
	UseMemory    bool // Whether to store in memory
	MaxKeys      int  // Maximum number of entries
	MaxKeySize   int  // Max key size
	MaxValueSize int  // Max value size
}

func NewStore(filename string, config StoreConfig) *Store {
	s := &Store{
		filename:     filename,
		useMemory:    config.UseMemory,
		data:         make(map[string]string),
		maxKeys:      config.MaxKeys,
		maxKeySize:   config.MaxKeySize,
		maxValueSize: config.MaxValueSize,
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	s.file = file

	if config.UseMemory {
		s.load()
	}

	return s
}

// build the in-memory map
func (s *Store) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filename)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			fmt.Println("Error parsing log entry:", err)
			continue
		}

		if entry.Deleted {
			delete(s.data, entry.Key)
		} else {
			s.data[entry.Key] = entry.Value
		}

		if len(s.data) > s.maxKeys {
			fmt.Println("Store exceeded max keys limit, consider compaction.")
			break
		}
	}
}

// safely set a key-value pair and append to the log file
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate key size
	if len(key) > s.maxKeySize {
		return fmt.Errorf("key exceeds max size of %d bytes", s.maxKeySize)
	}
	// Validate value size
	if len(value) > s.maxValueSize {
		return fmt.Errorf("value exceeds max size of %d bytes", s.maxValueSize)
	}
	// Check max keys limit
	if s.useMemory && len(s.data) >= s.maxKeys {
		return fmt.Errorf("store has reached max number of keys (%d)", s.maxKeys)
	}

	entry := Entry{Key: key, Value: value}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}

	_, err = s.file.WriteString(string(data) + "\n")
	if err != nil {
		return fmt.Errorf("error writing to log file: %v", err)
	}

	if s.useMemory {
		s.data[key] = value
	}

	return nil
}

// retrieve a value by key
func (s *Store) Get(key string) (string, bool) {
	if s.useMemory {
		s.mu.RLock()
		defer s.mu.RUnlock()
		val, exists := s.data[key]
		return val, exists
	}

	// File-only mode: Scan the log file for the most recent entry
	file, err := os.Open(s.filename)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return "", false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lastValue string
	var exists bool

	for scanner.Scan() {
		line := scanner.Text()
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.Key == key {
			if entry.Deleted {
				lastValue = ""
				exists = false
			} else {
				lastValue = entry.Value
				exists = true
			}
		}
	}

	return lastValue, exists
}

// mark a key as deleted in the log and remove it from memory.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := Entry{Key: key, Deleted: true}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}

	_, err = s.file.WriteString(string(data) + "\n")
	if err != nil {
		return fmt.Errorf("error writing to log file: %v", err)
	}

	if s.useMemory {
		delete(s.data, key)
	}

	return nil
}

// rewrite the log file, removing deleted and outdated entries
func (s *Store) Compact() {
	s.mu.Lock()
	defer s.mu.Unlock()

	tempFile := s.filename + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		fmt.Println("Error creating temp log file:", err)
		return
	}
	defer file.Close()

	// Use the latest data to write a clean log
	for key, value := range s.data {
		entry := Entry{Key: key, Value: value}
		data, _ := json.Marshal(entry)
		file.WriteString(string(data) + "\n")
	}

	// replace old log with compacted version
	os.Rename(tempFile, s.filename)
	s.file.Close()
	s.file, _ = os.OpenFile(s.filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
}

func (s *Store) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.file.Close()
}

func (s *Store) FindByFunction(fn func(string, string) bool) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results = make(map[string]string)
	for key, value := range s.data {
		if fn(key, value) {
			results[key] = value
		}
	}

	// file-only mode
	if !s.useMemory {
		file, err := os.Open(s.filename)
		if err != nil {
			fmt.Println("Error opening log file:", err)
			return nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			var entry Entry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				continue
			}
			if entry.Deleted {
				continue
			}
			if fn(entry.Key, entry.Value) {
				results[entry.Key] = entry.Value
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading log file:", err)
			return nil, err
		}
	}

	var result []Entry
	for key, value := range results {
		result = append(result, Entry{Key: key, Value: value})
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no entries found matching the criteria")
	}

	return result, nil
}
