package wordlist

import (
	"crypto/md5"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/response"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/google/uuid"
	"net/http"
)

type Handler struct {
	Usecase *usecase.Usecase
}

// UploadWordlist upload wordlist
func (u Handler) UploadWordlist(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	var request entities.UploadWordlistRequest

	if err = utils.ValidateJSON(&request, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	genId := uuid.New().String()

	wordlist := &entities.Wordlist{
		UUID:                genId,
		UserUUID:            userID.String(),
		ClientUUID:          request.ClientUUID,
		WordlistName:        request.FileName,
		WordlistHash:        fmt.Sprintf("%x", md5.Sum(request.FileBytes)),
		WordlistSize:        int64(len(request.FileBytes)),
		WordlistFileContent: request.FileBytes,
	}

	err = u.Usecase.CreateWordlist(wordlist)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entities.UploadWordlistResponse{
		WordlistID: genId,
	})
}
