package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

var (
	//go:embed html/upload.html
	uploadHTML string
	version    = "0.0.1a"
)

type Server struct {
	port string
	path string
}
type Template struct {
	Version string
}

func (s *Server) applyTemplate(htmlString string, w http.ResponseWriter) error {
	t := template.New("t")
	if _, err := t.Parse(htmlString); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to parse rootHtml->t.Parse(rootHtml)::%v\n", err)
	}
	if err := t.Execute(w, Template{Version: version}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to execute template->t.Execute(w, &t)::%v\n", err)
	}
	return nil
}

func (s *Server) start() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := s.applyTemplate(uploadHTML, w); err != nil {
			log.Fatalf("%v", err)
		}
	})
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			log.Panicf("r.FormFile(): %v\n", err)
		}
		file, handler, err := r.FormFile("file")
		if err != nil {
			log.Printf("r.FormFile(): %v\n", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Panicf("%v\v", err)
			}
		}()
		f, err := os.OpenFile(fmt.Sprintf("%s/%s", s.path, handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Panicf("%v\n", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("%v\v", err)
			}
		}()
		if _, err := io.Copy(f, file); err != nil {
			log.Printf("%v\n", err)
		}
		response := fmt.Sprintf("Received File: %s", handler.Filename)
		if _, err := w.Write([]byte(response)); err != nil {
			log.Printf("%v\n", err)
		}
	})
	if err := http.ListenAndServe(":"+s.port, nil); err != nil {
		log.Printf("listenAndServe():: %v\n", err)
	}
}

func makeDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("Failed to create dir: %v\n", err)
		}
	}
	return nil
}

func main() {

	s := Server{}
	flag.StringVar(&s.port, "port", "8605", "port to listen on")
	flag.StringVar(&s.path, "path", ".", "override file location")
	flag.Parse()
	ip := func() string {
		adders, err := net.InterfaceAddrs()
		if err != nil {
			log.Panicf("net.InterfaceAddrs:%v\n", err)
		}
		for _, address := range adders {
			if inet, ok := address.(*net.IPNet); ok && !inet.IP.IsLoopback() {
				fmt.Printf("Network Interface: %v %s\n", inet.IP, inet.String())
				if inet.IP.To4() != nil {
					return inet.IP.String()
				}
			}
		}
		return ""
	}()
	fmt.Printf("http://%s:%s/\n", ip, s.port)
	s.start()
}
