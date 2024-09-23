package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Configuration constants
const (
	CacheDuration = 24 * time.Hour // since the airtable is static, this doesn't currently matter
)

// Structure of the Airtable API response
type AirtableResponse struct {
	Records []struct {
		Fields Employee `json:"fields"`
	} `json:"records"`
}

type EmployeeCache struct {
	data      []Employee
	lastFetch time.Time
	mutex     sync.RWMutex
}

var cache EmployeeCache

func checkCache() ([]Employee, bool) {
	// Acquire a read lock (can be obtained by multiple threads)
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	// If cache is valid (not empty and not expired), return the cached data
	if len(cache.data) > 0 && time.Since(cache.lastFetch) < CacheDuration {
		return cache.data, true
	}

	// Cache is either empty or expired
	return nil, false
}

// fetchFromAPI constructs the request and fetches data from the Airtable API
func fetchFromAPI() ([]byte, error) {
	apiURL := os.Getenv("API_URL")
	apiKey := os.Getenv("API_KEY")

	// Create a new HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Set the authorization header
	req.Header.Add("Authorization", "Bearer "+apiKey)

	// Make the API request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	return ioutil.ReadAll(resp.Body)
}

// parseResponse parses the API response and updates the cache
func parseResponse(body []byte) ([]Employee, error) {
	var airtableResp AirtableResponse
	err := json.Unmarshal(body, &airtableResp)
	if err != nil {
		return nil, err
	}

	employees := make([]Employee, len(airtableResp.Records))
	for i, record := range airtableResp.Records {
		employees[i] = record.Fields
	}

	// Update the cache with the new data using write lock
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.data = employees
	cache.lastFetch = time.Now()

	return employees, nil
}

// fetchEmployeeData retrieves employee data from the cache or Airtable API
func fetchEmployeeData() ([]Employee, error) {
	// First, try to get data from the cache
	if cachedData, found := checkCache(); found {
		return cachedData, nil
	}

	body, err := fetchFromAPI()
	if err != nil {
		return nil, err
	}

	return parseResponse(body)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Set up Gin router
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// Can optionally specify another route than /
	r.GET("/", func(c *gin.Context) {
		employees, err := fetchEmployeeData()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		processedEmployees := ProcessEmployeeData(employees)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"employees": processedEmployees,
		})
	})

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default port if not specified
	}

	fmt.Printf("Server is running on http://localhost:%s\n", port)
	r.Run(":" + port)
}