// +build !go1.5

package main

import "io/ioutil"
import "fmt"
import "strings"
import tfw "github.com/tsnow/typing-for-war/cmd/tfw"

type nope int

const YEP nope = 0
const NOPE nope = 1

func Fuzz(data []byte) int {
	nope, obj, attempt := split(data)
	if nope != 0 {
		return int(nope)
	}
	tfw.GoodBadLeft(obj, attempt)
	return 0
}
func dlen(data []byte) (nope, int) {
	if len(data) == 0 {
		return NOPE, 0
	}
	if len(data) > int(data[0]) {
		return YEP, int(data[0])
	}
	return YEP, len(data) - 1
}
func split(data []byte) (nope, string, string) {
	nope, dlen := dlen(data)
	if nope != 0 {
		return nope, "", ""
	}
	return nope, string(data[:dlen]), string(data[dlen:])
}

func printCorp(filename string) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	nope, obj, attempt := split(buf)
	if nope != 0 {
		return
	}
	testname := strings.Replace(filename, "corpus/", "", -1)
	testname = strings.Replace(testname, "-", "_", -1)
	fmt.Printf("func TestFuzz%s(t *testing.T){\n    gbl := tfw.GoodBadLeft(%q, %q)\n    t.Error(gbl) \n}", testname, obj, attempt)

}
func main() {
	files, err := ioutil.ReadDir("corpus")
	if err != nil {
		panic(err)
	}
	fmt.Print("package tfw\nimport tfw \"github.com/tsnow/typing-for-war/cmd/tfw\"\nimport \"testing\"\n")
	for _, fInfo := range files {
		printCorp("corpus/" + fInfo.Name())
		fmt.Print("\n\n")
	}
}
