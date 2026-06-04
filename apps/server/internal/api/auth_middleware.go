package api

import "github.com/gin-gonic/gin"

type apiRole string

const (
	apiRoleOperator apiRole = "operator"
	apiRoleRecorder apiRole = "recorder"
	apiRoleRobot    apiRole = "robot"
	apiRoleSystem   apiRole = "system"
)

func (s *Server) operatorAuthMiddleware() gin.HandlerFunc {
	return roleAuthPlaceholder(apiRoleOperator)
}

func (s *Server) recorderAuthMiddleware() gin.HandlerFunc {
	return roleAuthPlaceholder(apiRoleRecorder)
}

func (s *Server) robotAuthMiddleware() gin.HandlerFunc {
	return roleAuthPlaceholder(apiRoleRobot)
}

func (s *Server) systemAuthMiddleware() gin.HandlerFunc {
	return roleAuthPlaceholder(apiRoleSystem)
}

func roleAuthPlaceholder(role apiRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("apiRole", string(role))
		c.Next()
	}
}
