package middlewares

import (
	"GoApp/lib"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const siteVerifyURL = "https://www.google.com/recaptcha/api/siteverify"

type SiteVerifyResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

type SiteVerifyRequest struct {
	RecaptchaResponse string `json:"g-recaptcha-response"`
}

func checkRecaptcha(secret, response, action string) error {
	req, err := http.NewRequest(http.MethodPost, siteVerifyURL, nil)
	if err != nil {
		return err
	}

	// Add necessary request parameters.
	q := req.URL.Query()
	q.Add("secret", secret)
	q.Add("response", response)
	req.URL.RawQuery = q.Encode()

	// Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Decode response.
	var body SiteVerifyResponse
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	// Check recaptcha verification success.
	if !body.Success {
		return errors.New("unsuccessful recaptcha verify request")
	}

	// Check response score.
	if body.Score < 0.5 {
		return errors.New("lower received score than expected")
	}

	// Check response action.
	if action != "" && body.Action != action {
		return errors.New("mismatched recaptcha action")
	}

	return nil
}

func RecaptchaMiddleware(secret, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret != "" {
			var dto SiteVerifyRequest

			if err := c.ShouldBind(&dto); err != nil {
				lib.ErrorResponse(c, http.StatusUnauthorized, "")
				return
			}

			if err := checkRecaptcha(secret, dto.RecaptchaResponse, action); err != nil {
				lib.ErrorResponse(c, http.StatusUnauthorized, err.Error())
				return
			}
		}

		c.Next()
	}
}
