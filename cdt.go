package main

import "github.com/jlaffaye/ftp"

type entries struct {
	entry string
}

func main() {
	user := "jdale215@gmail.com"
	pass := ">hh]dR4KK2%:+?n^~u%J"
	remote := "sftp://sftp.bitport.io"
	port := ":2022"

	results, error := connect()
	println(results)

}
func connect() (entries entries, err error) {
	var results entries
	client, err := ftp.Dial("localhost:21")
	if err != nil {
		return entries, err
	}

	if err := client.Login("root", "password"); err != nil {
		return entries, err
	}

	entries.entry = client.List()
	return entries
}
