package server

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"strconv"

	"xroad/pkg/errors"
	"xroad/pkg/service"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	service.EHSService
	http.Handler
}

func (h Handler) get(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, h.EHSService.GetElections())
}

func (h Handler) getSeqNo(c *gin.Context) {
	s, ok := requireCode(c)
	if !ok {
		return
	}
	res, err := h.GetSeqNo(s)
	respond(c, http.StatusOK, res, err)
}

func (h Handler) getBatch(c *gin.Context) {
	s, ok := requireCode(c)
	if !ok {
		return
	}
	no, ok := requireSeqNo(c)
	if !ok {
		return
	}
	res, err := h.GetBatch(s, no)
	respond(c, http.StatusOK, res, err)

}

func CreateHandler(service service.EHSService, apiPath string) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(initLogger())
	h := Handler{service, router}
	api := router.Group("xroad/v1")
	router.StaticFile("openapi", apiPath)
	api.GET("/elections", h.get)
	api.GET("/elections/:electionId/lastseqno", h.getSeqNo)
	api.GET("/elections/:electionId/evotingsbatchfrom/:fromSeqNo", h.getBatch)
	return router
}

func initLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.SetFormatter(&log.JSONFormatter{})

		info := struct {
			RemoteAddr    string `json:"RemoteAddr"`
			RequestMethod string `json:"RequestMethod"`
			RequestUri    string `json:"RequestUri"`
		}{c.Request.RemoteAddr, c.Request.Method, c.Request.URL.RequestURI()}
		log.Info(info)
		c.Next()
	}
}

func requireCode(c *gin.Context) (string, bool) {
	electionId := c.Param("electionId")
	if electionId == "" {
		c.JSON(http.StatusBadRequest, errors.FieldError{
			Code:  errors.ErrBadRequest.Error(),
			Field: "electionId",
			Value: electionId,
		})
		return "", false
	}
	return electionId, true
}

func requireSeqNo(c *gin.Context) (int, bool) {
	id := c.Param("fromSeqNo")
	id64, err := strconv.ParseInt(id, 10, 0)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, errors.FieldError{
			Code:  errors.ErrBadRequest.Error(),
			Field: "fromSeqNo",
			Value: id,
		})
		return 0, false
	}
	return int(id64), true
}

// respond sends appropriate response according to the provided error and value.
func respond(c *gin.Context, status int, value interface{}, err error) {
	if err != nil {
		var fieldErr errors.FieldError
		json.Unmarshal([]byte(err.Error()), &fieldErr)
		if fieldErr.Code == errors.ErrNotFound.Error() {
			c.JSON(httpErr(errors.ErrNotFound), fieldErr)
			return
		}
		c.JSON(httpErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(status, value)
}

// httpErr transforms the given error to proper HTTP error response
func httpErr(err error) int {
	switch {
	case stderrors.Is(err, errors.ErrNotFound):
		return http.StatusNotFound
	case stderrors.Is(err, errors.ErrBadRequest):
		return http.StatusBadRequest
	case stderrors.Is(err, errors.ErrVotingEnd):
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}
