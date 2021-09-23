package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"bufio"
	"strings"
    "math/rand"
    "time"
	"net/url"
)

type Boat struct {
	x, y int
	down bool
}

type Game struct {
	board   Grid
	myBoats []*Boat
}

// GLOBAL VARIABLES
var boats []*Boat
var myGame Game
var board Grid
var port, portAdver string

func main() {
	port = os.Args[1]
	portAdver = os.Args[2]

	// Init
	board = NewGrid(10, 10)
	placeBoats()
	board.reset()
	myGame = Game{board, boats}


	// SERVER
	http.HandleFunc("/board", boardHandler)
	http.HandleFunc("/boats", boatsHandler)
	http.HandleFunc("/hit", hitHandler)
	message := ":"+port
	go http.ListenAndServe(message, nil)


	// GAME
	reader := bufio.NewReader(os.Stdin)
	channel := make(chan string)

	fmt.Println("--- BATAILLE NAVALE ---")

	fmt.Println("Si tous les joueurs sont prêts, appuyez sur une touche")

	test, _ := reader.ReadString('\n')
	test = strings.TrimSuffix(test, "\n")
	fmt.Println(" ")

	for {
		go getBoard(channel)
		resBoard := <-channel
		fmt.Println("Plateau de l'adversaire :\n", resBoard)
		
		go getBoats(channel)
		resBoats := <-channel
		fmt.Println("Nombre de bateaux restants :\n", resBoats)

		fmt.Println("Veuillez entrer les coordonnées de la case que vous souhaitez attaquer (0-9)\nEx: 1 5")
		commande, _ := reader.ReadString('\n')
		commande = strings.TrimSuffix(commande, "\n")
		cmd := strings.Split(commande, " ")

		for len(cmd) != 2 {
			fmt.Println("Mauvais format. Recommencez (x y)")
			commande, _ = reader.ReadString('\n')
			commande = strings.TrimSuffix(commande, "\n")
			cmd = strings.Split(commande, " ")
		}

		go postHit(cmd[0], cmd[1], channel)
		resHit := <-channel
		fmt.Println("Verdict : ", resHit)

		time.Sleep(2*time.Second)
	}
}

// CLIENT FUNCTIONS

func placeBoats() {
	// Aleatoire
	for i := 0; i < 3; i++ {
		rand.Seed(time.Now().UnixNano())
		x := rand.Intn(board.width-1)
		y := rand.Intn(board.height-1)

		for _, boat := range boats {	
			for boat.x == x || boat.x == x+1 {
				rand.Seed(time.Now().UnixNano())
				x = rand.Intn(board.width-1)
			}
			for boat.y == y || boat.y == y+1 {
				rand.Seed(time.Now().UnixNano())
				y = rand.Intn(board.height-1)
			}
		}
		
		boats = append(boats, &Boat{x, y, false})
	}
}

func updateBoard(x, y int, hit bool) {
	if hit == true {
		board.set('o', x, y)
	} else {
		board.set('x', x, y)
	}
}

func getBoard(channel chan string) {
	response, err := http.Get("http://localhost:"+portAdver+"/board")
	if(err != nil) {
		fmt.Println(err.Error())
	}
	defer response.Body.Close()
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	channel <- string(responseData)
}

func getBoats(channel chan string) {
	response, err := http.Get("http://localhost:"+portAdver+"/boats")
	if(err != nil) {
		fmt.Println(err.Error())
	}
	defer response.Body.Close()
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	channel <- string(responseData)
}

func postHit(x, y string, channel chan string) {
	response, err := http.PostForm("http://localhost:"+portAdver+"/hit", url.Values{"x": {x}, "y": {y}})
	if(err != nil) {
		fmt.Println(err.Error())
	}
	defer response.Body.Close()
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	channel <- string(responseData)
}

// SERVER ROUTES

func boardHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		fmt.Fprintln(w, board)
	}
}

func boatsHandler(w http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {

		stillUp := 0
		for _, boat := range boats {
			if boat.down == false {
				stillUp++
			}
		}

		fmt.Fprintln(w, stillUp)
	}
}

func hitHandler(w http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			fmt.Println("Something went bad")
			fmt.Fprintln(w, "Something went bad")
			return
		}

		x, _ := strconv.Atoi(req.FormValue("x"))
		y, _ := strconv.Atoi(req.FormValue("y"))
		hit := false

		for _, boat := range boats {
			if (boat.x == x && boat.y == y) {
				boat.down = true
				hit = true
				break
			}
		}

		updateBoard(x, y, hit)

		if hit == true {
			fmt.Fprintln(w, "Touché !")
		} else {
			fmt.Fprintln(w, "Loupé...")
		}

	}
}
