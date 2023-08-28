package main

import (
	"Intranet_penetration/utility"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) { io.WriteString(w, "Hello from a HandleFunc #1!\n") })
	http.HandleFunc("/img", func(w http.ResponseWriter, _ *http.Request) {
		file, err := os.Open("../阿里云盘/download/2.mp4")
		if err != nil {
			log.Println("打开失败" + err.Error())
		}
		defer file.Close()
		io.Copy(w, file)
	})
	http.ListenAndServe(utility.Localhost, nil)

}
