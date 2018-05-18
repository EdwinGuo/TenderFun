package main

import (
        "fmt"
        "strconv"
        "flag"
        "tenderfun/tendermint/util"
        "time"
)

// Assume that each Alien will start at a random city
// Assume that each Alien will move in a sequential fashion


// AlienCommander is the head of all the aliens, it control
// the moves of them, alien commander will only stop under three condition
// 1) All the cities are destroyed
// 2) All the aliens are destroyed
// 3) It exaust all the target iteration(10000)

func AlienCommander(iters int, toCommanderSignalChan chan string, aliens map[Alien]int) <-chan int {
        c := make(chan int)
        go func() {
                counter := 1
                command, more := <-toCommanderSignalChan
                if more {
                        if (command == "continue" && counter < iters && len(aliens) > 0) {
                                c <- counter
                                counter++
                                time.Sleep(10 * time.Second)
                        } else {
                                close(c)
                        }
                } else {
                        fmt.Println("All job are finished, commander is exiting.....")
                        close(c)
                }
        }()
        return c
}


func alienMove(alien Alien, lookupCity map[string]map[string]bool, cityAlienLookup map[string]map[Alien]bool) {
        neighbour := lookupCity[alien.city]
        city := util.RandMove(neighbour)
        if city != "" {
                // move to the new city
                alien.city = city
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
        city := alien.city
        alis := cityAlienLookup[city]

        if (len(alis) > 1){
                // Clean up and destroy
                // send a termination signal to the aliens
                // update lookupcity
                // and citya connect to cityb,
                // then cityb must connect to citya
                for alien, _ := range alis {
                        close(alien.commandChan)
                        delete(alis, alien)
                }

                delete(cityAlienLookup, city)

                for ci, _ := range lookupCity[city]{
                        delete(lookupCity[city], ci)
                }

                delete(lookupCity, city)
        }
}

func consumer(alien Alien, terminatorChan chan int, aggregatorSignalChan chan Alien) {
        fmt.Println("Alien ", alien.name, " had been activated")
        for {
                _, more := <-alien.commandChan
                if more {
                        fmt.Println("Alien ", alien.name, " is in motion!")
                        aggregatorSignalChan <- alien
                } else {
                        fmt.Println("I'm done: ", alien.name)
                        terminatorChan <- 1
                        return
                }
        }
}

func currentRoundFinish(aliens map[Alien]int) bool {
        marker := 0
        index := 0
        for _, round := range aliens {
                if index == 0 {
                        marker = round
                } else {
                        if round != marker {
                                return false
                        }
                }
                index++
        }

        return true
}

// aggregator is in charge of collect the info from alien consumer and send the signal to aliencommander
func aggregator(ch chan Alien, toCommanderSignalChan chan string, aliens map[Alien]int, lookupCity map[string]map[string]bool, cityAlienLookup map[string]map[Alien]bool ) {
        go func() {
                fmt.Println("aggregator is in work...")
                for {
                        alien, more := <- ch
                        if more{
                                fmt.Println("aggregator receive work...")

                                // make the move
                                // if able to move, then move and check the condition in the city
                                // if there is a hit (two aliens in the same city), then update the lookupcity and
                                // cityAlienlookup and destroy the alien
                                alienMove(alien, lookupCity, cityAlienLookup)

                                aliens[alien]++

                                cityCheckup(alien, lookupCity, cityAlienLookup)

                                // if all the alive aliens finish their current move, then
                                // we flush the signal to the next commander for the next round
                                // if all aliens are destroyed, then end of the game
                                if (len(aliens) == 0) {
                                        fmt.Println("ALl the aliens had been destroyed, terminating the game")
                                        close(ch)
                                } else if (currentRoundFinish(aliens)){
                                        fmt.Println("ALl aliens had made their move for the current round... Next round will start shortly...")
                                        toCommanderSignalChan <- "continue"
                                } else {
                                        fmt.Println("aggregator just receive a signal from alien, waiting other aliens.. from ", alien)
                                }

                        } else {
                                toCommanderSignalChan <- "done"
                                close(ch)
                        }
                }
        }()
}


func Terminator(terminatorChan <-chan int, donec chan bool, size int){
        go func() {
                total := 0
                for {
                        fmt.Println("start to working on my summing job.....", total)
                        n, more := <-terminatorChan
                        fmt.Println("I'm sumer and i'm about to served: ", n)

                        if more {
                                total += n
                                fmt.Println("I'm sumer, the current total is: ", total)
                                if total == size{
                                        donec <- true
                                }

                        } else {
                                fmt.Println("I'm done yo....: ", n)
                                donec <- true
                                return
                        }
                }
        }()
}

// This function is to fan out the command to the aliens from the commander
func PassCommandToAlien(ch <-chan int, aliens map[Alien]int) {
        go func() {
                fmt.Println("sup yo")
                for i := range ch {
                        fmt.Println("sup ... ", i)
                        for alien, _ := range aliens {
                                alien.commandChan <- i
                        }
                }

                fmt.Println("done here")

                for alien, _ := range aliens {
                        // close all our fanOut channels when the input channel is exhausted.
                        close(alien.commandChan)
                }
        }()

}

type Alien struct {
        name        string
        commandChan chan int
        city  string
}

func GenerateAliens(num int, cities []string) map[Alien]int {
        // use the int to keep track of which round the alien is at
        aliens := make(map[Alien]int)

        for ali := 1; ali <= num; ali++ {
                name := "Alien-" + strconv.Itoa(ali)
                ch := make(chan int)
                city := util.RandCity(cities)
                alien := Alien{name, ch, city}
                // initialize with 0
                aliens[alien] = 0
        }
        return aliens
}

// generate a map that for a given city, ablt to check what aliens are in the city
func GenerateCityALienLookup(aliens map[Alien]int, cityLookup map[string]map[string]bool) map[string]map[Alien]bool {
        result := make(map[string](map[Alien]bool), len(cityLookup))
        for city, _ := range cityLookup {
                for alien, _ := range aliens {
                        if alien.city == city {
                                m := make(map[Alien]bool)
                                m[alien] = true
                                result[city] = m
                        }
                }
        }
        return result
}

// TODO< to check when one interation is finished, we need to signal the commander

func main() {
        var numOfAliens int
        var mapFile string

        flag.IntVar(&numOfAliens, "numofaliens", 2, "a integer that indicate the number of of aliens")
        flag.StringVar(&mapFile, "mapfile", "", "the file that contains the map and direction information")

        flag.Parse()

        cityLookup := util.ParseCity(mapFile)

        fmt.Println("Hello there, there are total", len(cityLookup), " cities and ", numOfAliens, " aliens")

        cities := make([]string, len(cityLookup))

        // populate the cities
        cnt := 0
        for city, _ := range cityLookup {
                cities[cnt] = city
                cnt++
        }

        // TODO, change this back to 10000 once finished development
        numOfCommands := 10

        aliens := GenerateAliens(numOfAliens, cities)

        // given a city, which alien(s) is/are there
        cityAlienLookup := GenerateCityALienLookup(aliens, cityLookup)

        // To notify commander that downstream are clear and ready to make the next move
        toCommanderSignalChan := make(chan string)

        // to initialize the job
        go func(){toCommanderSignalChan <- "continue"}()

        // start the commander:
        commandChannel := AlienCommander(numOfCommands, toCommanderSignalChan, aliens)

        terminatorChan := make(chan int)

        gateKeeperChan := make(chan bool)

        aggregatorSignalChan := make(chan Alien, len(aliens))

        PassCommandToAlien(commandChannel, aliens)

        aggregator(aggregatorSignalChan, toCommanderSignalChan, aliens, cityLookup, cityAlienLookup)

        Terminator(terminatorChan, gateKeeperChan, numOfAliens)

        for alien, _ := range aliens {
                go consumer(alien, terminatorChan, aggregatorSignalChan)
        }

        <-gateKeeperChan
        //time.Sleep(10* time.Second)
        fmt.Println("Cool, game finished, hope you enjoyed it!")
}
