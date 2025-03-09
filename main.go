package main

import (
	"fmt"
)

func main() {
	// Create a new store with in-memory storage enabled
	store := NewStore("test_store.log", true, 100, 256, 4096)

	// Test Set functionality
	err := store.Set("key1", "value1")
	if err != nil {
		fmt.Println("Error setting key1:", err)
	} else {
		fmt.Println("Successfully set key1 with value1")
	}

	// Test Get functionality
	value, found := store.Get("key1")
	if found {
		fmt.Println("Retrieved value for key1:", value)
	} else {
		fmt.Println("Key1 not found")
	}

	// Test Delete functionality
	err = store.Delete("key1")
	if err != nil {
		fmt.Println("Error deleting key1:", err)
	} else {
		fmt.Println("Successfully deleted key1")
	}

	// Test Get after deletion
	value, found = store.Get("key1")
	if found {
		fmt.Println("Retrieved value for key1 after deletion:", value)
	} else {
		fmt.Println("Key1 not found after deletion")
	}

	// Test compaction (optional)
	store.Compact()
	fmt.Println("Compaction complete.")
}
