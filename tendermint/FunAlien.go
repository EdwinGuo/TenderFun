package main

import (
        "fmt"
        "strconv"
        "flag"
        "tenderfun/tendermint/util"
        //"time"
)

// Assume that each Alien will start at a random city
// Assume that each Alien will move in a sequential fashion


// AlienCommander is the head of all the aliens, it control
// the moves of them, alien commander will only stop under three condition
// 1) All the cities are destroyed
// 2) All the aliens are destroyed
// 3) It exaust all the target iteration(10000)


// TODO, change this back to 10000 once finished development
var numOfCommands = 3


var commanderCounter = 1
var aliens map[Alien]Status
var cityLookup map[string]map[string]bool
var cityAlienLookup map[string]map[Alien]bool
var toCommanderSignalChan chan string
var terminatorChan chan int
var gateKeeperChan chan bool

var aggregatorSignalChan chan Alien

func AlienCommander(iters int) <-chan int {
        c := make(chan int)
        go func() {
                for {
                        counter := 1
                        command, more := <-toCommanderSignalChan
                        fmt.Printf("AlienCommander: I will trigger the %dth round of attack\n", commanderCounter)
                        if more {
                                if (command == "continue" && commanderCounter < iters && len(aliens) > 0) {
                                        c <- counter
                                        commanderCounter++
                                        //time.Sleep(1 * time.Second)
                                } else {
                                        close(c)
                                }
                        } else {
                                fmt.Println("AlienCommander: All job are finished, commander is exiting.....")
                                close(c)
                        }
                }
        }()
        return c
}


func alienMove(alien Alien, lookupCity map[string]map[string]bool, cityAlienLookup map[string]map[Alien]bool) {
        neighbour := lookupCity[aliens[alien].city]
        city := util.RandMove(neighbour)
        fmt.Println("Aggregator: what is the neighbour: ", neighbour, " for alien ", aliens[alien])
        fmt.Println("Aggregator: what is the new city: ", city, " for alien ", aliens[alien])
        if city != "" {
                // move to the new city
                status := aliens[alien]
                status.city = city
                aliens[alien] = status
                // update cityAlienlookup
                aliensLookup := cityAlienLookup[city]
                if len(aliensLookup) == 0{
                        newMap := make(map[Alien]bool)
                        newMap[alien] = true
                        cityAlienLookup[city] = newMap
                } else{
                        aliensLookup[alien] = true
                }

        }
}

// after everymove, check whether the city has two aliens stay there,
// if yes, then we need to clean up the lookupcity by update each city's
// neighbour and also update cityAlienlookup to remove the destroyed aliens
// and cities
func cityCheckup(alien Alien, lookupCity map[string]map[string]bool, cityAlienLookup map[string]map[Alien]bool) {
        city := aliens[alien].city
        alis := cityAlienLookup[city]
        fmt.Println("Aggregator: I'm in cityCheckup, aliens: ", aliens)
        fmt.Println("Aggregator: I'm in cityCheckup, lookupCity", lookupCity)
        fmt.Println("Aggregator: I'm in cityCheckup, cityAlienLookup",  cityAlienLookup)

        if (len(alis) > 1){
                // Clean up and destroy
                // send a termination signal to the aliens
                // update lookupcity
                // and citya connect to cityb,
                // then cityb must connect to citya
                for alien, _ := range alis {
                        fmt.Println("Aggregator: I'm in cityCheckup, and I'm close channels",  alis)
                        close(alien.commandChan)
                        delete(alis, alien)
                        delete(aliens, alien)
                }

                delete(cityAlienLookup, city)

                for ci, _ := range lookupCity[city]{
                        delete(lookupCity[city], ci)
                }

                delete(lookupCity, city)
        }
}

func consumer(alien Alien, terminatorChan chan int, aggregatorSignalChan chan Alien) {
        for {
                _, more := <-alien.commandChan
                if more {
                        fmt.Println("Consumer: Alien ", alien.name, "had been activated and is in motion!")
                        aggregatorSignalChan <- alien
                } else {
                        fmt.Println("Consumer: I'm done: ", alien.name)
                        terminatorChan <- 1
                        return
                }
        }
}

func currentRoundFinish() bool {
        marker := 0
        index := 0
        for _, status := range aliens {
                if index == 0 {
                        marker = status.counter
                } else {
                        if status.counter != marker {
                                return false
                        }
                }
                index++
        }

        return true
}

