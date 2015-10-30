package utils

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

var (
	// TatHeaderUsername is a Header send by user with his username
	TatHeaderUsername = "Tat_username"

	// TatHeaderUsernameLower is a Header in lowercase
	TatHeaderUsernameLower = strings.ToLower(TatHeaderUsername)

	// TatHeaderUsernameLowerDash is a Header in lowercase, and dash : tat-username
	TatHeaderUsernameLowerDash = strings.Replace(TatHeaderUsernameLower, "_", "-", -1)

	// TatCtxIsAdmin is used in Gin Context True if user is admin
	TatCtxIsAdmin = "Tat_isAdmin"

	// TatCtxIsSystem is used in Gin Context True if user is a system user
	TatCtxIsSystem = "Tat_isSystem"
)

// IsTatAdmin return true if user is admin. Get value in gin.Context
func IsTatAdmin(ctx *gin.Context) bool {
	value, exist := ctx.Get(TatCtxIsAdmin)
	if value != nil && exist && value.(bool) == true {
		return true
	}
	return false
}

// IsTatSystem return true if user is a system user. Get value in gin.Context
func IsTatSystem(ctx *gin.Context) bool {
	value, exist := ctx.Get(TatCtxIsSystem)
	if value != nil && exist && value.(bool) == true {
		return true
	}
	return false
}

// GetCtxUsername return username, getting value in gin.Context
func GetCtxUsername(ctx *gin.Context) string {
	username, exist := ctx.Get(TatHeaderUsername)
	if username == nil || !exist {
		log.Debug("Username is null in context !")
		return ""
	}
	return username.(string)
}
