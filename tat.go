package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/ovh/tat/controllers"
	"github.com/ovh/tat/models"
	"github.com/ovh/tat/routes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mainCmd = &cobra.Command{
	Use:   "tat",
	Short: "Run Tat Engine",
	Long:  `Run Tat Engine`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetEnvPrefix("tat")
		viper.AutomaticEnv()

		if viper.GetBool("production") {
			// Only log the warning severity or above.
			log.SetLevel(log.WarnLevel)
			log.Info("Set Production Mode ON")
			gin.SetMode(gin.ReleaseMode)
		} else {
			log.SetLevel(log.DebugLevel)
		}

		if viper.GetString("tat_log_level") != "" {
			switch viper.GetString("tat_log_level") {
			case "debug":
				log.SetLevel(log.DebugLevel)
			case "info":
				log.SetLevel(log.InfoLevel)
			case "error":
				log.SetLevel(log.ErrorLevel)
			}
		}

		router := gin.Default()
		router.Use(cors.Middleware(cors.Config{
			Origins:         "*",
			Methods:         "GET, PUT, POST, DELETE",
			RequestHeaders:  "Origin, Authorization, Content-Type, Accept, Tat_Password, Tat_Username",
			ExposedHeaders:  "Tat_Password, Tat_Username",
			MaxAge:          50 * time.Second,
			Credentials:     true,
			ValidateHeaders: false,
		}))

		models.NewStore()
		routes.InitRoutesGroups(router)
		routes.InitRoutesMessages(router)
		routes.InitRoutesPresences(router)
		routes.InitRoutesTopics(router)
		routes.InitRoutesUsers(router)
		routes.InitRoutesStats(router)
		routes.InitRoutesSystem(router)
		routes.InitRoutesSockets(router)
		router.Run(":" + viper.GetString("listen_port"))
	},
}

var versionNewLine bool

// The version command prints this service.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version.",
	Long:  "The version of Tat Engine.",
	Run: func(cmd *cobra.Command, args []string) {
		if versionNewLine {
			fmt.Println(controllers.VERSION)
		} else {
			fmt.Print(controllers.VERSION)
		}
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&versionNewLine, "versionNewLine", "", true, "New line after version number")
	mainCmd.AddCommand(versionCmd)

	flags := mainCmd.Flags()

	flags.Bool("production", false, "Production mode")
	flags.Bool("no-smtp", false, "No SMTP mode")
	flags.String("tat-log-level", "", "Tat Log Level: debug, info or warn")
	flags.String("listen-port", "8080", "Tat Engine Listen Port")
	flags.String("exposed-scheme", "http", "Tat URI Scheme http or https exposed to client")
	flags.String("exposed-host", "localhost", "Tat Engine Hostname exposed to client")
	flags.String("exposed-port", "8080", "Tat Engine Port exposed to client")
	flags.String("exposed-path", "", "Tat Engine Path exposed to client, ex: host:port/tat/engine /tat/engine is exposed path")
	flags.String("db-addr", "127.0.0.1:27017", "Address of the mongodb server")
	flags.String("db-user", "", "User to authenticate with the mongodb server. If \"false\", db-user is not used")
	flags.String("db-password", "", "Password to authenticate with the mongodb server. If \"false\", db-password is not used")
	flags.String("db-rs-tags", "", "Link hostname with tag on mongodb replica set - Optional: hostnameA:tagName:value,hostnameB:tagName:value. If \"false\", db-rs-tags is not used")
	flags.String("smtp-host", "", "SMTP Host")
	flags.String("smtp-port", "", "SMTP Port")
	flags.Bool("smtp-tls", false, "SMTP TLS")
	flags.String("smtp-user", "", "SMTP Username")
	flags.String("smtp-password", "", "SMTP Password")
	flags.String("smtp-from", "", "SMTP From")
	flags.String("allowed-domains", "", "Users have to use theses emails domains. Empty: no-restriction. Ex: --allowed-domains=domainA.org,domainA.com")
	flags.String("default-group", "", "Default Group for new user")
	flags.Bool("username-from-email", false, "Username are extracted from first part of email. first.lastame@domainA.org -> username: first.lastname")
	flags.Bool("websocket-enabled", false, "Enable or not websockets on this instance")
	flags.String("header-trust-username", "", "Header Trust Username: for example, if X-Remote-User and X-Remote-User received in header -> auto accept user without testing tat_password. Use it with precaution")
	flags.String("trusted-usernames-emails-fullnames", "", "Tuples trusted username / email / fullname. Example: username:email:Firstname1_Fullname1,username2:email2:Firstname2_Fullname2")
	flags.String("default-domain", "", "Default domains for mail for trusted username")

	viper.BindPFlag("production", flags.Lookup("production"))
	viper.BindPFlag("no_smtp", flags.Lookup("no-smtp"))
	viper.BindPFlag("tat_log_level", flags.Lookup("tat-log-level"))
	viper.BindPFlag("listen_port", flags.Lookup("listen-port"))
	viper.BindPFlag("exposed_scheme", flags.Lookup("exposed-scheme"))
	viper.BindPFlag("exposed_host", flags.Lookup("exposed-host"))
	viper.BindPFlag("exposed_port", flags.Lookup("exposed-port"))
	viper.BindPFlag("exposed_path", flags.Lookup("exposed-path"))
	viper.BindPFlag("db_addr", flags.Lookup("db-addr"))
	viper.BindPFlag("db_user", flags.Lookup("db-user"))
	viper.BindPFlag("db_password", flags.Lookup("db-password"))
	viper.BindPFlag("db_rs_tags", flags.Lookup("db-rs-tags"))
	viper.BindPFlag("smtp_host", flags.Lookup("smtp-host"))
	viper.BindPFlag("smtp_port", flags.Lookup("smtp-port"))
	viper.BindPFlag("smtp_tls", flags.Lookup("smtp-tls"))
	viper.BindPFlag("smtp_user", flags.Lookup("smtp-user"))
	viper.BindPFlag("smtp_password", flags.Lookup("smtp-password"))
	viper.BindPFlag("smtp_from", flags.Lookup("smtp-from"))
	viper.BindPFlag("allowed_domains", flags.Lookup("allowed-domains"))
	viper.BindPFlag("default_group", flags.Lookup("default-group"))
	viper.BindPFlag("username_from_email", flags.Lookup("username-from-email"))
	viper.BindPFlag("websocket_enabled", flags.Lookup("websocket-enabled"))
	viper.BindPFlag("header_trust_username", flags.Lookup("header-trust-username"))
	viper.BindPFlag("trusted_usernames_emails_fullnames", flags.Lookup("trusted-usernames-emails-fullnames"))
	viper.BindPFlag("default_domain", flags.Lookup("default-domain"))
}

func main() {
	mainCmd.Execute()
}
