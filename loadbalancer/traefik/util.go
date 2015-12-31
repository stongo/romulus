package traefik

import (
	"fmt"
	"path"
	"strconv"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/emilevauge/traefik/types"
	"github.com/timelinelabs/romulus/kubernetes"
	"github.com/timelinelabs/romulus/loadbalancer"
)

func buildRoute(rt *kubernetes.Route) map[string]types.Route {
	if rt.Empty() {
		return defaultTraefikRoute
	}

	routes := map[string]types.Route{}
	for rule, val := range rt.GetParts() {
		routes[rule] = types.Route{Rule: rule, Value: val}
	}
	return routes
}

func getBackend(s Store, prefix, id string) (*backend, error) {
	kp := path.Join(prefix, "backends", id)
	logger.Debugf("[%v] Lookup Backend %q", id, kp)

	b := new(types.Backend)
	lb, er := s.GetString(path.Join(kp, "loadbalancer", "method"))
	if er != nil {
		return nil, er
	}
	cb, er := s.GetString(path.Join(kp, "circuitbreaker", "expression"))
	if er != nil {
		return nil, er
	}

	if lb != "" {
		b.LoadBalancer = &types.LoadBalancer{Method: lb}
	}
	if cb != "" {
		b.CircuitBreaker = &types.CircuitBreaker{Expression: cb}
	}

	servers, er := s.Keys(path.Join(kp, "servers"))
	if er != nil {
		logger.Debugf("[%v] Key read error: %v", er)
		return &backend{Backend: *b, id: id}, nil
	}
	b.Servers = make(map[string]types.Server)
	for _, server := range servers {
		srvID := path.Base(server)
		if srvID == "." || srvID == "/" {
			continue
		}
		u, er := s.GetString(path.Join(server, "url"))
		if er != nil {
			logger.Debugf("[%v] Key read error: %v", id, er)
			continue
		}
		w, er := s.GetString(path.Join(server, "weight"))
		if er != nil {
			logger.Debugf("[%v] Key read error: %v", id, er)
			continue
		}
		i, _ := strconv.Atoi(w)
		b.Servers[srvID] = types.Server{URL: u, Weight: i}
	}
	return &backend{Backend: *b, id: id}, nil
}

func getFrontend(s Store, prefix, id string) (*frontend, error) {
	kp := path.Join(prefix, "frontends", id)
	logger.Debugf("[%v] Lookup Frontend %q", id, kp)

	f := new(types.Frontend)
	bnd, er := s.GetString(path.Join(kp, "backend"))
	if er != nil {
		return nil, fmt.Errorf("Key read error: %v", er)
	}
	pas, er := s.GetString(path.Join(kp, "passHostHeader"))
	if er != nil {
		return nil, fmt.Errorf("Key read error: %v", er)
	}
	val, _ := strconv.ParseBool(pas)
	f.Backend = bnd
	f.PassHostHeader = val

	routes, er := s.Keys(path.Join(kp, "routes"))
	if er != nil {
		logger.Debugf("[%v] Key read error: %v", id, er)
		return &frontend{Frontend: *f, id: id}, nil
	}

	f.Routes = make(map[string]types.Route)
	for _, route := range routes {
		rtID := path.Base(route)
		if id == "." || id == "/" {
			continue
		}
		r, er := s.GetString(path.Join(route, "rule"))
		if er != nil {
			logger.Debugf("[%v] Key read error: %v", id, er)
			continue
		}
		v, er := s.GetString(path.Join(route, "value"))
		if er != nil {
			logger.Debugf("[%v] Key read error: %v", id, er)
			continue
		}
		f.Routes[rtID] = types.Route{Rule: r, Value: v}
	}
	return &frontend{Frontend: *f, id: id}, nil
}

func getServers(s Store, prefix, id string) (list []loadbalancer.Server) {
	kp := path.Join(prefix, "backends", id)
	logger.Debugf("[%v] Lookup Servers for Backend %q", id, kp)

	servers, er := s.Keys(path.Join(kp, "servers"))
	if er != nil {
		logger.Warnf("[%v] Key read error: %v", id, er)
		return list
	}
	for _, srv := range servers {
		srvID := path.Base(srv)
		if srvID == "." || srvID == "/" {
			continue
		}
		u, er := s.GetString(path.Join(srv, "url"))
		if er != nil {
			logger.Debugf("[%v] Key read error: %v", id, er)
			continue
		}
		w, er := s.GetString(path.Join(srv, "weight"))
		if er != nil {
			logger.Debugf("[%v] Key read error: %v", id, er)
			continue
		}
		i, _ := strconv.Atoi(w)
		sr := &server{
			id:     srvID,
			Server: types.Server{URL: u, Weight: i},
		}
		list = append(list, sr)
	}
	return list
}
