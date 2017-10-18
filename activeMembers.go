package main

func newActiveMembers() *activeMembers {
	active := activeMembers{
		data: make(map[string]string),
	}
	return &active
}

type activeMembers struct {
	data map[string]string
}

func (a *activeMembers) add(key, val string) {
	a.data[key] = val
}

func (a *activeMembers) delete(key string) {
	delete(a.data, key)
}

func (a *activeMembers) get(key string) string {
	if val, ok := a.data[key]; ok {
		return val
	}
	return ""
}

func (a *activeMembers) keyIn(key string) bool {
	if _, ok := a.data[key]; ok {
		return true
	}
	return false
}
