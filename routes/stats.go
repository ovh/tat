package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/controllers"
)

// InitRoutesStats initialized routes for Stats Controller
func InitRoutesStats(router *gin.Engine) {
	statsCtrl := &controllers.StatsController{}

	admin := router.Group("/stats")
	admin.Use(CheckPassword(), CheckAdmin())
	{
		admin.GET("/count", statsCtrl.Count)
		admin.GET("/instance", statsCtrl.Instance)
		admin.GET("/distribution", statsCtrl.Distribution)
		admin.GET("/db/stats", statsCtrl.DBStats)
		admin.GET("/db/replSetGetConfig", statsCtrl.DBReplSetGetConfig)
		admin.GET("/db/serverStatus", statsCtrl.DBServerStatus)
		admin.GET("/db/replSetGetStatus", statsCtrl.DBReplSetGetStatus)
		admin.GET("/db/collections", statsCtrl.DBStatsCollections)
		admin.GET("/db/slowestQueries", statsCtrl.DBGetSlowestQueries)
		admin.GET("/checkHeaders", statsCtrl.CheckHeaders)
	}
}
