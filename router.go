package routes

import (
	"account/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoute(r *gin.Engine) {
	r.POST("/atm-users", controllers.CreateAccount)
	//r.GET("/atm/users", controllers.CreateAccount)
	r.POST("/atm-deposit", controllers.Deposit)
	r.POST("/atm-withdraw", controllers.Withdraw)
	r.POST("/atm-transfer", controllers.Transfer)
	r.POST("/atm-setpin", controllers.SetPin)
	r.GET("/atm-bankstatement", controllers.BankStatement)
}

