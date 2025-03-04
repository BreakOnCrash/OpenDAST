package geoip

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/BreakOnCrash/opendast/pkg/download"
	"github.com/oschwald/geoip2-golang"
)

var (
	ErrVaildIP = errors.New("query should be valid IP")
)

const (
	// TODO you can replace it
	ASNDB  = "https://raw.githubusercontent.com/ac0d3r/finder/data/geoip/ASN.mmdb"
	CityDB = "https://raw.githubusercontent.com/ac0d3r/finder/data/geoip/City.mmdb"

	LocalASN  = "geo/ASN.mmdb"
	LocalCity = "geo/City.mmdb"
)

var dbPaths = map[string]string{
	LocalASN:  ASNDB,
	LocalCity: CityDB,
}

type GeoIP struct {
	mux  sync.Mutex
	asn  *geoip2.Reader
	city *geoip2.Reader
}

// new geoip from database file
func NewGeoIP() (*GeoIP, error) {
	g := &GeoIP{}

	if err := g.fetchDB(false); err != nil {
		return nil, err
	}
	if err := g.compileDB(); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *GeoIP) Update() (err error) {
	if err := g.fetchDB(true); err != nil {
		return err
	}
	if err := g.compileDB(); err != nil {
		return err
	}
	return nil
}

func (g *GeoIP) fetchDB(force bool) error {
	for local, url := range dbPaths {
		if !download.FileExist(local) || force {
			if err := download.Download(local, url); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *GeoIP) compileDB() error {
	g.mux.Lock()
	defer g.mux.Unlock()

	asn, err := geoip2.Open(LocalASN)
	if err != nil {
		return err
	}
	city, err := geoip2.Open(LocalCity)
	if err != nil {
		return err
	}

	g.asn = asn
	g.city = city
	return nil
}

func (g *GeoIP) Find(ip net.IP, lang string) (result Result, err error) {
	g.mux.Lock()
	defer g.mux.Unlock()

	if ip == nil {
		return result, ErrVaildIP
	}

	if g.asn == nil || g.city == nil {
		return result, nil
	}

	record, err := g.city.City(ip)
	if err != nil {
		return
	}

	if lang == "" {
		lang = "en"
	}

	result = Result{
		Country: record.Country.Names[lang],
		Area:    record.City.Names[lang],
	}

	if record, err := g.asn.ASN(ip); err != nil {
		return result, err
	} else {
		result.ASN = record.AutonomousSystemNumber
		result.Org = record.AutonomousSystemOrganization
	}
	return
}

type Result struct {
	Country string `json:"country"`
	Area    string `json:"area"`
	ASN     uint   `json:"asn"`
	Org     string `json:"org"`
}

func (r Result) ASNStr() string {
	return fmt.Sprintf("AS%d", r.ASN)
}
