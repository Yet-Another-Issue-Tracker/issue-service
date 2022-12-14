package sprint

import (
	"encoding/json"
	"fmt"
	"issue-service/app/issue-api/routes/models"
	"issue-service/internal"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var sprintNumber = "12345"
var projectId = 1

func TestCreateSprint(testCase *testing.T) {
	config, err := internal.GetConfig("../../../../.env")
	if err != nil {
		log.Fatalf("Error reading env configuration: %s", err.Error())
		return
	}

	database, err := internal.ConnectDatabase(config)
	if err != nil {
		log.Fatalf("Error connecting to database %s", err.Error())
		return
	}
	expectedSprintNumber := "12345"
	expectedResponse := 1
	expectedJsonReponse, _ := json.Marshal(expectedResponse)
	projectId := 1
	inputSprint := models.Sprint{
		Number:    expectedSprintNumber,
		ProjectID: projectId,
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 7),
		Completed: false,
	}

	testCase.Run("createSprint return the new id", func(t *testing.T) {
		internal.SetupAndResetDatabase(database)
		inputProject := models.Project{
			Name:   internal.GetRandomStringName(10),
			Type:   "project-type",
			Client: "project-client",
		}
		database.Create(&inputProject)
		response, err := createSprint(database, inputSprint)

		var foundSprint models.Sprint

		database.First(&foundSprint)

		require.Equal(t, nil, err)
		require.Equal(t, string(expectedJsonReponse), fmt.Sprint(response))
	})

	testCase.Run("successfully create two sprint", func(t *testing.T) {
		expectedSprint2Number := "98765"

		internal.SetupAndResetDatabase(database)
		inputProject := models.Project{
			Name:   internal.GetRandomStringName(10),
			Type:   "project-type",
			Client: "project-client",
		}
		database.Create(&inputProject)

		createSprint(database, inputSprint)

		inputSprint2 := inputSprint
		inputSprint2.Number = expectedSprint2Number

		createSprint(database, inputSprint2)

		var foundSprints []models.Sprint

		result := database.Find(&foundSprints)
		log.Printf("number of rows %d", result.RowsAffected)
		require.Equal(t, nil, err)
		require.Equal(t, 2, int(result.RowsAffected))
		require.Equal(t, expectedSprintNumber, foundSprints[0].Number)
		require.Equal(t, expectedSprint2Number, foundSprints[1].Number)
	})

	testCase.Run("createSprint returns error if sprint with same number already exits", func(t *testing.T) {
		expectedError := fmt.Sprintf("Sprint with number \"%s\" already exists", expectedSprintNumber)

		internal.SetupAndResetDatabase(database)
		inputProject := models.Project{
			Name:   internal.GetRandomStringName(10),
			Type:   "project-type",
			Client: "project-client",
		}
		database.Create(&inputProject)

		_, err1 := createSprint(database, inputSprint)

		require.Equal(t, nil, err1)

		_, err2 := createSprint(database, inputSprint)

		require.Equal(t, expectedError, err2.Error())
	})

	testCase.Run("createSprint returns error if project does not exists", func(t *testing.T) {
		expectedError := fmt.Sprintf("Project with id \"%d\" does not exists", projectId)

		internal.SetupAndResetDatabase(database)

		_, err := createSprint(database, inputSprint)

		require.Equal(t, expectedError, err.Error())
	})
}

func TestPatchSprint(testCase *testing.T) {
	config, err := internal.GetConfig("../../../../.env")
	if err != nil {
		log.Fatalf("Error reading env configuration: %s", err.Error())
		return
	}

	database, err := internal.ConnectDatabase(config)
	if err != nil {
		log.Fatalf("Error connecting to database %s", err.Error())
		return
	}

	testCase.Run("patchSprint update the Completed field only", func(t *testing.T) {
		internal.SetupAndResetDatabase(database)
		_, sprintId := internal.CreateProjectAndSprint(database)

		inputSprint := models.Sprint{
			ID:        sprintId,
			Completed: true,
		}

		err := patchSprint(database, inputSprint)
		require.Equal(t, nil, err)

		var foundSprint models.Sprint
		database.First(&foundSprint)

		require.Equal(t, true, foundSprint.Completed)
		require.Equal(t, sprintNumber, foundSprint.Number)
	})

	testCase.Run("createSprint returns error if project does not exists", func(t *testing.T) {
		nonExistingProjectId := 99999
		expectedError := fmt.Sprintf("Project with id \"%d\" does not exists", nonExistingProjectId)
		internal.SetupAndResetDatabase(database)
		_, sprintId := internal.CreateProjectAndSprint(database)

		patchSprintInput := models.Sprint{
			ID:        sprintId,
			ProjectID: nonExistingProjectId,
			Completed: true,
		}

		err := patchSprint(database, patchSprintInput)
		require.Equal(t, expectedError, err.Error())
	})

	testCase.Run("patchSprint return error if sprint does not exist", func(t *testing.T) {
		internal.SetupAndResetDatabase(database)
		wrongSprintId := uint(999999)
		expectedError := fmt.Sprintf("Sprint with id \"%d\" does not exists", wrongSprintId)

		patchSprintInput := models.Sprint{
			ID:        wrongSprintId,
			Completed: true,
		}

		err := patchSprint(database, patchSprintInput)
		require.Equal(t, expectedError, err.Error())
	})
}

func TestGetSprints(testCase *testing.T) {
	config, err := internal.GetConfig("../../../../.env")
	if err != nil {
		log.Fatalf("Error reading env configuration: %s", err.Error())
		return
	}

	database, err := internal.ConnectDatabase(config)
	if err != nil {
		log.Fatalf("Error connecting to database %s", err.Error())
		return
	}

	testCase.Run("getSprint return one sprint", func(t *testing.T) {
		internal.SetupAndResetDatabase(database)
		_, sprintId := internal.CreateProjectAndSprint(database)

		foundSprints, err := getSprints(database, projectId)

		require.Equal(t, nil, err)
		require.Equal(t, sprintNumber, foundSprints[0].Number)
		require.Equal(t, sprintId, foundSprints[0].ID)
	})

	testCase.Run("getSprint return a list of sprints", func(t *testing.T) {
		internal.SetupAndResetDatabase(database)
		newSprintNumber := "newSprint"

		projectId, sprint1Id := internal.CreateProjectAndSprint(database)
		sprint2Id := internal.CreateTestSprint(database, newSprintNumber, int(projectId))

		foundSprints, err := getSprints(database, int(projectId))

		require.Equal(t, nil, err)
		require.Equal(t, sprintNumber, foundSprints[0].Number)
		require.Equal(t, newSprintNumber, foundSprints[1].Number)
		require.Equal(t, sprint1Id, foundSprints[0].ID)
		require.Equal(t, sprint2Id, foundSprints[1].ID)
	})

	testCase.Run("getSprint return empty list if project does not exist", func(t *testing.T) {
		nonExistingProjectId := 99999
		internal.SetupAndResetDatabase(database)

		internal.CreateProjectAndSprint(database)

		foundSprints, err := getSprints(database, nonExistingProjectId)

		require.Equal(t, nil, err)
		require.Equal(t, []models.GetSprintResponse{}, foundSprints)

	})

}
