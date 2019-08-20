package asy

import (
	"reflect"
)

type asyncRun struct {
	Handler reflect.Value
	Params  []reflect.Value
}

type Async struct {
	count int
	tasks map[string]asyncRun
}

func NewAsync() Async {
	return New()
}

func New() Async {
	return Async{tasks: make(map[string]asyncRun)}
}

func (a *Async) Add(name string, handler interface{}, params ...interface{}) bool {
	if _, e := a.tasks[name]; e {
		return false
	}

	handlerValue := reflect.ValueOf(handler)
	if handlerValue.Kind() == reflect.Func {

		paramNum := len(params)

		a.tasks[name] = asyncRun{
			Handler: handlerValue,
			Params:  make([]reflect.Value, paramNum),
		}

		if paramNum > 0 {
			for k, v := range params {
				a.tasks[name].Params[k] = reflect.ValueOf(v)
			}
		}

		a.count++
		return true
	}

	return false
}

func (a *Async) Run() (chan map[string][]interface{}, bool) {
	if a.count < 1 {
		return nil, false
	}
	result := make(chan map[string][]interface{})
	chans := make(chan map[string]interface{}, a.count)

	go func(result chan map[string][]interface{}, chans chan map[string]interface{}) {
		rs := make(map[string][]interface{})
		defer func(rs map[string][]interface{}) {
			result <- rs
		}(rs)
		for {
			if a.count < 1 {
				break
			}

			select {
			case res := <-chans:
				a.count--
				rs[res["name"].(string)] = res["result"].([]interface{})
			}
		}
	}(result, chans)

	for k, v := range a.tasks {
		go func(name string, chans chan map[string]interface{}, async asyncRun) {
			result := make([]interface{}, 0)
			defer func(name string, chans chan map[string]interface{}) {
				chans <- map[string]interface{}{"name": name, "result": result}
			}(name, chans)

			values := async.Handler.Call(async.Params)

			if valuesNum := len(values); valuesNum > 0 {
				resultItems := make([]interface{}, valuesNum)
				for k, v := range values {
					resultItems[k] = v.Interface()
				}
				result = resultItems
				return
			}
		}(k, chans, v)
	}

	return result, true
}

func (a *Async) Clean() {
	a.tasks = make(map[string]asyncRun)
}
