package core

type IUISource interface {
	FileName() string
	IsLoaded() bool
	Load(callback func())
}
