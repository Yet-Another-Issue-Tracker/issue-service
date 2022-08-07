package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

func createProject(database *gorm.DB, projectName string, projectType string, projectClient string) ([]byte, error) {
	project := Project{Name: projectName, Client: projectClient, Type: projectType}

	result := database.Create(&project)

	if result.Error != nil {
		log.WithField("error", result.Error.Error()).Error("Error creating new project")
		return nil, result.Error
	}

	response := CreateProjectResponse{Id: fmt.Sprint(project.ID)}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func CreateAddProjectHandler(database *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		projectId, err := createProject(database, "", "", "")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(projectId)
	}
}