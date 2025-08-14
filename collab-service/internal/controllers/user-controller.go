package controllers

import (
	"collab-service/config"
	"collab-service/internal/dto"
	"collab-service/pkg/client"
	"encoding/csv"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Import users from CSV
// @Security BearerAuth
// @Tags users
// @Description Import users concurrently using GraphQL mutation and worker pool
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file"
// @Success 200 {object} dto.ImportUserSummary
// @Failure 400 {object} object "Bad request"
// @Router /import-users [post]
func ImportUsersHandler(c *gin.Context) {
	var graphqlClient = client.NewGraphQLClient(config.GetConfig().UserServiceEndpoint)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	reader := csv.NewReader(src)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV must have at least one user"})
		return
	}

	type job struct {
		Line  int
		Input client.CreateUserInput
	}

	type result struct {
		Line    int
		Success bool
		Message string
	}

	numWorkers := 20
	jobs := make(chan job, len(records)-1)
	results := make(chan result, len(records)-1)

	// Worker pool
	for i := 0; i < numWorkers; i++ {
		go func() {
			for j := range jobs {
				resp, err := graphqlClient.CreateUser(j.Input)
				if err != nil || !resp.Success {
					msg := err.Error()
					if resp != nil {
						msg = resp.Message
					}
					results <- result{Line: j.Line, Success: false, Message: msg}
				} else {
					results <- result{Line: j.Line, Success: true, Message: "Created"}
				}
			}
		}()
	}

	// Push jobs
	for i, row := range records[1:] {
		if len(row) < 4 {
			results <- result{Line: i + 2, Success: false, Message: "Invalid row format"}
			continue
		}
		jobs <- job{
			Line: i + 2,
			Input: client.CreateUserInput{
				Username: row[0],
				Email:    row[1],
				Password: row[2],
				Role:     client.UserType(row[3]),
			},
		}
	}
	close(jobs)

	var summary dto.ImportUserSummary
	summary.Total = len(records) - 1

	for i := 0; i < summary.Total; i++ {
		res := <-results
		summary.Results = append(summary.Results, dto.ImportUserResult(res))
		if res.Success {
			summary.Success++
		} else {
			summary.Failed++
		}
	}

	c.JSON(http.StatusOK, summary)
}
