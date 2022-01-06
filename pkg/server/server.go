package server

import (
	_ "embed"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

//go:embed dir_template.html
var dirTemplateHtml string

type Server struct {
	BindAddr      string
	Datadir       string
	Favicon       []byte
	IncludeHidden bool
}

func (s *Server) ListenAndServe() error {
	if len(s.Datadir) > 1 && strings.HasSuffix(s.Datadir, "/") {
		// housekeeping on startup, remove trailing slash to allow easy removing later
		// for server root
		s.Datadir = s.Datadir[:len(s.Datadir)-1]
	}
	log.Printf("serving '%s' on addr '%s'", s.Datadir, s.BindAddr)
	return http.ListenAndServe(s.BindAddr, s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Path

	if s.Favicon != nil && p == "/favicon.ico" {
		_, _ = w.Write(s.Favicon)
		s.logRequest(http.StatusOK, p)
		return
	}

	// this is a security hack right now, it should be solved with a chroot
	if strings.Contains(p, "..") {
		s.logRequest(403, p)
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	// we have a json api for directories, and a static file server for files

	fullpath := path.Join(s.Datadir, p)

	info, err := os.Stat(fullpath)

	if err != nil {
		if os.IsNotExist(err) {
			s.logRequest(404, p)
			http.NotFound(w, req)
		} else {
			s.logRequest(500, p)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if info.IsDir() {
		entries, err := s.getDirectoryEntries(fullpath)
		if err != nil {
			s.logRequest(500, p)
			http.Error(w, "500 directory not implemented", http.StatusInternalServerError)
			return
		}

		// do we write json or html?
		if strings.Contains(req.Header.Get("Accept"), "application/json") {
			json, err := json.Marshal(&DirectoryList{Path: p, Contents: entries})
			if err != nil {
				s.logRequest(500, p)
				http.Error(w, "500 json encoding error", http.StatusInternalServerError)
				return
			}
			s.logRequest(200, p)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(json)
		} else {
			s.logRequest(200, p)
			contentBuilder := strings.Builder{}

			// this is html only, and is just a handy wee thing to have
			if p != "/" {
				// add an entry for the parent directory
				_, _ = contentBuilder.WriteString("<li><a href=\"" + path.Dir(p) + "\">↑&nbsp;..</a></li>")
			}
			for _, entry := range entries {
				_, _ = contentBuilder.WriteString(entry.ToHtmlListItem())
				_, _ = contentBuilder.WriteString("\n")
			}
			fullContent := strings.Replace(dirTemplateHtml, "{{content}}", contentBuilder.String(), 1)
			fullContent = strings.Replace(fullContent, "{{path}}", p, 1)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(fullContent))
		}
		return
	}

	// the file exists and is not a directory, serve it!
	s.logRequest(200, p)
	http.ServeFile(w, req, s.Datadir+p)
}

func (s *Server) getDirectoryEntries(p string) ([]DirectoryEntry, error) {
	infos, err := ioutil.ReadDir(p)
	if err != nil {
		return nil, err
	}

	entries := make([]DirectoryEntry, 0, len(infos))

	for _, info := range infos {
		name := info.Name()
		if strings.HasPrefix(name, ".") && !s.IncludeHidden {
			continue
		}
		if info.IsDir() {
			name += "/"
		}
		entries = append(entries, DirectoryEntry{
			Name:  name,
			Size:  uint64(info.Size()),
			Path:  strings.TrimPrefix(path.Join("/", p, name), s.Datadir), // path is absolute, but relative to datadir
			IsDir: info.IsDir(),
		})
	}
	return entries, nil
}

func (s *Server) logRequest(responseCode int, urlPath string) {
	log.Printf("%d %s\n", responseCode, urlPath)
}

type DirectoryList struct {
	Path     string           `json:"path"`
	Contents []DirectoryEntry `json:"contents"`
}

type DirectoryEntry struct {
	Name  string `json:"name"`
	Size  uint64 `json:"size"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

func (d *DirectoryEntry) ToHtmlListItem() string {
	var size string
	if d.IsDir {
		size = "-"
	} else {
		size = fmt.Sprintf("%d", d.Size)
	}
	return fmt.Sprintf("<li><a href=\"%s\">%s</a> (%s)</li>", d.Path, d.Name, size)
}
