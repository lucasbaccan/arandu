package api

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// devProxy encaminha o frontend para o Vite, tentando vários endereços em ordem
// (localhost e, no WSL2, o IP do host Windows onde o Vite pode estar rodando).
func devProxy(targets string) http.Handler {
	upstreams := parseUpstreams(targets)
	if len(upstreams) == 0 {
		log.Printf("api: nenhum endereço de frontend válido")
		return nil
	}
	return &fallbackProxy{upstreams: upstreams}
}

func parseUpstreams(targets string) []*url.URL {
	var list []string
	if strings.TrimSpace(targets) != "" {
		for _, t := range strings.Split(targets, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				list = append(list, t)
			}
		}
	} else {
		list = append(list, "http://localhost:5173")
		if gw := wslWindowsHostIP(); gw != "" {
			list = append(list, "http://"+gw+":5173")
		}
	}

	var upstreams []*url.URL
	for _, t := range list {
		u, err := url.Parse(t)
		if err != nil || u.Host == "" {
			log.Printf("api: VITE_DEV_URL inválido: %q", t)
			continue
		}
		upstreams = append(upstreams, u)
	}
	return upstreams
}

// wslWindowsHostIP retorna o IP do host Windows quando rodando dentro do WSL2
// (lê o gateway padrão de /proc/net/route), ou "" caso contrário.
func wslWindowsHostIP() string {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[1] != "00000000" {
			continue
		}
		gw, err := strconv.ParseUint(fields[2], 16, 32)
		if err != nil || gw == 0 {
			continue
		}
		return fmt.Sprintf("%d.%d.%d.%d", byte(gw), byte(gw>>8), byte(gw>>16), byte(gw>>24))
	}
	return ""
}

type fallbackProxy struct {
	mu         sync.Mutex
	upstreams  []*url.URL
	currentIdx int
	lastProbe  time.Time
}

func (p *fallbackProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idx := p.aliveIndex()
	if idx < 0 {
		writeError(w, http.StatusBadGateway,
			"Frontend (Vite) não está rodando. Execute 'make frontend-dev' (ou 'make dev' para subir os dois juntos).")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(p.upstreams[idx])
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		p.mu.Lock()
		if p.currentIdx == idx {
			p.currentIdx = -1
			p.lastProbe = time.Time{}
		}
		p.mu.Unlock()
		writeError(w, http.StatusBadGateway,
			"Frontend (Vite) não está respondendo. Execute 'make frontend-dev' (ou 'make dev' para subir os dois juntos).")
	}
	proxy.ServeHTTP(w, r)
}

// aliveIndex devolve o índice do primeiro upstream acessível, testando a cada
// sonda (com cache de 1 segundo).
func (p *fallbackProxy) aliveIndex() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentIdx >= 0 && time.Since(p.lastProbe) < time.Second {
		return p.currentIdx
	}
	for i, u := range p.upstreams {
		if reachable(u) {
			p.currentIdx = i
			p.lastProbe = time.Now()
			return i
		}
	}
	p.currentIdx = -1
	p.lastProbe = time.Now()
	return -1
}

func reachable(u *url.URL) bool {
	conn, err := net.DialTimeout("tcp", u.Host, 300*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
