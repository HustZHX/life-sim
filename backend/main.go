package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"life-sim/backend/ai"
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

	if n, err := st.RecoverStaleJobs(); err != nil {
		log.Printf("恢复中断任务警告: %v", err)
	} else if n > 0 {
		log.Printf("已将 %d 个中断中的任务标记为失败", n)
	}
	log.Printf("数据库路径: %s", dbPath)

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
	narrSvc := service.NewNarrativeService(st, aiClient)
	charHandler := handler.NewCharacterHandler(charSvc, narrSvc)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	r.GET("/health", charHandler.Health)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/history", charHandler.ListHistory)
		v1.GET("/models", charHandler.ListModels)
		v1.POST("/suggest-names", charHandler.SuggestNames)
		v1.POST("/characters", charHandler.CreateCharacter)
		v1.GET("/characters/:id", charHandler.GetCharacter)
		v1.POST("/characters/:id/resolve", charHandler.Resolve)
		v1.POST("/characters/:id/confirm", charHandler.Confirm)
		v1.POST("/characters/:id/profile/generate", charHandler.GenerateProfile)
		v1.PATCH("/characters/:id/profile", charHandler.UpdateProfile)
		v1.POST("/characters/:id/profile/randomize-field", charHandler.RandomizeProfileField)
		v1.POST("/characters/:id/profile/import-template", charHandler.ImportTemplate)
		v1.GET("/characters/:id/profile", charHandler.GetProfile)
		v1.POST("/characters/:id/timeline/recommendations", charHandler.RecommendTimeline)
		v1.POST("/characters/:id/timeline/generate", charHandler.GenerateTimeline)
		v1.GET("/characters/:id/timelines", charHandler.ListTimelines)
		v1.GET("/characters/:id/timelines/:tid/versions", charHandler.ListVersions)
		v1.GET("/characters/:id/timeline", charHandler.GetTimeline)
		v1.GET("/characters/:id/nodes/:nodeId", charHandler.GetNode)
		v1.POST("/characters/:id/nodes/:nodeId/lifespan/preview", charHandler.PreviewLifespan)
		v1.PATCH("/characters/:id/nodes/:nodeId", charHandler.PatchNode)
		v1.GET("/characters/:id/versions", charHandler.ListVersions)
		v1.GET("/characters/:id/versions/:vid/diff", charHandler.GetVersionDiff)
		v1.POST("/characters/:id/versions/:vid/rollback", charHandler.Rollback)
		v1.GET("/characters/:id/narratives", charHandler.GetNarrative)
		v1.POST("/characters/:id/narratives/light-novel", charHandler.GenerateLightNovel)
		v1.POST("/characters/:id/nodes/:nodeId/narratives/:kind", charHandler.GenerateNodeNarrative)
		v1.GET("/jobs/:jobId", charHandler.GetJob)
	}

	addr := ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("life-sim 服务启动，监听 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
