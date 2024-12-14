package middlewares

type SuperAdminAccessToken struct {
	SuperAdminBotToken string `header:"superadmin_bot_token"`
}

/*
func (u *TokenAuth) IsAdminMiddlware(c *gin.Context) {
	var request SuperAdminAccessToken

	if err := c.ShouldBindHeader(&request); err != nil {
		statusCode := http.StatusInternalServerError
		c.AbortWithStatusJSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    "error while binding headers from middleware",
		})
		return
	}

	if request.SuperAdminBotToken != "" {
		if constants.SuperAdminBotToken != request.SuperAdminBotToken {
			statusCode := http.StatusUnauthorized
			c.AbortWithStatusJSON(statusCode, response.UniformResponse{
				StatusCode: statusCode,
				Details:    "you're not an admin",
			})
			return
		}
		c.Next() // token verified
		return
	}

	// otherwise check token
	token := u.TokenValidation(c)

	if token == "" {
		c.Abort()
		return
	}

	jwtClaims, err := u.Usecase.GetDataFromToken(token)

	if err != nil {
		statusCode := http.StatusInternalServerError
		c.AbortWithStatusJSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    "Error occurred, admin middlware , canot get data from token",
		})
		return
	}

	role, ok := jwtClaims[constants.RoleString]

	if !ok {
		statusCode := http.StatusInternalServerError
		c.AbortWithStatusJSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    "role not parsed from jwt token",
		})
		return
	}

	if role.(string) != string(constants.ADMIN) {
		statusCode := http.StatusUnauthorized
		c.AbortWithStatusJSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    "you're not an admin",
		})
		return
	}
}
*/
