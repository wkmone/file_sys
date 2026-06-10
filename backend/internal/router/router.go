package router

import (
	"file_sys/backend/config"
	"file_sys/backend/internal/handler"
	"file_sys/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	fileHandler *handler.FileHandler,
	folderHandler *handler.FolderHandler,
	teamHandler *handler.TeamHandler,
	ooHandler *handler.OnlyOfficeHandler,
	versionHandler *handler.VersionHandler,
	searchHandler *handler.SearchHandler,
	trashHandler *handler.TrashHandler,
	permHandler *handler.PermissionHandler,
	loginLimiter *middleware.RateLimiter,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	if cfg.AppEnv == "development" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), middleware.JSONRecovery())
	r.Use(middleware.ErrorLogger(), middleware.CORS())
	r.NoRoute(middleware.NoRouteJSON())
	r.NoMethod(middleware.NoMethodJSON())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		// Auth (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", loginLimiter.Middleware(), authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", middleware.AuthRequired(cfg.JWTAccessSecret), authHandler.Me)
			auth.PATCH("/password", middleware.AuthRequired(cfg.JWTAccessSecret), authHandler.ChangePassword)
			auth.POST("/logout-all", middleware.AuthRequired(cfg.JWTAccessSecret), authHandler.LogoutAll)
		}

		// OnlyOffice callback + file access (no auth, uses JWT in query)
		oo := api.Group("/oo")
		{
			oo.POST("/callback/:fileId", ooHandler.Callback)
			oo.POST("/callback", ooHandler.Callback) // legacy fallback
			oo.GET("/file/:fileId", ooHandler.ServeFile)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired(cfg.JWTAccessSecret))
		{
			// Users
			users := protected.Group("/users")
			{
				users.GET("", authHandler.ListUsers)
				users.GET("/search", authHandler.SearchUsers)
				users.GET("/:id", authHandler.GetUser)
			}

			// Permissions
			perms := protected.Group("/permissions")
			{
				perms.PATCH("/:id", permHandler.Update)
				perms.DELETE("/:id", permHandler.Delete)
			}

			// Folders
			folders := protected.Group("/folders")
			{
				folders.GET("", folderHandler.List)
				folders.POST("", folderHandler.Create)
				folders.GET("/shared-with-me", permHandler.SharedWithMeFolders)
				folders.GET("/:id", folderHandler.Get)
				folders.PATCH("/:id", folderHandler.Update)
				folders.DELETE("/:id", folderHandler.Delete)
				folders.GET("/:id/tree", folderHandler.Tree)
				folders.POST("/:id/share", folderHandler.Share)
				folders.GET("/:id/permissions", permHandler.ListByFolder)
			}

			// Files
			files := protected.Group("/files")
			{
				files.GET("", fileHandler.List)
				files.POST("", fileHandler.Upload)
				files.POST("/batch", fileHandler.BatchUpload)
				files.POST("/blank", fileHandler.CreateBlank)
				files.GET("/shared-with-me", permHandler.SharedWithMeFiles)
				files.GET("/:id", fileHandler.Get)
				files.PATCH("/:id", fileHandler.Update)
				files.DELETE("/:id", fileHandler.Delete)
				files.GET("/:id/download", fileHandler.Download)
				files.POST("/:id/copy", fileHandler.Copy)
				files.POST("/:id/share", fileHandler.Share)
				files.GET("/:id/permissions", permHandler.ListByFile)
			}

			// OnlyOffice editor config
			protected.POST("/oo/editor-config", ooHandler.EditorConfig)

			// Versions
			versions := protected.Group("/files/:id/versions")
			{
				versions.GET("", versionHandler.List)
				versions.GET("/:vid", versionHandler.Get)
				versions.GET("/:vid/download", versionHandler.Download)
				versions.POST("/:vid/restore", versionHandler.Restore)
			}

			// Teams
			teams := protected.Group("/teams")
			{
				teams.GET("", teamHandler.List)
				teams.GET("/discover", teamHandler.Discover)
				teams.POST("", teamHandler.Create)
				teams.GET("/:id", teamHandler.Get)
				teams.PATCH("/:id", teamHandler.Update)
				teams.DELETE("/:id", teamHandler.Delete)
				teams.GET("/:id/members", teamHandler.Members)
				teams.POST("/:id/members", teamHandler.AddMember)
				teams.PATCH("/:id/members/:userId", teamHandler.UpdateMember)
				teams.DELETE("/:id/members/:userId", teamHandler.RemoveMember)
				teams.POST("/:id/join", teamHandler.RequestJoin)
				teams.GET("/:id/pending", teamHandler.PendingRequest)
				teams.GET("/:id/requests", teamHandler.ListJoinRequests)
				teams.PATCH("/:id/requests/:requestId", teamHandler.HandleJoinRequest)
			}

			// Search
			protected.GET("/search", searchHandler.Search)

			// Trash
			trash := protected.Group("/trash")
			{
				trash.GET("", trashHandler.List)
				trash.POST("/:type/:id/restore", trashHandler.Restore)
				trash.DELETE("/:type/:id", trashHandler.PermanentDelete)
			}
		}
	}

	return r
}
