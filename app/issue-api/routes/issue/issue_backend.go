package issue

import (
	"fmt"
	"issue-service/app/issue-api/routes/models"
	"issue-service/internal"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func createIssue(database *gorm.DB, issue models.Issue) (uint, error) {
	result := database.Create(&issue)

	if result.Error != nil {
		log.WithField("error", result.Error.Error()).Error("Error creating new issue")

		if internal.IsForeignKeyError(result.Error) {
			entity := "Sprint"
			value := issue.SprintID
			if strings.Contains(result.Error.Error(), "project") {
				entity = "Project"
				value = issue.ProjectID
			}

			return 0, &models.ErrorResponse{
				ErrorMessage: fmt.Sprintf("%s with id \"%d\" does not exists", entity, value),
				ErrorCode:    404,
			}
		}

		return 0, &models.ErrorResponse{
			ErrorMessage: result.Error.Error(),
			ErrorCode:    500,
		}

	}

	return issue.ID, nil
}

func getIssues(database *gorm.DB, projectId int, sprintId int) ([]models.GetIssueResponse, error) {
	issues := []models.GetIssueResponse{}

	result := database.Model(&models.Issue{}).Where("project_id = ? and sprint_id = ?", projectId, sprintId).Find(&issues)

	if result.Error != nil {
		return []models.GetIssueResponse{}, &models.ErrorResponse{
			ErrorMessage: result.Error.Error(),
			ErrorCode:    500,
		}
	}
	return issues, nil
}

func patchIssue(database *gorm.DB, issue models.Issue) error {
	result := database.Model(&issue).Updates(issue)

	if result.Error != nil {
		if internal.IsForeignKeyError(result.Error) {
			entity := "Sprint"
			value := issue.SprintID
			if strings.Contains(result.Error.Error(), "project") {
				entity = "Project"
				value = issue.ProjectID
			}

			return &models.ErrorResponse{
				ErrorMessage: fmt.Sprintf("%s with id \"%d\" does not exists", entity, value),
				ErrorCode:    404,
			}
		}
		return &models.ErrorResponse{
			ErrorMessage: result.Error.Error(),
			ErrorCode:    500,
		}
	}
	if result.RowsAffected == 0 {
		return &models.ErrorResponse{
			ErrorMessage: fmt.Sprintf("Issue with id \"%d\" does not exists", issue.ID),
			ErrorCode:    404,
		}
	}
	return nil
}
