package kuberesolver

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/naming"
)

// kubeResolver resolves service names using Kubernetes endpoints.
type kubeResolver struct {
	k8sClient *k8sClient
	namespace string
	target    targetInfo
	watcher   *Watcher
}

// NewResolver returns a new Kubernetes resolver.
func newResolver(client *k8sClient, namespace string, targetInfo targetInfo) *kubeResolver {
	if namespace == "" {
		namespace = "default"
	}
	return &kubeResolver{client, namespace, targetInfo, nil}
}

// Resolve creates a Kubernetes watcher for the named target.
func (r *kubeResolver) Resolve(target string) (naming.Watcher, error) {
	resultChan := make(chan watchResult)
	stopCh := make(chan struct{})

	go Until(func() {
		err := r.watch(target, stopCh, resultChan)
		if err != nil {
			grpclog.Printf("kuberesolver: watching ended with error='%v', will reconnect againg", err)
		}
	}, time.Second, stopCh)

	r.watcher = &Watcher{
		target:    r.target,
		endpoints: make(map[string]interface{}),
		stopCh:    stopCh,
		result:    resultChan,
	}
	return r.watcher, nil
}

func (r *kubeResolver) watch(target string, stopCh <-chan struct{}, resultCh chan<- watchResult) error {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/watch/namespaces/%s/endpoints/%s",
		r.k8sClient.host, r.namespace, target))
	if err != nil {
		return err
	}
	req, err := r.k8sClient.getRequest(u.String())
	if err != nil {
		return err
	}
	resp, err := r.k8sClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return fmt.Errorf("invalid response code %d", resp.StatusCode)
	}
	sw := newStreamWatcher(resp.Body)
	for {
		select {
		case <-stopCh:
			return nil
		case up, more := <-sw.ResultChan():
			if more {
				resultCh <- watchResult{err: nil, ep: &up}
			} else {
				return nil
			}
		}
	}
}
