package api

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/api/middleware"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

func RegisterRoutes(container *restful.Container, handler *Handler) {
	ws := new(restful.WebService)

	ws.
		Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// Health endpoint
	ws.
		Route(ws.GET("health").
			To(handler.Health).
			Doc("Health check").
			Metadata(restfulspec.KeyOpenAPITags, []string{"health"}).
			Writes(HealthResponse{}).
			Returns(200, "OK", HealthResponse{}))

	ws.
		Route(ws.POST("/evaluate").
			To(handler.Evaluate).
			Doc("Evaluate agent response").
			Metadata(restfulspec.KeyOpenAPITags, []string{"evaluate"}).
			Reads(models.EvaluationRequest{}).
			Writes(models.EvaluationResult{}).
			Returns(200, "OK", models.EvaluationResult{}).
			Returns(400, "Bad Request", middleware.ErrorResponse{}).
			Returns(500, "Internal Server Error", middleware.ErrorResponse{}))

	container.Add(ws)
}
