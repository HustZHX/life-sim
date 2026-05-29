package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"life-sim/backend/ai"
	"life-sim/backend/auth"
	"life-sim/backend/config"
	"life-sim/backend/handler"
	"life-sim/backend/service"
	"life-sim/backend/store"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	dbPath := config.ResolveDatabasePath(cfg.DatabasePath)

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer st.Close()

	if cfg.Auth.Enabled && len(cfg.Auth.BootstrapUsers) > 0 {
		n, err := st.BootstrapUsers(cfg.Auth.BootstrapUsers)
		if err != nil {
			log.Fatalf("初始化用户失败: %v", err)
		}
		if n > 0 {
			log.Printf("已 bootstrap %d 个用户", n)
		}
	}

	if n, err := st.RecoverStaleJobs(); err != nil {
		log.Printf("恢复中断任务警告: %v", err)
	} else if n > 0 {
		log.Printf("已将 %d 个中断中的任务标记为失败", n)
	}
	log.Printf("数据库路径: %s", dbPath)
	if cfg.Auth.Enabled {
		log.Printf("鉴权已启用")
	}

	promptDir := "prompts"
	candidates := []string{
		"prompts",
		filepath.Join("backend", "prompts"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "prompts"))
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "prompts"))
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			promptDir = p
			break
		}
	}

	aiClient := ai.NewClient(&cfg.DeepSeek, promptDir)
	charSvc := service.NewCharacterService(st, aiClient, cfg)
	gameSvc := service.NewGameService(st, aiClient, charSvc)
	charSvc.BindGameService(gameSvc)
	narrSvc := service.NewNarrativeService(st, aiClient)
	dialogueSvc := service.NewDialogueService(st, aiClient, charSvc)
	authSvc := service.NewAuthService(st, &cfg.Auth)
	charHandler := handler.NewCharacterHandler(charSvc, narrSvc)
	gameHandler := handler.NewGameHandler(gameSvc)
	dialogueHandler := handler.NewDialogueHandler(dialogueSvc, charSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	authMW := auth.NewMiddleware(&cfg.Auth)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Cookie"},
		AllowCredentials: true,
	}))

	r.GET("/health", charHandler.Health)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/auth/status", authHandler.Status)
		v1.POST("/auth/gate", authHandler.Gate)
		v1.POST("/auth/login", authHandler.Login)
		v1.POST("/auth/refresh", authHandler.Refresh)

		protected := v1.Group("")
		protected.Use(authMW.GateRequired(), authMW.AuthRequired())
		{
			protected.GET("/auth/me", authHandler.Me)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.POST("/auth/change-password", authHandler.ChangePassword)

			protected.GET("/history", charHandler.ListHistory)
			protected.GET("/models", charHandler.ListModels)
			protected.POST("/suggest-names", charHandler.SuggestNames)
			protected.POST("/characters", charHandler.CreateCharacter)
			protected.POST("/game/characters", gameHandler.CreateGameCharacter)
			protected.GET("/characters/:id", charHandler.GetCharacter)
			protected.PATCH("/characters/:id/game/config", gameHandler.UpdateConfig)
			protected.POST("/characters/:id/game/era-options", gameHandler.EraOptions)
			protected.POST("/characters/:id/game/profile/generate", gameHandler.GenerateProfile)
			protected.POST("/characters/:id/game/timeline/start", gameHandler.StartTimeline)
			protected.GET("/characters/:id/nodes/:nodeId/game/choices", gameHandler.GetChoices)
			protected.POST("/characters/:id/nodes/:nodeId/game/choices", gameHandler.PostChoices)
			protected.POST("/characters/:id/nodes/:nodeId/game/choose", gameHandler.Choose)
			protected.POST("/characters/:id/resolve", charHandler.Resolve)
			protected.POST("/characters/:id/confirm", charHandler.Confirm)
			protected.POST("/characters/:id/profile/generate", charHandler.GenerateProfile)
			protected.PATCH("/characters/:id/profile", charHandler.UpdateProfile)
			protected.POST("/characters/:id/profile/randomize-field", charHandler.RandomizeProfileField)
			protected.POST("/characters/:id/profile/import-template", charHandler.ImportTemplate)
			protected.GET("/characters/:id/profile", charHandler.GetProfile)
			protected.POST("/characters/:id/timeline/recommendations", charHandler.RecommendTimeline)
			protected.POST("/characters/:id/timeline/generate", charHandler.GenerateTimeline)
			protected.POST("/characters/:id/timeline/narrative-change", charHandler.ApplyNarrativeChange)
			protected.GET("/characters/:id/timelines", charHandler.ListTimelines)
			protected.GET("/characters/:id/timelines/:tid/versions", charHandler.ListVersions)
			protected.GET("/characters/:id/timelines/:tid/branches", charHandler.ListBranches)
			protected.GET("/characters/:id/timelines/:tid/branches/overview", charHandler.GetBranchOverview)
			protected.POST("/characters/:id/timelines/:tid/branches/:vid/activate", charHandler.ActivateBranch)
			protected.GET("/characters/:id/timeline", charHandler.GetTimeline)
			protected.PATCH("/characters/:id/timelines/:tid/world-line", charHandler.UpdateWorldLine)
			protected.POST("/characters/:id/timelines/:tid/world-line/refresh", charHandler.RefreshWorldLine)
			protected.GET("/characters/:id/nodes/:nodeId", charHandler.GetNode)
			protected.POST("/characters/:id/nodes/:nodeId/lifespan/preview", charHandler.PreviewLifespan)
			protected.POST("/characters/:id/nodes/:nodeId/regenerate-events", charHandler.RegenerateNodeEvents)
			protected.POST("/characters/:id/nodes/:nodeId/rollback-to", charHandler.RollbackToNode)
			protected.PATCH("/characters/:id/nodes/:nodeId", charHandler.PatchNode)
			protected.GET("/characters/:id/versions", charHandler.ListVersions)
			protected.GET("/characters/:id/versions/:vid/diff", charHandler.GetVersionDiff)
			protected.POST("/characters/:id/versions/:vid/rollback", charHandler.Rollback)
			protected.GET("/characters/:id/narratives", charHandler.GetNarrative)
			protected.GET("/light-novels", charHandler.ListSavedLightNovels)
			protected.GET("/light-novels/:id", charHandler.GetSavedLightNovel)
			protected.POST("/characters/:id/narratives/light-novel", charHandler.GenerateLightNovel)
			protected.GET("/characters/:id/narratives/light-novel/jobs", charHandler.ListLightNovelJobs)
			protected.GET("/chronicles", charHandler.ListSavedChronicles)
			protected.GET("/chronicles/:id", charHandler.GetSavedChronicle)
			protected.POST("/characters/:id/narratives/chronicle", charHandler.GenerateChronicle)
			protected.GET("/characters/:id/narratives/chronicle/jobs", charHandler.ListChronicleJobs)
			protected.POST("/characters/:id/nodes/:nodeId/narratives/:kind", charHandler.GenerateNodeNarrative)
			protected.GET("/jobs/:jobId", charHandler.GetJob)
			protected.POST("/jobs/:jobId/retry", charHandler.RetryJob)

			protected.GET("/characters/:id/nodes/:nodeId/dialogue/identity-options", dialogueHandler.GetSavedIdentityOptions)
			protected.POST("/characters/:id/nodes/:nodeId/dialogue/identity-options", dialogueHandler.GenerateIdentityOptions)
			protected.GET("/characters/:id/nodes/:nodeId/dialogue/latest-session", dialogueHandler.GetLatestSessionForNode)
			protected.POST("/characters/:id/nodes/:nodeId/dialogue/sessions", dialogueHandler.CreateSession)
			protected.GET("/characters/:id/dialogue/sessions", dialogueHandler.ListSessions)
			protected.GET("/characters/:id/dialogue/sessions/:sessionId", dialogueHandler.GetSession)
			protected.POST("/characters/:id/dialogue/sessions/:sessionId/messages", dialogueHandler.SendMessage)
			protected.POST("/characters/:id/nodes/:nodeId/dialogue/apply-impact", dialogueHandler.ApplyImpact)
			protected.GET("/characters/:id/memories", dialogueHandler.ListMemories)
			protected.POST("/characters/:id/memories", dialogueHandler.CreateMemory)
			protected.POST("/characters/:id/memories/summarize", dialogueHandler.SummarizeMemory)
			protected.PATCH("/characters/:id/memories/:memoryId", dialogueHandler.UpdateMemory)
			protected.DELETE("/characters/:id/memories/:memoryId", dialogueHandler.DeleteMemory)
		}
	}

	addr := ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("life-sim 服务启动，监听 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
