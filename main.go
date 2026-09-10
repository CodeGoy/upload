package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
)

var (
	//go:embed html/root.html
	rootHtml string
	version  = "0.0.1a"
)

type Server struct {
	port     string
	tlsPort  string
	path     string
	certFile string
	keyFile  string
	endpoint string
}
type Template struct {
	Version  string
	Endpoint string
}

func (s *Server) applyTemplate(htmlString string, w http.ResponseWriter) error {
	t := template.New("t")
	if _, err := t.Parse(htmlString); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to parse rootHtml->t.Parse(rootHtml)::%v\n", err)
	}
	if err := t.Execute(w, Template{Version: version, Endpoint: s.endpoint}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("failed to execute template->t.Execute(w, &t)::%v\n", err)
	}
	return nil
}

func (s *Server) start() {
	http.HandleFunc("/"+s.endpoint, func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		if err := s.applyTemplate(rootHtml, w); err != nil {
			log.Fatalf("%v", err)
		}
	})
	http.HandleFunc("/"+s.endpoint+"upload", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			log.Panicf("r.FormFile(): %v\n", err)
		}
		defer func(MultipartForm *multipart.Form) {
			fmt.Println("Removing MultipartForm cache")
			err := MultipartForm.RemoveAll()
			if err != nil {
				log.Printf("failed to remove MultiPartForm cache: %v\n", err)
			}
		}(r.MultipartForm)
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
	if s.certFile != "" && s.keyFile != "" {
		go func() {
			log.Println("Starting HTTP redirect server on port", s.port)
			if err := http.ListenAndServe(":"+s.port, http.HandlerFunc(s.httpsRedirect)); err != nil {
				log.Printf("HTTP server failed: %v", err)
			}
		}()
		if err := http.ListenAndServeTLS(":"+s.tlsPort, s.certFile, s.keyFile, nil); err != nil {
			log.Printf("HTTP server failed: %v", err)
		}
	} else {
		if err := http.ListenAndServe(":"+s.port, nil); err != nil {
			log.Printf("listenAndServe():: %v\n", err)
		}
	}
}

func (s *Server) httpsRedirect(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}
	target := fmt.Sprintf("https://%s:%s%s", host, s.tlsPort, r.URL.Path)
	fmt.Printf("Redirecting to: %s\n", target)
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
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
	c, err := loadConfig("/etc/upload.conf")
	if err != nil {
		log.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config: %+v\n", c)
	s := Server{}
	flag.StringVar(&s.endpoint, "endpoint", c.Endpoint, "HTTP endpoint ")
	flag.StringVar(&s.port, "port", c.Port, "port to listen on")
	flag.StringVar(&s.tlsPort, "tls-port", c.TlsPort, "TLS port to listen on")
	flag.StringVar(&s.path, "path", c.Path, "override file location")
	flag.StringVar(&s.certFile, "cert", c.Cert, "cert file")
	flag.StringVar(&s.keyFile, "key", c.Key, "key file")
	flag.Parse()
	fmt.Printf("Server: %+v\n", s)
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
	if s.path[len(s.path)-1] != '/' {
		s.path += "/"
	}
	if err := makeDir(s.path); err != nil {
		log.Fatalf("%v\n", err)
	}
	fmt.Printf("Path: %s\n", s.path)
	fmt.Printf("http://%s:%s/%s\n", ip, s.port, s.endpoint)
	fmt.Printf("https://%s:%s/%s\n", ip, s.tlsPort, s.endpoint)
	s.start()
}
