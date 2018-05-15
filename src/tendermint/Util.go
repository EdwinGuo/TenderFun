package main

import (
   "fmt"
   "io/ioutil"
   "strings"
)

// parseCity will parse a file and return a data structure  will serve the purpose of 
// give a city and a directory, get the next city or empty string if not exist
func parseCity(file string) map[string]map[string]string{
	data := make(map[string]map[string]string)
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
		dirCityMap := make(map[string]string)

		for _, direcAndCity := range direcAndCitys {
			temps := strings.Split(direcAndCity, "=")
			if (len(temps) > 1){
				dirCityMap[temps[0]] = temps[1]
			}
		}
		data[cityName] = dirCityMap
        }
        return data
}

func randCity(cities []string) string {
	//	TODO, update this so that return a city in a random fashion
	return cities[0]
}

func main() {
	lookup := parseCity("/Users/edwinguo/edwin/TenderFun/src/test/testcase1.csv")
	fmt.Println(len(lookup))
	for a, b := range lookup {
		fmt.Println("ppp: ", a, " ,", b)
	}
	
}

