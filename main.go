package main

import (
	"log"
	"net/http"
	"strconv"

	"erp-app/bizlogic"
	"erp-app/dataaccess"
	"erp-app/glow"
	helpdesk "erp-app/help-desk-skill"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading environment variables from shell")
	}

	dataaccess.InitDB("erp.db")

	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/suppliers", func(c *gin.Context) {
			suppliers, err := bizlogic.GetAllSuppliers()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, suppliers)
		})

		api.GET("/insights", func(c *gin.Context) {
			dashboard, err := bizlogic.GetInsightsDashboard()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, dashboard)
		})

		hd := api.Group("/helpdesk")
		{
			hd.GET("/incidents", func(c *gin.Context) {
				incidents, err := helpdesk.GetIncidents()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, incidents)
			})

			hd.GET("/questionnaire", func(c *gin.Context) {
				qType := helpdesk.QuestionnaireType(c.DefaultQuery("type", "root_cause_analysis"))
				incidentID, _ := strconv.Atoi(c.DefaultQuery("incident_id", "0"))
				draft, err := helpdesk.GetQuestionnaireDraft(qType, incidentID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, draft)
			})
		}

		api.POST("/glow/chat", func(c *gin.Context) {
			var req glow.ChatRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
				return
			}
			resp, err := glow.Run(c.Request.Context(), req.Messages)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, resp)
		})
	}

	r.Static("/assets", "./frontend/dist/assets")
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	r.Run(":8080")
}
