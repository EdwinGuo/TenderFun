package main

import (
        "fmt"
        "time"
        "strconv"
        "flag"
        "tenderfun/tendermint/util"
)

// Assume that each Alien will start at a random city

// AlienCommander is the head of all the aliens, it control
// the moves of them, alien commander will only stop under three condition
// 1) All the cities are destroyed
// 2) All the aliens are destroyed
// 3) It exaust all the target iteration(10000)

func AlienCommander(iters int) <-chan int {
        c := make(chan int)
        go func() {
                for i := 0; i < iters; i++ {
                        c <- i
                        time.Sleep(1 * time.Second)
                }
                close(c)
        }()
        return c
}

func consumer(cin <-chan int, terminatorChan chan int) {
        for {
                fmt.Println("haha !!!")
                n, more := <-cin
                if more {
                        fmt.Println("I receive the value: ", n)
                } else {
                        fmt.Println("I'm done: ", n)
                        terminatorChan <- 1
                        return
                }
        }
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
func PassCommandToAlien(ch <-chan int, aliens [](Alien)) {
        go func() {
                fmt.Println("sup yo")
                for i := range ch {
                        fmt.Println("sup ... ", i)
                        for _, alien := range aliens {
                                alien.commandChan <- i
                        }
                }

                fmt.Println("done here")

                for _, alien := range aliens {
                        // close all our fanOut channels when the input channel is exhausted.
                        close(alien.commandChan)
                }
        }()

}

type Alien struct {
        name        string
        commandChan chan int
        currentLoc  string
}

func GenerateAliens(num int, cities []string) ([]Alien, map[string](Alien)) {
        aliens := []Alien{}
        lookup := make(map[string](Alien))

        for ali := 1; ali <= num; ali++ {
                name := "Alien-" + strconv.Itoa(ali)
                ch := make(chan int)
                city := util.RandCity(cities)
                alien := Alien{name, ch, city}
                aliens = append(aliens, alien)
                lookup[name] = alien
        }
        return aliens, lookup
}

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

        // start the commander:
        commandChannel := AlienCommander(numOfCommands)

        aliens, lookup := GenerateAliens(numOfAliens, cities)

        fmt.Println(lookup)

        terminatorChan := make(chan int)

        gateKeeperChan := make(chan bool)

        PassCommandToAlien(commandChannel, aliens)

        Terminator(terminatorChan, gateKeeperChan, numOfAliens)


        for i :=0; i < numOfAliens; i++ {
                go consumer(aliens[i].commandChan, terminatorChan)
        }

        <-gateKeeperChan
        //time.Sleep(10* time.Second)
        fmt.Println("Cool, game finished, hope you enjoyed it!")
}
