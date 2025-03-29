package main

import (
	"fmt"
	"os"

	"github.com/jere-mie/keyvalue"
)

func testStore(useMemory bool, fileName string) {
	// Create a new store with in-memory storage enabled
	store := keyvalue.NewStore(fileName, keyvalue.StoreConfig{
		UseMemory:    useMemory,
		MaxKeys:      100,
		MaxKeySize:   256,
		MaxValueSize: 1024,
	})

	// Add 5 entries, k1 to k5 with values v1 to v5
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf("v%d", i)
		err := store.Set(key, value)
		if err != nil {
			fmt.Printf("❌ Error setting key %s: %v\n", key, err)
		} else {
			fmt.Printf("✅ Set key %s to value %s\n", key, value)
		}
	}

	// get the values of k1 to k5
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("k%d", i)
		value, exists := store.Get(key)
		if !exists {
			fmt.Printf("❌ Key %s does not exist\n", key)
		} else {
			fmt.Printf("✅ Key %s has value %s\n", key, value)
		}
	}

	// Delete k3
	err := store.Delete("k3")
	if err != nil {
		fmt.Printf("❌ Error deleting key k3: %v\n", err)
	} else {
		fmt.Println("✅ Deleted key k3")
	}

	// Check if k3 exists
	value, exists := store.Get("k3")
	if !exists {
		fmt.Println("✅ Key k3 does not exist after deletion")
	} else {
		fmt.Printf("❌ Key k3 still exists with value %s\n", value)
	}

	// use store.FindByFunction to find "k4", and "v5"
	entries, err := store.FindByFunction(func(key, value string) bool {
		return key == "k4" || value == "v5"
	})
	if err != nil {
		fmt.Printf("❌ Error finding entries: %v\n", err)
	} else {
		fmt.Println("✅ Found entries:")
		for _, entry := range entries {
			fmt.Printf("✅ Key: %s, Value: %s, Deleted: %t\n", entry.Key, entry.Value, entry.Deleted)
		}
	}

	// Test compaction (optional)
	store.Compact()
	fmt.Println("✅ Compaction complete.")

	store.Close()
	fmt.Println("✅ Store closed.")

	// remove the log file
	if err := os.Remove(fileName); err != nil {
		fmt.Printf("❌ Error removing log file %s: %v\n", fileName, err)
	}
	fmt.Printf("✅ Removed log file %s\n", fileName)
}

func main() {
	// Test with in-memory storage
	fmt.Println("Testing with in-memory storage:")
	testStore(true, "inmemory_store.log")

	// Test with file-based storage
	fmt.Println("\nTesting with file-based storage:")
	testStore(false, "file_store.log")
}