// aggregator is in charge of collect the info from alien consumer and send the signal to aliencommander
func Aggregator(ch chan Alien) {
        go func() {
                fmt.Println("Aggregator, I'm in work...")
                for {
                        alien, more := <- ch
                        if more{
                                fmt.Println("Aggregator: receive work...")

                                // make the move
                                // if able to move, then move and check the condition in the city
                                // if there is a hit (two aliens in the same city), then update the lookupcity and
                                // cityAlienlookup and destroy the alien
                                fmt.Println("Aggregator: I'm checking alien before the move: ", aliens[alien])
                                alienMove(alien, cityLookup, cityAlienLookup)
                                fmt.Println("Aggregator: I'm checking alien after the move: ", aliens[alien])

                                status := aliens[alien]
                                status.counter++
                                aliens[alien] = status

                                cityCheckup(alien, cityLookup, cityAlienLookup)

                                // if all the alive aliens finish their current move, then
                                // we flush the signal to the next commander for the next round
                                // if all aliens are destroyed, then end of the game
                                if (len(aliens) == 0) {
                                        fmt.Println("ALl the aliens had been destroyed, terminating the game")
                                        close(ch)
                                } else if (currentRoundFinish()){
                                        fmt.Println("ALl aliens had made their move for the current round... Next round will start shortly...")
                                        toCommanderSignalChan <- "continue"
                                } else {
                                        fmt.Println("aggregator just receive a signal from alien, waiting other aliens.. from ", alien)
                                }

                        } else {
                                toCommanderSignalChan <- "done"
                                fmt.Println("ALl the aliens had been destroyed, terminating the game, done.....")
                                close(ch)
                        }
                }
        }()
}


func Terminator(size int){
        go func() {
                total := 0
                for {
                        fmt.Println("Terminator: start to working on my summing job.....", total)
                        n, more := <-terminatorChan
                        fmt.Println("Terminator: I'm sumer and i'm about to served: ", n)

                        if more {
                                total += n
                                fmt.Println("Terminator: I'm sumer, the current total is: ", total)
                                if total == size{
                                        gateKeeperChan <- true
                                }

                        } else {
                                fmt.Println("Terminator: I'm done yo....: ", n)
                                gateKeeperChan <- true
                                return
                        }
                }
        }()
}

// This function is to fan out the command to the aliens from the commander
func PassCommandToAlien(ch <-chan int) {
        go func() {
                fmt.Println("PassCommandToAlien: I'm in work mode now")
                for i := range ch {
                        for alien, _ := range aliens {
                                alien.commandChan <- i
                        }
                }

                fmt.Println("PassCommandToAlien: I'm done here...")
                for alien, _ := range aliens {
                        // close all our fanOut channels when the input channel is exhausted.
                        fmt.Println("PassCommandToAlien: Work done now, close channels")
                        close(alien.commandChan)
                }
        }()

}

type Alien struct {
        name        string
        commandChan chan int
}

type Status struct{
        city string
        counter int
}

func GenerateAliens(num int, cities []string) map[Alien]Status {
        // use the int to keep track of which round the alien is at
        alis := make(map[Alien]Status)

        for ali := 1; ali <= num; ali++ {
                name := "Alien-" + strconv.Itoa(ali)
                ch := make(chan int)
                city := util.RandCity(cities)
                alien := Alien{name, ch}
                status := Status{city, 0}
                // initialize with 0
                alis[alien] = status
        }
        fmt.Println("Factory, my aliens: ", alis)
        return alis
}

// generate a map that for a given city, ablt to check what aliens are in the city
func GenerateCityALienLookup(aliens map[Alien]Status) map[string]map[Alien]bool {
        result := make(map[string](map[Alien]bool), len(cityLookup))
        for city, _ := range cityLookup {
                for alien, status := range aliens {
                        if status.city == city {
                                m := make(map[Alien]bool)
                                m[alien] = true
                                result[city] = m
                        }
                }
        }
        return result
}

func initParas(numOfAliens int, mapFile string) {
        cityLookup = util.ParseCity(mapFile)
        cities := make([]string, len(cityLookup))

        // populate the cities
        cnt := 0
        for city, _ := range cityLookup {
                cities[cnt] = city
                cnt++
        }

        aliens = GenerateAliens(numOfAliens, cities)

        // given a city, which alien(s) is/are there
        cityAlienLookup = GenerateCityALienLookup(aliens)

        // To notify commander that downstream are clear and ready to make the next move
        toCommanderSignalChan = make(chan string)

        terminatorChan = make(chan int)

        gateKeeperChan = make(chan bool)

        aggregatorSignalChan = make(chan Alien)
}


func main() {
        var numOfAliens int
        var mapFile string

        flag.IntVar(&numOfAliens, "numofaliens", 2, "a integer that indicate the number of of aliens")
        flag.StringVar(&mapFile, "mapfile", "", "the file that contains the map and direction information")

        flag.Parse()

        initParas(numOfAliens, mapFile)

        fmt.Println("Hello there, there are total", len(cityLookup), " cities and ", numOfAliens, " aliens")

        // to initialize the job
        go func(){toCommanderSignalChan <- "continue"}()

        // start the commander:
        commandChannel := AlienCommander(numOfCommands)

        PassCommandToAlien(commandChannel)

        Aggregator(aggregatorSignalChan)

        Terminator(numOfAliens)

        for alien, _ := range aliens {
                go consumer(alien, terminatorChan, aggregatorSignalChan)
        }

        <-gateKeeperChan
        fmt.Println("Cool, game finished, hope you enjoyed it!")
}
