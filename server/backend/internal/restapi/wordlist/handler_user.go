package wordlist

import (
	"crypto/md5" // #nosec G501 disable weak hash alert, it is not used for crypto stuff
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
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

	if len(request.FileBytes) > errors.MaxUploadSize {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    errors.ErrFileTooBig,
		})
		return
	}

	genID := uuid.New().String()

	wordlist := &entities.Wordlist{
		UUID:                genID,
		UserUUID:            userID.String(),
		ClientUUID:          request.ClientUUID,
		WordlistName:        request.FileName,
		WordlistHash:        fmt.Sprintf("%x", md5.Sum(request.FileBytes)), // #nosec G401 disable weak hash alert, it is not used for crypto stuff
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
		WordlistID: genID,
	})
}
