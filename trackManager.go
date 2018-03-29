package data

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nimezhu/indexed"
)

/*
type DataManager interface {
	AddURI(uri string, key string) error
	Del(string) error
	ServeTo(*mux.Router)
	List() []string
	Get(string) (string, bool)
	Move(key1 string, key2 string) bool
}
*/
/* TrackManager supports bigbed bigwig and hic , tabixImage */
type TrackManager struct {
	id        string
	uriMap    map[string]string
	formatMap map[string]string
	managers  map[string]Manager
}

func NewTrackManager(uri string, dbname string) *TrackManager {
	m := InitTrackManager(dbname)
	uriMap := LoadURI(uri)
	for k, v := range uriMap {
		m.AddURI(v, k)
	}
	return m
}
func InitTrackManager(dbname string) *TrackManager {
	uriMap := make(map[string]string)
	formatMap := make(map[string]string)
	dataMap := make(map[string]Manager)
	m := TrackManager{
		dbname,
		uriMap,
		formatMap,
		dataMap,
	}
	return &m
}
func newManager(prefix string, format string) Manager {
	if format == "bigwig" {
		return InitBigWigManager(prefix + ".bigwig")
	}
	if format == "bigbed" {
		return InitBigBedManager(prefix + ".bigbed")
	}
	if format == "bigbedLarge" {
		return InitBigBedManager(prefix + ".bigbedLarge")
	}
	if format == "hic" {
		return InitHicManager(prefix + ".hic")
	}
	if format == "image" {
		return InitTabixImageManager(prefix + ".image")
	}
	return nil
}
func (m *TrackManager) Add(key string, reader io.ReadSeeker, uri string) error {
	format, _ := indexed.MagicReadSeeker(reader)
	if format == "hic" || format == "bigwig" || format == "bigbed" {
		reader.Seek(0, 0)
		if _, ok := m.managers[format]; !ok {
			m.managers[format] = newManager(m.id, format)
		}
		m.managers[format].Add(key, reader, uri)
		m.formatMap[key] = format
		m.uriMap[key] = uri
	}
	return nil
}
func (m *TrackManager) AddURI(uri string, key string) error {
	format, _ := indexed.Magic(uri)
	m.formatMap[key] = format
	m.uriMap[key] = uri
	//HANDLE binindex in mem
	/*
		if format == "binindex" {
			m.uriMap[key] = strings.Replace(uri, "_format_:binindex:", "", 1)
			return nil
		}
	*/
	if _, ok := m.managers[format]; !ok {
		if _m := newManager(m.id, format); _m != nil {
			m.managers[format] = _m
			m.managers[format].AddURI(uri, key)
		} else {
			return errors.New("format not support yet")
		}
	} else {
		m.managers[format].AddURI(uri, key)
	}
	return nil
}

func (m *TrackManager) Del(k string) error {
	if uri, ok := m.uriMap[k]; ok {
		format, _ := indexed.Magic(uri)
		//TODO binindex
		delete(m.uriMap, k)
		return m.managers[format].Del(k)
	}
	return errors.New("key not found")
}

func (m *TrackManager) ServeTo(router *mux.Router) {
	prefix := "/" + m.id
	router.HandleFunc(prefix+"/ls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		jsonHic, _ := json.Marshal(m.uriMap)
		w.Write(jsonHic)
	})
	router.HandleFunc(prefix+"/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		a := make([]map[string]string, 0)
		for k, _ := range m.uriMap {
			a = append(a, map[string]string{"id": k, "format": m.formatMap[k]})
		}
		jsonHic, _ := json.Marshal(a)
		w.Write(jsonHic)
	})
	router.HandleFunc(prefix+"/{id}/{cmd:.*}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		cmd := params["cmd"]
		id := params["id"]
		format, _ := m.formatMap[id]
		//w.Write([]byte(code))
		//TODO redirect with format
		/*
			if format == "binindex" { //binindex in memory
				uri, _ := m.uriMap[id] //name
				a1 := strings.Replace(r.URL.String(), prefix+"/", "", 1)
				a2 := strings.Replace(a1, id, uri, 1)
				//url := "/" + uri + "/" + cmd
				//fmt.Println(a2)
				//fmt.Println(url)
				url := "/" + a2
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			} else {
		*/
		url := prefix + "." + format + "/" + id + "/" + cmd
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		//}
	})
	for _, v := range m.managers {
		v.ServeTo(router)
	}
}

func (m *TrackManager) List() []string {
	s := make([]string, 0)
	for k, _ := range m.uriMap {
		s = append(s, k)
	}
	return s
}
func (m *TrackManager) Get(key string) (string, bool) {
	v, ok := m.uriMap[key]
	return v, ok
}
func (m *TrackManager) Move(key1 string, key2 string) bool { //TODO
	if v, ok := m.uriMap[key1]; ok {
		m.uriMap[key2] = v
		delete(m.uriMap, key1)
		format, _ := indexed.Magic(v) // TODO Fix this with ReadSeeker.
		m.managers[format].Move(key1, key2)
		return true
	}
	return false
}
