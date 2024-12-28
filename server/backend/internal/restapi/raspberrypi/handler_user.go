package raspberrypi

import (
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/Virgula0/progetto-dp/server/entities"
	"net/http"

	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/response"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
)

type Handler struct {
	Usecase *usecase.Usecase
}

type ReturnRaspberryPiDevicesRequest struct {
	Page uint `query:"page" validate:"required,min=0"`
}

func (u Handler) GetRaspberryPIDevices(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	var request ReturnRaspberryPiDevicesRequest

	if err := utils.ValidateQueryParameters(&request, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	rspDevices, counted, err := u.Usecase.GetRaspberryPI(userID.String(), request.Page)

	if counted == 0 {
		c.JSON(http.StatusNotFound, entities.UniformResponse{
			StatusCode: http.StatusNotFound,
			Details:    errors.ErrElementNotFound.Error(),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	temp := make([]*entities.CustomRaspberryPIResponse, 0)

	for _, dev := range rspDevices {
		tt := entities.CustomRaspberryPIResponse{
			UserUUID:        dev.UserUUID,
			RaspberryPIUUID: dev.RaspberryPIUUID,
			MachineID:       dev.MachineID,
		}

		temp = append(temp, &tt)
	}

	c.JSON(http.StatusOK, entities.ReturnRaspberryPiDevicesResponse{
		Length:  counted,
		Devices: temp,
	})

}
