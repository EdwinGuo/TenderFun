package util

import (
        "fmt"
        "io/ioutil"
        "strings"
        "math/rand"
        "time"
)

// parseCity will parse a file and return a map(key is the city name) of
// adjacent cities in map (key is city name)
func ParseCity(file string) map[string]map[string]bool{
        data := make(map[string]map[string]bool)
        b, err := ioutil.ReadFile(file)
        if err != nil {
                fmt.Print(err)
        }

        str := string(b)
        lines := strings.Split(str,"\n")

        for _, line := range lines {
                if line == "" {
                        continue
                }
                chunks := strings.Split(line, " ")
                cityName := chunks[0]
                direcAndCitys := chunks[1:]
                dirCityMap := make(map[string]bool)

                for _, direcAndCity := range direcAndCitys {
                        temps := strings.Split(direcAndCity, "=")
                        if (len(temps) > 1){
                                dirCityMap[temps[1]] = true
                        }
                }
                data[cityName] = dirCityMap
        }
        return data
}

func RandCity(cities []string) string {
        s1 := rand.NewSource(time.Now().UnixNano())
        r1 := rand.New(s1)
        num := r1.Intn(len(cities))
        return cities[num]
}

func RandMove(cities map[string]bool) string {
        s1 := rand.NewSource(time.Now().UnixNano())
        r1 := rand.New(s1)
        num := r1.Intn(len(cities))
        result := ""
        i := 0

        for city, _ := range cities {
                if i == num{
                        result = city
                        break
                } else {
                        i++
                }
        }

        return result
}
