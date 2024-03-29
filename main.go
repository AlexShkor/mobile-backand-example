package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"bitbucket.org/AlexShkor/cozytime/data"
	"bitbucket.org/AlexShkor/cozytime/routes"
	"bitbucket.org/AlexShkor/cozytime/settings"

	"github.com/googollee/go-socket.io"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"

	"labix.org/v2/mgo"
)

var db *mgo.Database

func main() {

	conf := settings.Get()

	if conf == nil {
		fmt.Println("No config found, terminating.")
		os.Exit(-1)
	}


	mongoSession, err := mgo.Dial(conf.ConnectionString)
	if err != nil {
		panic(err)
	}
	db = mongoSession.DB(conf.DatabaseName)
	tokens := data.NewTokensService(db.C(conf.CollectionName))
	games := data.NewGamesService(db.C("games"))
    
    
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
    server.SetMaxConnection(1000)
	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
       
       so.On("test", func (msg string)  {
			log.Println("emit:", so.Emit("test", msg))
       })
       
		so.On("userconnect", func(token string) {
            userId, authError := tokens.IsAuthorized(token) 
            if authError != nil {
            log.Fatal(authError)
            return
            }
            so.Join(userId)
            })
            
            so.On("disconnection", func() {
                log.Println("on disconnect")
            })
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

    
	http.Handle("/socket.io/", server)
    
	e := echo.New()

	e.Hook(stripTrailingSlash)

	e.Use(mw.Logger())
	e.Use(mw.Recover())

	router := routes.NewRouter(tokens)
	e.Get("/", router.HelloWorld)
	e.Post("/authorize", router.Register)
	e.Post("/submitcode", router.SubmitCode)
    e.Post("/log", router.Log)


	authorizedGroup := e.Group("/api", BearerAuth(tokens))
	authorizedGroup.Post("/setname", router.SetName)
	authorizedGroup.Post("/findbyphones", router.GetFriendsList)

	game := routes.NewGameRouter(games, tokens, server)

	authorizedGroup.Post("/game/start", game.StartGame)
	authorizedGroup.Post("/game/create", game.CreateGame)
	authorizedGroup.Post("/game/join", game.JoinGame)
	authorizedGroup.Post("/game/leave", game.LeaveGame)
	authorizedGroup.Post("/game/delete", game.DeleteGame)
	authorizedGroup.Post("/game/stop", game.StopGame)
	authorizedGroup.Post("/game/get", game.GetGame)
	authorizedGroup.Post("/game/all", game.GetMyGames)

	authorizedGroup.Post("/secret", func(c *echo.Context) error {
		return c.String(200, "You are authorized!\n")
	})

	adminGroup := e.Group("/admin")

	adminGroup.Static("/assets", "assets")
	adminGroup.Get("", router.AdminIndex)

    http.Handle("/", e)
    
    log.Fatal(http.ListenAndServe(":" + conf.Port, nil))
}
