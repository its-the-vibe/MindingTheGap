// Package server implements the dashboard HTTP server.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/its-the-vibe/MindingTheGap/internal/config"
	"github.com/its-the-vibe/MindingTheGap/internal/tfl"
)

// Source abstracts the TfL client for testing.
type Source interface {
	LineStatuses(ctx context.Context, ids []string) ([]tfl.LineStatus, error)
	Arrivals(ctx context.Context, line, stationID string) ([]tfl.Arrival, error)
}

type Server struct {
	cfg *config.Config
	src Source
}

func New(cfg *config.Config, src Source) *Server {
	return &Server{cfg: cfg, src: src}
}

type StatusResponse struct {
	Lines               []tfl.LineStatus                    `json:"lines"`
	Arrivals            map[string][]tfl.Arrival            `json:"arrivals"`
	ArrivalsByDirection map[string]map[string][]tfl.Arrival `json:"arrivalsByDirection"`
	StationOrder        []string                            `json:"stationOrder"`
	Errors              []string                            `json:"errors,omitempty"`
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.index)
	mux.HandleFunc("GET /api/v1/status", s.status)
	return mux
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	resp := StatusResponse{
		Arrivals:            map[string][]tfl.Arrival{},
		ArrivalsByDirection: map[string]map[string][]tfl.Arrival{},
		StationOrder:        append([]string{}, s.cfg.TFL.MonitoredStations...),
	}
	ctx := r.Context()
	lines, err := s.src.LineStatuses(ctx, s.cfg.TFL.MonitoredLines)
	if err != nil {
		resp.Errors = append(resp.Errors, "line status unavailable")
		log.Printf("line status: %v", err)
	}
	resp.Lines = orderLineStatuses(lines, s.cfg.TFL.MonitoredLines)
	for _, entry := range s.cfg.TFL.MonitoredStations {
		line, stationID, ok := strings.Cut(entry, "/")
		if !ok || line == "" || stationID == "" {
			resp.Errors = append(resp.Errors, "invalid monitored station entry "+entry)
			continue
		}
		arr, err := s.src.Arrivals(ctx, line, stationID)
		if err != nil {
			resp.Errors = append(resp.Errors, "arrivals unavailable for "+entry)
			log.Printf("arrivals %s: %v", entry, err)
			continue
		}
		sort.Slice(arr, func(i, j int) bool { return arr[i].TimeToStation < arr[j].TimeToStation })
		resp.Arrivals[entry] = arr
		directions := map[string][]tfl.Arrival{}
		for _, arrival := range arr {
			direction := strings.TrimSpace(arrival.Direction)
			if direction == "" {
				direction = "Unknown direction"
			}
			directions[direction] = append(directions[direction], arrival)
		}
		resp.ArrivalsByDirection[entry] = directions
	}
	w.Header().Set("Content-Type", "application/json")
	if len(resp.Errors) > 0 && len(resp.Lines) == 0 && len(resp.Arrivals) == 0 {
		w.WriteHeader(http.StatusBadGateway)
	}
	json.NewEncoder(w).Encode(resp)
}

func orderLineStatuses(lines []tfl.LineStatus, configured []string) []tfl.LineStatus {
	order := make(map[string]int, len(configured))
	for i, id := range configured {
		if _, exists := order[id]; !exists {
			order[id] = i
		}
	}
	sort.SliceStable(lines, func(i, j int) bool {
		iOrder, iConfigured := order[lines[i].ID]
		jOrder, jConfigured := order[lines[j].ID]
		if iConfigured != jConfigured {
			return iConfigured
		}
		return iConfigured && iOrder < jOrder
	})
	return lines
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := strings.Replace(indexHTML, "__REFRESH_INTERVAL_MS__", fmt.Sprint(s.cfg.Dashboard.RefreshIntervalSeconds*1000), 1)
	w.Write([]byte(html))
}

// Run serves until ctx is cancelled, then shuts down gracefully.
func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.cfg.Server.Port),
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}
	sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(sctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

const indexHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>London Tube Dashboard</title>
<style>
body{font-family:system-ui,sans-serif;margin:0;padding:1rem;background:#f4f4f4;color:#111}
h1{font-size:1.4rem}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(240px,1fr));gap:.75rem}
.card{background:#fff;border-radius:6px;padding:.75rem;box-shadow:0 1px 3px #0003}
.good{border-left:6px solid #0a0}.minor{border-left:6px solid #f0ad00}.bad{border-left:6px solid #d00}
.line-status{display:block;width:100%;font:inherit;color:inherit;text-align:left;background:none;border:0;padding:0;cursor:pointer}
.line-reason{white-space:pre-line}
h3{text-transform:capitalize}
table{width:100%;border-collapse:collapse}td{padding:.2rem .4rem;border-bottom:1px solid #eee}
</style></head><body>
<h1>London Tube Dashboard</h1>
<h2>Line status</h2><div id="lines" class="grid"></div>
<h2>Arrivals</h2><div id="arrivals"></div>
<p id="err"></p>
<script>
function el(t,c,x){const e=document.createElement(t);if(c)e.className=c;if(x!==undefined)e.textContent=x;return e}
async function load(){
  try{
    const d=await (await fetch('/api/v1/status')).json();
    const L=document.getElementById('lines');L.replaceChildren();
    (d.lines||[]).forEach(l=>{
      const statuses=l.lineStatuses||[];
      const s=statuses.map(x=>x.statusSeverityDescription).join(', ');
      const c=el('div','card '+(statuses.some(x=>x.statusSeverityDescription==='Minor Delays')?'minor':s==='Good Service'?'good':'bad'));
      const reasons=statuses.map(x=>x.reason).filter(Boolean).join('\n');
      const status=el(reasons?'button':'div','line-status',reasons?undefined:s);
      if(reasons){
        status.type='button';status.setAttribute('aria-expanded','false');
        status.append(el('strong',null,l.name),document.createElement('br'),document.createTextNode(s));
        const reason=el('div','line-reason',reasons);reason.hidden=true;
        status.addEventListener('click',()=>{reason.hidden=!reason.hidden;status.setAttribute('aria-expanded',String(!reason.hidden))});
        c.append(status,reason);
      }else c.append(el('strong',null,l.name),status);
      L.append(c)});
    const A=document.getElementById('arrivals');A.replaceChildren();
    (d.stationOrder||Object.keys(d.arrivals||{})).forEach(id=>{
      const arr=(d.arrivals||{})[id];if(!arr)return;
      const c=el('div','card');
      c.append(el('strong',null,(arr[0]&&arr[0].stationName)||id));
      Object.entries((d.arrivalsByDirection||{})[id]||{}).forEach(([direction,group])=>{
        const lineName=(group[0]&&group[0].lineName)||id.split('/')[0];
        c.append(el('h3',null,lineName+' - '+direction));const t=el('table');
        group.slice(0,10).forEach(a=>{const r=t.insertRow();
          r.insertCell().textContent=a.towards||'';
          r.insertCell().textContent=Math.round(a.timeToStation/60)+' min'});
        c.append(t)});
      A.append(c)});
    document.getElementById('err').textContent=(d.errors||[]).join('; ');
  }catch(e){document.getElementById('err').textContent='Failed to load data'}
}
load();setInterval(load,__REFRESH_INTERVAL_MS__);
</script></body></html>`
