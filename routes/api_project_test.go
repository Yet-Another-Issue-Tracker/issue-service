package routes

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/require"
)

func TestCreateProject(testCase *testing.T) {
	config, err := GetConfig("../.env")
	if err != nil {
		log.Fatalf("Error reading env configuration: %s", err.Error())
		return
	}
	database, err := ConnectDatabase(config)

	if err != nil {
		log.Fatalf("Error connecting to database %s", err.Error())
		return
	}
	expectedResponse := CreateProjectResponse{
		Id: "1",
	}
	expectedJsonReponse, _ := json.Marshal(expectedResponse)

	testCase.Run("createProject return the new id", func(t *testing.T) {
		SetupAndResetDatabase(database)

		response, err := createProject(database, "", "", "")

		var foundProject Project

		database.First(&foundProject)

		require.Equal(t, nil, err)
		require.Equal(t, string(expectedJsonReponse), string(response))
	})

	testCase.Run("createProject with specific name and type", func(t *testing.T) {
		SetupAndResetDatabase(database)
		expectedProjectName := "project-name"
		expectedType := "project-type"
		expectedClient := "project-client"

		createProject(database, expectedProjectName, expectedType, expectedClient)

		var foundProject Project

		database.First(&foundProject)

		require.Equal(t, nil, err)
		require.Equal(t, expectedProjectName, foundProject.Name)
		require.Equal(t, expectedType, foundProject.Type)
		require.Equal(t, expectedClient, foundProject.Client)
	})

	testCase.Run("create two projects", func(t *testing.T) {
		SetupAndResetDatabase(database)

		createProject(database, "project-1", "", "")
		createProject(database, "project-2", "", "")

		var foundProjects []Project

		result := database.Find(&foundProjects)
		log.Printf("number of rows %d", result.RowsAffected)
		require.Equal(t, nil, err)
		require.Equal(t, 2, int(result.RowsAffected))
	})

	testCase.Run("createProject returns error if project with same name already exits", func(t *testing.T) {
		SetupAndResetDatabase(database)
		expectedError := "ERROR: duplicate key value violates unique constraint \"idx_projects_name\" (SQLSTATE 23505)"

		expectedProjectName := "project-name"
		expectedType := "project-type"
		expectedClient := "project-client"

		_, err1 := createProject(database, expectedProjectName, expectedType, expectedClient)

		require.Equal(t, nil, err1)

		_, err2 := createProject(database, expectedProjectName, expectedType, expectedClient)

		require.Equal(t, expectedError, err2.Error())
	})
}

type Case struct {
	expected    CreateProjectResponse
	input       Project
	description string
}
type TestCases []Case

func TestCreateProjectHandler(test *testing.T) {
	config, err := GetConfig("../.env")
	if err != nil {
		log.Fatalf("Error reading env configuration: %s", err.Error())
		return
	}
	testRouter := NewRouter(config)
	database, err := ConnectDatabase(config)
	if err != nil {
		log.Fatalf("Error connecting to database %s", err.Error())
		return
	}

	cases := TestCases{
		{
			input: Project{
				Name:   "project-name",
				Client: "client-name",
				Type:   "project-type",
			}, expected: CreateProjectResponse{
				Id: "1",
			},
			description: "Project is created",
		},
	}

	for _, testCase := range cases {
		test.Run(testCase.description, func(t *testing.T) {
			SetupAndResetDatabase(database)
			expectedJsonReponse, _ := json.Marshal(testCase.expected)
			requestBody, err := json.Marshal(testCase.input)

			if err != nil {
				log.WithField("error", err.Error()).Error("Error marshaling json")
			}

			bodyReader := bytes.NewReader(requestBody)

			responseRecorder := httptest.NewRecorder()
			request, requestError := http.NewRequest(http.MethodPost, "/v1/projects", bodyReader)
			require.NoError(t, requestError, "Error creating the /projects request")

			testRouter.ServeHTTP(responseRecorder, request)
			statusCode := responseRecorder.Result().StatusCode
			require.Equal(t, http.StatusOK, statusCode, "The response statusCode should be 200")

			rawBody := responseRecorder.Result().Body
			body, readBodyError := ioutil.ReadAll(rawBody)
			require.NoError(t, readBodyError)
			require.Equal(t, string(expectedJsonReponse), string(body), "The response body should be the expected one")
		})
	}

}