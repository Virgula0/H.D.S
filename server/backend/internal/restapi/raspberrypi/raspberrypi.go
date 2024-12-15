package raspberrypi

import (
	"net/http"

	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/response"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
)

type Handler struct {
	Usecase *usecase.Usecase
}

// Needed to avoid to display encryption key
type CustomRaspberryPIResponse struct {
	UserUUID        string
	RaspberryPIUUID string
	MachineID       string
}

type ReturnRaspberryPiDevicesResponse struct {
	Length  int                          `json:"length"`
	Devices []*CustomRaspberryPIResponse `json:"devices"`
}

func (u Handler) GetRaspberryPIDevices(w http.ResponseWriter, r *http.Request) {
	c := response.ResponseInitializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	rspDevices, len, err := u.Usecase.GetRaspberryPI(userID.String(), 1) // TODO: handle offset from request

	if len <= 0 {
		c.JSON(http.StatusNotFound, response.UniformResponse{
			StatusCode: http.StatusNotFound,
			Details:    errors.ErrElementNotFound.Error(),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	temp := make([]*CustomRaspberryPIResponse, 0)

	for _, dev := range rspDevices {
		tt := CustomRaspberryPIResponse{
			UserUUID:        dev.UserUUID,
			RaspberryPIUUID: dev.RaspberryPIUUID,
			MachineID:       dev.MachineID,
		}

		temp = append(temp, &tt)
	}

	c.JSON(http.StatusOK, ReturnRaspberryPiDevicesResponse{
		Length:  len,
		Devices: temp,
	})

}
