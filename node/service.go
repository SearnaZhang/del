package node
import (
	"reflect"
	"github.com/DEL-ORG/del/accounts"
	"github.com/DEL-ORG/del/ethdb"
	"github.com/DEL-ORG/del/event"
	"github.com/DEL-ORG/del/p2p"
	"github.com/DEL-ORG/del/rpc"
)
type ServiceContext struct {
	config         *Config
	services       map[reflect.Type]Service 
	EventMux       *event.TypeMux           
	AccountManager *accounts.Manager        
}
func (ctx *ServiceContext) OpenDatabase(name string, cache int, handles int) (ethdb.Database, error) {
	if ctx.config.DataDir == "" {
		return ethdb.NewMemDatabase()
	}
	db, err := ethdb.NewLDBDatabase(ctx.config.resolvePath(name), cache, handles)
	if err != nil {
		return nil, err
	}
	return db, nil
}
func (ctx *ServiceContext) ResolvePath(path string) string {
	return ctx.config.resolvePath(path)
}
func (ctx *ServiceContext) Service(service interface{}) error {
	element := reflect.ValueOf(service).Elem()
	if running, ok := ctx.services[element.Type()]; ok {
		element.Set(reflect.ValueOf(running))
		return nil
	}
	return ErrServiceUnknown
}
type ServiceConstructor func(ctx *ServiceContext) (Service, error)
type Service interface {
	Protocols() []p2p.Protocol
	APIs() []rpc.API
	Start(server *p2p.Server) error
	Stop() error
}
