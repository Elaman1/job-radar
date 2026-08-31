package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"job-radar/internal/config"
	"net/http"
	"net/url"
	"time"
)

type CheckerHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewCheckerHandler(db *pgxpool.Pool, cfg *config.Config) *CheckerHandler {
	return &CheckerHandler{db, cfg}
}

func (h *CheckerHandler) CheckerHandler(c *gin.Context) {
	ctx := c.Request.Context()
	client := http.Client{
		Timeout: time.Second * 5,
	}

	//body := []byte(`{"name":"test"}`)
	//reader := bytes.NewReader(body)

	// --------------------------------
	//u, err := url.Parse("https://api.adzuna.com/v1/api/jobs/gb/search/1")
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"message": fmt.Sprintf("error while parsing url: %s", err.Error()),
	//	})
	//	return
	//}
	//
	//query := u.Query()
	//
	//query.Set("app_id", h.cfg.Sources.Adzuna.AppId)
	//query.Set("app_key", h.cfg.Sources.Adzuna.AppKey)
	//query.Set("results_per_page", "50")
	//query.Set("what", "golang")
	//
	//u.RawQuery = query.Encode()

	u, err := url.Parse("https://search.api.careerjet.net/v4/query")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("error while parsing url: %s", err.Error()),
		})
		return
	}

	query := u.Query()

	query.Set("locale_code", "en_GB")
	query.Set("keywords", "golang")
	query.Set("page", "1")
	query.Set("page_size", "100")
	query.Set("sort", "date")

	query.Set("user_ip", "USER_IP")
	query.Set("user_agent", "Mozilla/5.0")

	u.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		u.String(),
		nil,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("request failed: %v", err),
		})
		return
	}

	request.SetBasicAuth(h.cfg.Sources.CareerJet.APIKey, "")
	request.Header.Set("Referer", "https://job-radar-2u8b.onrender.com/")

	response, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("response failed: %v", err),
		})
		return
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("response failed: status %v", response.Status),
		})
		return
	}

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("response failed: %v", err),
		})
		return
	}

	c.Data(
		response.StatusCode,
		"application/json",
		respBody,
	)
}
