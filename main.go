package main

import (
	"log"
	"net/http"
	"os"

	"gotask/config"
	"gotask/dto"
	"gotask/entity"
	"gotask/middleware"
	"gotask/repository"
	"gotask/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Connect Database
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	log.Println("Berhasil terhubung ke MongoDB")

	// 2. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	memberRepo := repository.NewProjectMemberRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// 3. Initialize Services
	userService := service.NewUserService(userRepo)
	projectService := service.NewProjectService(projectRepo, memberRepo, userRepo, taskRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo, memberRepo, userRepo)

	// 4. Setup Router
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Public Routes
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, dto.WebResponse{
				Success: true,
				Message: "GoTask API is healthy and running",
			})
		})

		api.POST("/register", func(c *gin.Context) {
			var req dto.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
				return
			}
			res, err := userService.Register(c.Request.Context(), req)
			if err != nil {
				c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
				return
			}
			c.JSON(http.StatusCreated, dto.WebResponse{
				Success: true,
				Message: "registrasi berhasil",
				Data:    res,
			})
		})

		api.POST("/login", func(c *gin.Context) {
			var req dto.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
				return
			}
			res, err := userService.Login(c.Request.Context(), req)
			if err != nil {
				c.JSON(http.StatusUnauthorized, dto.WebResponse{Success: false, Message: err.Error()})
				return
			}
			c.JSON(http.StatusOK, dto.WebResponse{
				Success: true,
				Message: "login berhasil",
				Data:    res,
			})
		})
	}

	// Protected Routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// ---------------- USER ENDPOINTS ----------------
		users := protected.Group("/users")
		{
			// GET /api/users (Admin only)
			users.GET("", middleware.AuthorizeRoles(entity.RoleAdmin), func(c *gin.Context) {
				res, err := userService.GetAll(c.Request.Context())
				if err != nil {
					c.JSON(http.StatusInternalServerError, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "daftar user berhasil diambil", Data: res})
			})

			// GET /api/users/:id (All authenticated)
			users.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				res, err := userService.GetByID(c.Request.Context(), id)
				if err != nil {
					c.JSON(http.StatusNotFound, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "detail user berhasil diambil", Data: res})
			})

			// PUT /api/users/:id (All authenticated - self update or admin)
			users.PUT("/:id", func(c *gin.Context) {
				id := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)

				if currentRole != entity.RoleAdmin && currentUserID != id {
					c.JSON(http.StatusForbidden, dto.WebResponse{Success: false, Message: "Anda hanya dapat memperbarui profil Anda sendiri"})
					return
				}

				var req dto.UpdateUserRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}

				// Non-admin cannot change their own role
				if currentRole != entity.RoleAdmin && req.Role != "" {
					req.Role = ""
				}

				res, err := userService.Update(c.Request.Context(), id, req)
				if err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "user berhasil diperbarui", Data: res})
			})

			// DELETE /api/users/:id (Admin only)
			users.DELETE("/:id", middleware.AuthorizeRoles(entity.RoleAdmin), func(c *gin.Context) {
				id := c.Param("id")
				if err := userService.Delete(c.Request.Context(), id); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "user berhasil dihapus"})
			})
		}

		// ---------------- PROJECT ENDPOINTS ----------------
		projects := protected.Group("/projects")
		{
			// POST /api/projects (Admin, Manager)
			projects.POST("", middleware.AuthorizeRoles(entity.RoleAdmin, entity.RoleManager), func(c *gin.Context) {
				currentUserID, _ := middleware.GetCurrentUserID(c)
				var req dto.CreateProjectRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				res, err := projectService.Create(c.Request.Context(), currentUserID, req)
				if err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusCreated, dto.WebResponse{Success: true, Message: "project berhasil dibuat", Data: res})
			})

			// GET /api/projects (All)
			projects.GET("", func(c *gin.Context) {
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				res, err := projectService.GetAll(c.Request.Context(), currentUserID, currentRole)
				if err != nil {
					c.JSON(http.StatusInternalServerError, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "daftar project berhasil diambil", Data: res})
			})

			// GET /api/projects/:id (All)
			projects.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				res, err := projectService.GetByID(c.Request.Context(), id, currentUserID, currentRole)
				if err != nil {
					c.JSON(http.StatusNotFound, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "detail project berhasil diambil", Data: res})
			})

			// PUT /api/projects/:id (Admin, Manager)
			projects.PUT("/:id", middleware.AuthorizeRoles(entity.RoleAdmin, entity.RoleManager), func(c *gin.Context) {
				id := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				var req dto.UpdateProjectRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				res, err := projectService.Update(c.Request.Context(), id, currentUserID, currentRole, req)
				if err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "project berhasil diperbarui", Data: res})
			})

			// DELETE /api/projects/:id (Admin, Manager)
			projects.DELETE("/:id", middleware.AuthorizeRoles(entity.RoleAdmin, entity.RoleManager), func(c *gin.Context) {
				id := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				if err := projectService.Delete(c.Request.Context(), id, currentUserID, currentRole); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "project berhasil dihapus"})
			})

			// POST /api/projects/:id/members (Admin, Manager)
			projects.POST("/:id/members", middleware.AuthorizeRoles(entity.RoleAdmin, entity.RoleManager), func(c *gin.Context) {
				id := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				var req dto.AddMemberRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				res, err := projectService.AddMember(c.Request.Context(), id, currentUserID, currentRole, req)
				if err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusCreated, dto.WebResponse{Success: true, Message: "member berhasil ditambahkan ke project", Data: res})
			})

			// GET /api/projects/:id/members (All)
			projects.GET("/:id/members", func(c *gin.Context) {
				id := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				res, err := projectService.GetMembers(c.Request.Context(), id, currentUserID, currentRole)
				if err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "daftar member project berhasil diambil", Data: res})
			})

			// DELETE /api/projects/:id/members/:user_id (Admin, Manager)
			projects.DELETE("/:id/members/:user_id", middleware.AuthorizeRoles(entity.RoleAdmin, entity.RoleManager), func(c *gin.Context) {
				projectID := c.Param("id")
				targetUserID := c.Param("user_id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				if err := projectService.RemoveMember(c.Request.Context(), projectID, targetUserID, currentUserID, currentRole); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "member berhasil dihapus dari project"})
			})

			// POST /api/projects/:id/tasks (Admin, Manager)
			projects.POST("/:id/tasks", middleware.AuthorizeRoles(entity.RoleAdmin, entity.RoleManager), func(c *gin.Context) {
				projectID := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				var req dto.CreateTaskRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				res, err := taskService.Create(c.Request.Context(), projectID, currentUserID, currentRole, req)
				if err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusCreated, dto.WebResponse{Success: true, Message: "task berhasil dibuat", Data: res})
			})

			// GET /api/projects/:id/tasks (All)
			projects.GET("/:id/tasks", func(c *gin.Context) {
				projectID := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				res, err := taskService.GetByProject(c.Request.Context(), projectID, currentUserID, currentRole)
				if err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "daftar task project berhasil diambil", Data: res})
			})
		}

		// ---------------- TASK ENDPOINTS ----------------
		tasks := protected.Group("/tasks")
		{
			// GET /api/tasks/me (All)
			tasks.GET("/me", func(c *gin.Context) {
				currentUserID, _ := middleware.GetCurrentUserID(c)
				res, err := taskService.GetMyTasks(c.Request.Context(), currentUserID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "daftar task saya berhasil diambil", Data: res})
			})

			// GET /api/tasks/:id (All)
			tasks.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				res, err := taskService.GetByID(c.Request.Context(), id, currentUserID, currentRole)
				if err != nil {
					c.JSON(http.StatusNotFound, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "detail task berhasil diambil", Data: res})
			})

			// PUT /api/tasks/:id (Admin, Manager)
			tasks.PUT("/:id", middleware.AuthorizeRoles(entity.RoleAdmin, entity.RoleManager), func(c *gin.Context) {
				id := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				var req dto.UpdateTaskRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				res, err := taskService.Update(c.Request.Context(), id, currentUserID, currentRole, req)
				if err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "task berhasil diperbarui", Data: res})
			})

			// PATCH /api/tasks/:id/status (All)
			tasks.PATCH("/:id/status", func(c *gin.Context) {
				id := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				var req dto.UpdateTaskStatusRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				res, err := taskService.UpdateStatus(c.Request.Context(), id, currentUserID, currentRole, req)
				if err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "status task berhasil diperbarui", Data: res})
			})

			// DELETE /api/tasks/:id (Admin, Manager)
			tasks.DELETE("/:id", middleware.AuthorizeRoles(entity.RoleAdmin, entity.RoleManager), func(c *gin.Context) {
				id := c.Param("id")
				currentUserID, _ := middleware.GetCurrentUserID(c)
				currentRole, _ := middleware.GetCurrentUserRole(c)
				if err := taskService.Delete(c.Request.Context(), id, currentUserID, currentRole); err != nil {
					c.JSON(http.StatusBadRequest, dto.WebResponse{Success: false, Message: err.Error()})
					return
				}
				c.JSON(http.StatusOK, dto.WebResponse{Success: true, Message: "task berhasil dihapus"})
			})
		}

		// ---------------- DASHBOARD ENDPOINTS ----------------
		// GET /api/dashboard (All)
		protected.GET("/dashboard", func(c *gin.Context) {
			currentUserID, _ := middleware.GetCurrentUserID(c)
			currentRole, _ := middleware.GetCurrentUserRole(c)
			res, err := taskService.GetDashboard(c.Request.Context(), currentUserID, currentRole)
			if err != nil {
				c.JSON(http.StatusInternalServerError, dto.WebResponse{Success: false, Message: err.Error()})
				return
			}
			c.JSON(http.StatusOK, dto.WebResponse{
				Success: true,
				Message: "dashboard berhasil diambil",
				Data:    res,
			})
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server berjalan di http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
