package firestorm

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type request struct {
	FSC        *FSClient
	loadPaths  []string
	mapperFunc MapperFunc
}

type MapperFunc func(map[string]interface{})

func (req *request) SetMapperFunc(mapperFunc MapperFunc) *request {
	req.mapperFunc = mapperFunc
	return req
}

func (req *request) SetLoadPaths(paths ...string) *request {
	req.loadPaths = paths
	return req
}

func (req *request) ToCollection(entity interface{}) *firestore.CollectionRef {
	path := getTypeName(entity)

	// prefix any parents
	for p := req.GetParent(entity); p != nil; p = req.GetParent(p) {
		n := getTypeName(p)
		path = n + "/" + req.GetID(p) + "/" + path
	}

	return req.FSC.Client.Collection(path)
}

func (req *request) GetParent(entity interface{}) interface{} {
	if v, err := getIDValue(req.FSC.ParentKey, entity); err != nil {
		return nil
	} else {
		return v.Interface()
	}
}

func (req *request) GetID(entity interface{}) string {
	if v, err := getIDValue(req.FSC.IDKey, entity); err != nil {
		panic(err)
	} else {
		return v.Interface().(string)
	}
}

func getIDValue(id string, entity interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(entity)
	if cv, ok := entity.(reflect.Value); ok {
		v = cv
	}
	v = reflect.Indirect(v)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		sf := v.Type().Field(i)

		switch f.Kind() {
		case reflect.Struct:
			if sf.Anonymous {
				if sv, err := getIDValue(id, f); err == nil {
					return sv, nil
				}
			}
		}

		// first check if id is statically set
		if strings.ToLower(sf.Name) == id {
			return f, nil
		}
		// otherwise use the tag
		/* not supported yet
		if tag, ok := sf.Tag.Lookup("firestorm"); ok {
			if tag == "id" {
				return f, nil
			}
		}
		*/
	}
	return v, errors.New(fmt.Sprintf("Entity has no id field defined: %v", entity))
}

func (req *request) SetID(entity interface{}, id string) {
	v, err := getIDValue(req.FSC.IDKey, entity)
	if err != nil {
		panic(err)
	}
	v.SetString(id)
}

func (req *request) ToRef(entity interface{}) *firestore.DocumentRef {
	return req.ToCollection(entity).Doc(req.GetID(entity))
}

func (req *request) GetEntity(ctx context.Context, entity interface{}) futureFunc {
	f := req.FSC.getEntities(ctx, req, &[]interface{}{entity})
	return func() error {
		_, err := f()
		return err
	}
}

func (req *request) GetEntities(ctx context.Context, entities interface{}) func() ([]interface{}, error) {
	return req.FSC.getEntities(ctx, req, entities)
}

func (req *request) CreateEntity(ctx context.Context, entity interface{}) futureFunc {
	return req.FSC.createEntity(ctx, req, entity)
}

func (req *request) CreateEntities(ctx context.Context, entities interface{}) futureFunc {
	return req.FSC.createEntities(ctx, req, entities)
}

func (req *request) UpdateEntity(ctx context.Context, entity interface{}) futureFunc {
	return req.FSC.updateEntity(ctx, req, entity)
}

func (req *request) UpdateEntities(ctx context.Context, entities interface{}) futureFunc {
	return req.FSC.updateEntities(ctx, req, entities)
}

func (req *request) DeleteEntities(ctx context.Context, entities interface{}) futureFunc {
	return req.FSC.deleteEntities(ctx, req, entities)
}

func (req *request) DeleteEntity(ctx context.Context, entity interface{}) futureFunc {
	return req.FSC.deleteEntity(ctx, req, entity)
}

func (req *request) QueryEntities(ctx context.Context, query firestore.Query, toSlicePtr interface{}) futureFunc {
	return req.FSC.queryEntities(ctx, req, query, toSlicePtr)
}
