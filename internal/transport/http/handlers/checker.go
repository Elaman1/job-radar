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

	query := u.Query()

	query.Set("keywords", "golang")

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
